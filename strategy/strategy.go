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
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"strconv"
	"time"
)

type Strategy struct {
	// These two channels must be checked and handled as part of the event loop
	CancelChan           chan string
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

func (om *Strategy) Cancel() {
	om.CancelChan <- ""
}

func (om *Strategy) GetParentOrderId() string {
	return om.ParentOrder.GetId()
}

func (om *Strategy) CancelParentOrder() error {
	if !om.ParentOrder.IsTerminalState() {
		om.Log.Print("cancelling order")
		err := om.ParentOrder.SetTargetStatus(model.OrderStatus_CANCELLED)

		if err != nil {
			return fmt.Errorf("failed to cancel order:%w", err)
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
					return fmt.Errorf("failed to cancel child order:%w", err)
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

func (om *Strategy) SendChildOrder(side model.Side, quantity *model.Decimal64, price *model.Decimal64, listing *model.Listing) error {

	if quantity.GreaterThan(om.ParentOrder.GetAvailableQty()) {
		return fmt.Errorf("cannot send child order for %v as it exceeds the available quantity on the parent order: %v", quantity,
			om.ParentOrder.GetAvailableQty())
	}

	params := &executionvenue.CreateAndRouteOrderParams{
		OrderSide:         side,
		Quantity:          quantity,
		Price:             price,
		ListingId:         listing.Id,
		Destination:       listing.Market.Mic,
		OriginatorId:      om.ExecVenueId,
		OriginatorRef:     om.GetParentOrderId(),
		RootOriginatorId:  om.ParentOrder.RootOriginatorId,
		RootOriginatorRef: om.ParentOrder.RootOriginatorRef,
	}

	id, err := om.orderRouter.CreateAndRouteOrder(context.Background(), params)

	if err != nil {
		return fmt.Errorf("failed to submit child order:%w", err)
	}

	pendingOrder := model.NewOrder(id.OrderId, params.OrderSide, params.Quantity, params.Price, params.ListingId,
		om.ExecVenueId, om.GetParentOrderId(), om.ParentOrder.RootOriginatorId, om.ParentOrder.RootOriginatorRef, listing.Market.Mic)

	// First persisted orders start at version 0, this is a placeholder until the first child order update is received
	pendingOrder.Version = -1

	om.ParentOrder.OnChildOrderUpdate(pendingOrder)

	return nil
}

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

func (om *Strategy) OnChildOrderUpdate(ok bool, co *model.Order) error {
	if ok {
		_, err := om.ParentOrder.OnChildOrderUpdate(co)
		return err
	} else {
		om.ErrLog.Printf("child order update chan unexpectedly closed, cancelling order")
		om.Cancel()
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

func (om *Strategy) CancelOrderWithErrorMsg(msg string) {
	om.Log.Printf("Cancelling order:%v", msg)
	om.CancelChan <- msg
}

func GetOrderRouter(clientSet *kubernetes.Clientset, maxConnectRetrySecs time.Duration) (executionvenue.ExecutionVenueClient, error) {
	namespace := "default"
	list, err := clientSet.CoreV1().Services(namespace).List(v1.ListOptions{
		LabelSelector: "app=order-router",
	})

	if err != nil {
		panic(err)
	}

	var client executionvenue.ExecutionVenueClient

	for _, service := range list.Items {

		var podPort int32
		for _, port := range service.Spec.Ports {
			if port.Name == "executionvenue" {
				podPort = port.Port
			}
		}

		if podPort == 0 {
			log.Printf("ignoring order router service as it does not have a port named executionvenue, service: %v", service)
			continue
		}

		targetAddress := service.Name + ":" + strconv.Itoa(int(podPort))

		log.Printf("connecting to order router service %v at: %v", service.Name, targetAddress)

		conn, err := grpc.Dial(targetAddress, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(maxConnectRetrySecs))

		if err != nil {
			panic(err)
		}

		client = executionvenue.NewExecutionVenueClient(conn)
		break
	}

	if client == nil {
		return nil, fmt.Errorf("failed to find order router")
	}

	return client, nil
}