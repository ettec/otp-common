package strategy

import (
	"context"
	"fmt"
	"github.com/ettec/otp-common/api/executionvenue"
	"github.com/ettec/otp-common/bootstrap"
	"github.com/ettec/otp-common/model"
	"github.com/ettec/otp-common/ordermanagement"
	"github.com/google/uuid"
	"log/slog"
	"sync"
)

type childOrderUpdates interface {
	Start(ctx context.Context)
	GetChildOrderStream(parentOrderId string, bufferSize int) ordermanagement.OrderStream
}

type orderStore interface {
	Write(ctx context.Context, order *model.Order) error
	LoadOrders(ctx context.Context, loadOrder func(order *model.Order) bool) (map[string]*model.Order, error)
	Close()
}

// Manager is responsible for the creation and recovery of new strategy instances and for forwarding requests
// to a strategy service to the target strategy.
type Manager struct {
	id                            string
	store                         orderStore
	orderRouter                   executionvenue.ExecutionVenueClient
	doneChan                      chan string
	orders                        sync.Map
	childOrderUpdates             childOrderUpdates
	executeFn                     func(om *Strategy)
	inboundChildUpdatesBufferSize int
}

func NewStrategyManager(ctx context.Context, id string, parentOrderStore orderStore, childOrderUpdates childOrderUpdates,
	orderRouter executionvenue.ExecutionVenueClient, executeFn func(om *Strategy)) (*Manager, error) {

	sm := &Manager{
		id:                            id,
		store:                         parentOrderStore,
		orderRouter:                   orderRouter,
		doneChan:                      make(chan string, 100),
		orders:                        sync.Map{},
		childOrderUpdates:             childOrderUpdates,
		inboundChildUpdatesBufferSize: bootstrap.GetOptionalIntEnvVar("STRATEGYMANAGER_INBOUND_CHILD_UPDATES_BUFFER_SIZE", 1000),
	}

	sm.executeFn = executeFn

	go func() {
		id := <-sm.doneChan
		sm.orders.Delete(id)
		slog.Info("order done", "id", id)
	}()

	parentOrders, err := sm.store.LoadOrders(ctx, func(o *model.Order) bool {
		return o.OwnerId == id
	})
	if err != nil {
		return nil, fmt.Errorf("failed to recover initial orders: %w", err)
	}

	for _, order := range parentOrders {
		if !order.IsTerminalState() {

			om := NewStrategyFromParentOrder(order, sm.store.Write, sm.id, sm.orderRouter,
				sm.childOrderUpdates.GetChildOrderStream(order.Id, sm.inboundChildUpdatesBufferSize),
				sm.doneChan)
			sm.orders.Store(om.getStrategyOrderId(), om)

			sm.executeFn(om)
		}
	}

	sm.childOrderUpdates.Start(ctx)
	return sm, nil
}

func (s *Manager) GetExecutionParametersMetaData(_ context.Context, empty *model.Empty) (*executionvenue.ExecParamsMetaDataJson, error) {
	return nil, fmt.Errorf("not implemented for strategy manager")
}

func (s *Manager) CreateAndRouteOrder(_ context.Context, params *executionvenue.CreateAndRouteOrderParams) (*executionvenue.OrderId, error) {

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to create unique id:%w", err)
	}

	om, err := NewStrategyFromCreateParams(id.String(), params, s.id, s.store.Write, s.orderRouter,
		s.childOrderUpdates.GetChildOrderStream(id.String(), s.inboundChildUpdatesBufferSize), s.doneChan)

	if err != nil {
		return nil, fmt.Errorf("failed to create strategy from parameters: %w", err)
	}

	s.orders.Store(om.getStrategyOrderId(), om)

	s.executeFn(om)

	return &executionvenue.OrderId{
		OrderId: id.String(),
	}, nil
}

func (s *Manager) ModifyOrder(_ context.Context, _ *executionvenue.ModifyOrderParams) (*model.Empty, error) {
	return nil, fmt.Errorf("order modification not supported")
}

func (s *Manager) CancelOrder(_ context.Context, params *executionvenue.CancelOrderParams) (*model.Empty, error) {

	if val, exists := s.orders.Load(params.OrderId); exists {
		om := val.(*Strategy)
		om.CancelChan <- ""
		return &model.Empty{}, nil
	} else {
		return nil, fmt.Errorf("no order found for id:%v", params.OrderId)
	}
}
