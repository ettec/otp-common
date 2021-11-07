// Types and functions to be used to create execution trading strategies.  To understand how to use the code in this
// package see the examples strategies at https://github.com/ettec/open-trading-platform/tree/master/go/execution-venues/smart-router
// and https://github.com/ettec/open-trading-platform/tree/master/go/execution-venues/vwap-strategy.
package strategy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ettec/otp-common/api/executionvenue"
	"github.com/ettec/otp-common/model"
	"github.com/ettec/otp-common/ordermanagement"
	"github.com/golang/protobuf/proto"
	"log"
	"os"
)

// See the examples referenced in the package doc to see how to make use of this to write a new strategy type.
type Strategy struct {
	// This channel must be checked and handled as part of the select statement in the strategies event loop
	CancelChan           chan string

	// This channel must be checked and handled as part of the select statement in the strategies event loop
	ChildOrderUpdateChan <-chan *model.Order

	ExecVenueId      string
	ParentOrder      *ordermanagement.ParentOrder
	ErrLog           *log.Logger
	Log              *log.Logger
	lastStoredOrder  []byte
	store            func(*model.Order) error
	orderRouter      executionvenue.ExecutionVenueClient
	childOrderStream ordermanagement.ChildOrderStream

	doneChan chan<- string
}

func NewStrategyFromCreateParams(parentOrderId string, params *executionvenue.CreateAndRouteOrderParams, execVenueId string,
	store func(*model.Order) error, orderRouter executionvenue.ExecutionVenueClient,
	childOrderStream ordermanagement.ChildOrderStream, doneChan chan<- string) (*Strategy, error) {

	initialState := model.NewOrder(parentOrderId, params.OrderSide, params.Quantity, params.Price, params.ListingId,
		params.OriginatorId, params.OriginatorRef, params.RootOriginatorId, params.RootOriginatorRef, params.Destination)

	err := initialState.SetTargetStatus(model.OrderStatus_LIVE)

	initialState.ExecParametersJson = params.ExecParametersJson

	if err != nil {
		return nil, err
	}

	om := NewStrategyFromParentOrder(initialState, store, execVenueId, orderRouter, childOrderStream, doneChan)
	return om, nil
}

func NewStrategyFromParentOrder(initialState *model.Order, store func(*model.Order) error, execVenueId string, orderRouter executionvenue.ExecutionVenueClient, childOrderStream ordermanagement.ChildOrderStream, doneChan chan<- string) *Strategy {
	po := ordermanagement.NewParentOrder(*initialState)
	return &Strategy{
		lastStoredOrder:      nil,
		CancelChan:           make(chan string, 10),
		store:                store,
		ParentOrder:          po,
		ExecVenueId:          execVenueId,
		orderRouter:          orderRouter,
		childOrderStream:     childOrderStream,
		ChildOrderUpdateChan: childOrderStream.GetStream(),
		doneChan:             doneChan,
		ErrLog:               log.New(os.Stderr, "order:"+po.Id+" ", log.Flags()),
		Log:                  log.New(log.Writer(), "order:"+po.Id+" ", log.Flags()),
	}
}


