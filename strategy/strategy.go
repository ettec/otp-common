// Package strategy contains types and functions to be used to create execution trading strategies.  To understand how to use the code in this
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
	"log/slog"
)

// Strategy is a base type to be used to build trading strategies.  For examples see
// https://github.com/ettec/open-trading-platform/tree/master/go/execution-venues/smart-router
// and https://github.com/ettec/open-trading-platform/tree/master/go/execution-venues/vwap-strategy.
type Strategy struct {
	// This channel must be checked and handled as part of the select statement in the strategies event loop
	CancelChan chan string

	// This channel must be checked and handled as part of the select statement in the strategies event loop
	ChildOrderUpdateChan <-chan *model.Order

	ExecVenueId      string
	ParentOrder      *ordermanagement.ParentOrder
	lastStoredOrder  []byte
	store            func(context.Context, *model.Order) error
	orderRouter      executionvenue.ExecutionVenueClient
	childOrderStream ordermanagement.OrderStream

	Log *slog.Logger

	doneChan chan<- string
}

func NewStrategyFromCreateParams(parentOrderId string, params *executionvenue.CreateAndRouteOrderParams, execVenueId string,
	store func(context.Context, *model.Order) error, orderRouter executionvenue.ExecutionVenueClient,
	childOrderStream ordermanagement.OrderStream, doneChan chan<- string) (*Strategy, error) {

	initialState := model.NewOrder(parentOrderId, params.OrderSide, params.Quantity, params.Price, params.ListingId,
		params.OriginatorId, params.OriginatorRef, params.RootOriginatorId, params.RootOriginatorRef, params.Destination)

	if err := initialState.SetTargetStatus(model.OrderStatus_LIVE); err != nil {
		return nil, fmt.Errorf("failed to set initial order state to live: %w", err)
	}

	initialState.ExecParametersJson = params.ExecParametersJson

	return NewStrategyFromParentOrder(initialState, store, execVenueId, orderRouter, childOrderStream, doneChan), nil
}

func NewStrategyFromParentOrder(initialState *model.Order, store func(context.Context, *model.Order) error, execVenueId string, orderRouter executionvenue.ExecutionVenueClient, childOrderStream ordermanagement.OrderStream, doneChan chan<- string) *Strategy {
	po := ordermanagement.NewParentOrder(*initialState)

	return &Strategy{
		lastStoredOrder:      nil,
		CancelChan:           make(chan string, 10),
		store:                store,
		ParentOrder:          po,
		ExecVenueId:          execVenueId,
		orderRouter:          orderRouter,
		childOrderStream:     childOrderStream,
		ChildOrderUpdateChan: childOrderStream.Chan(),
		doneChan:             doneChan,
		Log:                  slog.Default().With("order", po.Id),
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
		OrderSide:          side,
		Quantity:           quantity,
		Price:              price,
		ListingId:          listingId,
		Destination:        destination,
		OriginatorId:       om.ExecVenueId,
		OriginatorRef:      om.getStrategyOrderId(),
		RootOriginatorId:   om.ParentOrder.RootOriginatorId,
		RootOriginatorRef:  om.ParentOrder.RootOriginatorRef,
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

	if err = om.ParentOrder.OnChildOrderUpdate(pendingOrder); err != nil {
		return fmt.Errorf("failed to update parent order %s with pending child order:%w", om.ParentOrder.Id, err)
	}

	return nil
}

// CheckIfDone must be called in the strategies event handling loop, see example strategies as per package documentation.
func (om *Strategy) CheckIfDone(ctx context.Context) (done bool, err error) {
	done = false
	err = om.persistParentOrderChanges(ctx)
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
func (om *Strategy) OnChildOrderUpdate(childUpdatesChannelOpen bool, co *model.Order) error {
	if childUpdatesChannelOpen {
		if err := om.ParentOrder.OnChildOrderUpdate(co); err != nil {
			return fmt.Errorf("failed to update parent order %s with child order %s update:%w", om.ParentOrder.Id, co.Id, err)
		}
	} else {
		msg := "child order update chan unexpectedly closed, cancelling order"
		om.Log.Info(msg)
		om.CancelChan <- msg
	}

	return nil
}

// CancelChildOrdersAndStrategyOrder should only be called from with the strategy's event processing loop as per the example strategies
func (om *Strategy) CancelChildOrdersAndStrategyOrder() error {
	if !om.ParentOrder.IsTerminalState() {
		om.Log.Info("cancelling order")
		err := om.ParentOrder.SetTargetStatus(model.OrderStatus_CANCELLED)

		if err != nil {
			return fmt.Errorf("failed to Cancel order:%w", err)
		}

		pendingChildOrderCancels := false
		for _, co := range om.ParentOrder.ChildOrders {
			if !co.IsTerminalState() {
				pendingChildOrderCancels = true
				_, err := om.orderRouter.CancelOrder(context.Background(), &executionvenue.CancelOrderParams{
					OrderId:   co.Id,
					ListingId: co.ListingId,
					OwnerId:   co.OwnerId,
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

func (om *Strategy) persistParentOrderChanges(ctx context.Context) error {

	orderAsBytes, err := proto.Marshal(&om.ParentOrder.Order)

	if !bytes.Equal(om.lastStoredOrder, orderAsBytes) {

		if om.lastStoredOrder != nil {
			om.ParentOrder.Version = om.ParentOrder.Version + 1
		}

		toStore, err := proto.Marshal(&om.ParentOrder.Order)
		if err != nil {
			return fmt.Errorf("failed to marshal order:%w", err)
		}

		om.lastStoredOrder = toStore

		orderCopy := &model.Order{}
		if err = proto.Unmarshal(toStore, orderCopy); err != nil {
			return fmt.Errorf("failed to unmarshal order:%w", err)
		}

		if err = om.store(ctx, orderCopy); err != nil {
			return fmt.Errorf("failed to store order:%w", err)
		}
	}

	return err
}

func (om *Strategy) getStrategyOrderId() string {
	return om.ParentOrder.GetId()
}
