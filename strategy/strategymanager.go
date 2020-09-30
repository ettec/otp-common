package strategy

import (
	"context"
	"fmt"
	"github.com/ettec/otp-common/api/executionvenue"
	"github.com/ettec/otp-common/bootstrap"
	"github.com/ettec/otp-common/model"
	"github.com/ettec/otp-common/ordermanagement"
	"github.com/ettec/otp-common/orderstore"
	"github.com/google/uuid"
	"log"
	"sync"
)





type ChildOrderUpdates interface {
	Start()
	NewOrderStream(parentOrderId string, bufferSize int) ordermanagement.ChildOrderStream
}



type strategyManager struct {
	id                string
	store             orderstore.OrderStore
	orderRouter       executionvenue.ExecutionVenueClient
	doneChan          chan string
	orders            sync.Map
	childOrderUpdates ChildOrderUpdates
	executeFn         func(om *Strategy)
	inboundChildUpdatesBufferSize int
}

// StrategyManager is responsible for the creation and recovery of new strategy instances and for forwarding requests
// to a strategy service to the target strategy.
func NewStrategyManager(id string, parentOrderStore orderstore.OrderStore, childOrderUpdates ChildOrderUpdates,
	orderRouter executionvenue.ExecutionVenueClient, executeFn func(om *Strategy)) *strategyManager {

	sm := &strategyManager{
		id:                id,
		store:             parentOrderStore,
		orderRouter:       orderRouter,
		doneChan:          make(chan string, 100),
		orders:            sync.Map{},
		childOrderUpdates: childOrderUpdates,
		inboundChildUpdatesBufferSize: bootstrap.GetOptionalIntEnvVar("STRATEGYMANAGER_INBOUND_CHILD_UPDATES_BUFFER_SIZE", 1000),
	}

	sm.executeFn = executeFn

	go func() {
		id := <-sm.doneChan
		sm.orders.Delete(id)
		log.Printf("order %v done", id)
	}()

	parentOrders, err := sm.store.RecoverInitialCache(func(o *model.Order) bool {
		return o.OwnerId == id
	})
	if err != nil {
		panic(err)
	}

	for _, order := range parentOrders {
		if !order.IsTerminalState() {

			om := NewStrategyFromParentOrder(order, sm.store.Write, sm.id, sm.orderRouter,
				sm.childOrderUpdates.NewOrderStream(order.Id, sm.inboundChildUpdatesBufferSize ),
				sm.doneChan)
			sm.orders.Store(om.GetParentOrderId(), om)

			sm.executeFn(om)
		}
	}

	sm.childOrderUpdates.Start()
	return sm
}

func (s *strategyManager) GetExecutionParametersMetaData(_ context.Context, empty *model.Empty) (*executionvenue.ExecParamsMetaDataJson, error) {
	return nil, fmt.Errorf("not implemented for strategy manager")
}

func (s *strategyManager) CreateAndRouteOrder(_ context.Context, params *executionvenue.CreateAndRouteOrderParams) (*executionvenue.OrderId, error) {

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	om, err := NewStrategyFromCreateParams(id.String(), params, s.id, s.store.Write, s.orderRouter,
		s.childOrderUpdates.NewOrderStream(id.String(), s.inboundChildUpdatesBufferSize), s.doneChan)

	if err != nil {
		return nil, err
	}

	s.orders.Store(om.GetParentOrderId(), om)

	s.executeFn(om)

	return &executionvenue.OrderId{
		OrderId: id.String(),
	}, nil
}

func (s *strategyManager) ModifyOrder(_ context.Context, _ *executionvenue.ModifyOrderParams) (*model.Empty, error) {
	return nil, fmt.Errorf("order modification not supported")
}

func (s *strategyManager) CancelOrder(_ context.Context, params *executionvenue.CancelOrderParams) (*model.Empty, error) {

	if val, exists := s.orders.Load(params.OrderId); exists {
		om := val.(*Strategy)
		om.Cancel()
		return &model.Empty{}, nil
	} else {
		return nil, fmt.Errorf("no order found for id:%v", params.OrderId)
	}
}
