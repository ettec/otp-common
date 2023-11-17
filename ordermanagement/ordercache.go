package ordermanagement

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ettec/otp-common/model"
	"github.com/golang/protobuf/proto"
)

// OrderCache is an in memory cache of the latest order state for a given owner backed by a persistent store.
type OrderCache struct {
	store orderStore
	cache map[string]*model.Order
}

type orderStore interface {
	Write(ctx context.Context, order *model.Order) error
	LoadOrders(ctx context.Context, where func(order *model.Order) bool) (map[string]*model.Order, error)
	Close()
}

func NewOwnerOrderCache(ctx context.Context, ownerId string, store orderStore) (*OrderCache, error) {
	orderCache := OrderCache{
		store: store,
	}

	var err error
	orderCache.cache, err = store.LoadOrders(ctx, func(o *model.Order) bool {
		return o.OwnerId == ownerId
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load orders: %w", err)
	}

	return &orderCache, nil
}

func (oc *OrderCache) Store(ctx context.Context, order *model.Order) error {

	orderAsBytes, err := proto.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	orderCopy := &model.Order{}
	if err = proto.Unmarshal(orderAsBytes, orderCopy); err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	existingOrder, exists := oc.cache[orderCopy.Id]
	if exists {
		existingOrderAsBytes, err := proto.Marshal(existingOrder)
		if err != nil {
			return fmt.Errorf("failed to marshal existing order: %w", err)
		}

		if bytes.Equal(existingOrderAsBytes, orderAsBytes) {
			// no change, so do not store the order
			return nil
		}

		orderCopy.Version = existingOrder.Version + 1
	}

	if err := oc.store.Write(ctx, orderCopy); err != nil {
		return fmt.Errorf("failed to write order to store: %w", err)
	}

	oc.cache[orderCopy.Id] = orderCopy

	return nil
}

// GetOrder returns the order and true if found, otherwise a nil value and false
func (oc *OrderCache) GetOrder(orderId string) (*model.Order, bool, error) {
	order, exists := oc.cache[orderId]

	if exists {
		orderAsBytes, err := proto.Marshal(order)
		if err != nil {
			return nil, false, fmt.Errorf("failed to unmarshal order:%w", err)
		}
		orderCopy := &model.Order{}
		if err = proto.Unmarshal(orderAsBytes, orderCopy); err != nil {
			return nil, false, fmt.Errorf("failed to unmarshal order:%w", err)
		}

		return orderCopy, true, nil
	} else {
		return nil, false, nil
	}
}

func (oc *OrderCache) Close() {
	oc.store.Close()
}