// SendChildOrder Sends a child order to the given destination and ensures the parent order cannot become over exposed.
func (om *Strategy) SendChildOrder(side model.Side, quantity *model.Decimal64, price *model.Decimal64, listingId int32,
	destination string, execParametersJson string) error {

	if quantity.GreaterThan(om.ParentOrder.GetAvailableQty()) {
		return fmt.Errorf("cannot send child order for %v as it exceeds the available quantity on the parent order: %v", quantity,
			om.ParentOrder.GetAvailableQty())
	}

	params := &executionvenue.CreateAndRouteOrderParams{
		OrderSide:         side,
		Quantity:          quantity,
		Price:             price,
		ListingId:         listingId,
		Destination:       destination,
		OriginatorId:      om.ExecVenueId,
		OriginatorRef:     om.getStrategyOrderId(),
		RootOriginatorId:  om.ParentOrder.RootOriginatorId,
		RootOriginatorRef: om.ParentOrder.RootOriginatorRef,
		ExecParametersJson: execParametersJson,
	}

	id, err := om.orderRouter.CreateAndRouteOrder(context.Background(), params)

	if err != nil {
		return fmt.Errorf("failed to submit child order:%w", err)
	}

	pendingOrder := model.NewOrder(id.OrderId, params.OrderSide, params.Quantity, params.Price, params.ListingId,
		om.ExecVenueId, om.getStrategyOrderId(), om.ParentOrder.RootOriginatorId, om.ParentOrder.RootOriginatorRef, destination)

	// Orders start at version 0, this is a placeholder for the pending order until the first child order update is received
	pendingOrder.Version = -1

	_, err = om.ParentOrder.OnChildOrderUpdate(pendingOrder)
	if err != nil {
		return err
	}

	return nil
}


// CheckIfDone must be called in the strategies event handling loop, see example strategies as per package documentation.
func (om *Strategy) CheckIfDone() (done bool, err error) {
	done = false
	err = om.persistParentOrderChanges()
	if err != nil {
		return false, fmt.Errorf("failed to persist parent order changes:%w", err)
	}

	if om.ParentOrder.IsTerminalState() {
		done = true
		om.childOrderStream.Close()
		om.doneChan <- om.ParentOrder.GetId()
	}
	return done, nil
}

// OnChildOrderUpdate must be called in response to receipt of a child order update in the strategy's event processing loop as per example strategies
func (om *Strategy) OnChildOrderUpdate(ok bool, co *model.Order) error {
	if ok {
		_, err := om.ParentOrder.OnChildOrderUpdate(co)
		return err
	} else {
		msg := "child order update chan unexpectedly closed, cancelling order"
		om.ErrLog.Printf(msg)
		om.CancelChan <- msg
	}

	return nil
}

// CancelChildOrdersAndStrategyOrder should only be called from with the strategy's event processing loop as per the example strategies
func (om *Strategy) CancelChildOrdersAndStrategyOrder() error {
	if !om.ParentOrder.IsTerminalState() {
		om.Log.Print("cancelling order")
		err := om.ParentOrder.SetTargetStatus(model.OrderStatus_CANCELLED)

		if err != nil {
			return fmt.Errorf("failed to Cancel order:%w", err)
		}

		pendingChildOrderCancels := false
		for _, co := range om.ParentOrder.ChildOrders {
			if !co.IsTerminalState() {
				pendingChildOrderCancels = true
				_, err := om.orderRouter.CancelOrder(context.Background(), &executionvenue.CancelOrderParams{
					OrderId: co.Id,
					ListingId: co.ListingId,
					OwnerId: co.OwnerId,
				})

				if err != nil {
					return fmt.Errorf("failed to Cancel child order:%w", err)
				}

			}

		}

		if !pendingChildOrderCancels {
			err := om.ParentOrder.SetStatus(model.OrderStatus_CANCELLED)
			if err != nil {
				return fmt.Errorf("failed to set status of parent order: %w", err)
			}

		}

	}

	return nil
}


func (om *Strategy) persistParentOrderChanges() error {

	orderAsBytes, err := proto.Marshal(&om.ParentOrder.Order)

	if !bytes.Equal(om.lastStoredOrder, orderAsBytes) {

		if om.lastStoredOrder != nil {
			om.ParentOrder.Version = om.ParentOrder.Version + 1
		}

		toStore, err := proto.Marshal(&om.ParentOrder.Order)
		if err != nil {
			return err
		}

		om.lastStoredOrder = toStore

		orderCopy := &model.Order{}
		err = proto.Unmarshal(toStore, orderCopy)
		if err != nil {
			return err
		}

		err = om.store(orderCopy)
		if err != nil {
			return err
		}

	}

	return err
}

func (om *Strategy) getStrategyOrderId() string {
	return om.ParentOrder.GetId()
}

