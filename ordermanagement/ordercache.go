package ordermanagement

import (
	"bytes"
	"fmt"
	"github.com/ettec/otp-common/model"
	"github.com/ettec/otp-common/orderstore"
	"github.com/golang/protobuf/proto"
)

type OrderCache struct {
	store orderstore.OrderStore
	cache map[string]*model.Order
}

// In memory cache of the latest order state backed by a store.
func NewOrderCache(store orderstore.OrderStore, ownerId string) (*OrderCache, error) {
	orderCache := OrderCache{
		store: store,
	}

	var err error
	orderCache.cache, err = store.RecoverInitialCache(func(o * model.Order) bool {
		return o.OwnerId == ownerId
	})
	if err != nil {
		return nil, err
	}

	return &orderCache, nil
}

func (oc *OrderCache) Store(order *model.Order) error {

	orderAsBytes, err := proto.Marshal(order)
	if err != nil {
		return err
	}
	orderCopy := &model.Order{}
	err = proto.Unmarshal(orderAsBytes, orderCopy)
	if err != nil {
		return err
	}
	order = orderCopy


	existingOrder, exists := oc.cache[order.Id]
	if exists {

		orderAsBytes, err := proto.Marshal(order)
		if err != nil {
			return fmt.Errorf("failed to compare order: %w", err)
		}

		existingOrderAsBytes, err := proto.Marshal(existingOrder)
		if err != nil {
			return fmt.Errorf("failed to compare order: %w", err)
		}

		if bytes.Equal(existingOrderAsBytes, orderAsBytes)  {
			// no change, so do not store the order
			return nil
		}

		order.Version = existingOrder.Version + 1
	}

	e := oc.store.Write(order)
	if e != nil {
		return e
	}

	oc.cache[order.Id] = order

	return nil
}

// Returns the order and true if found, otherwise a nil value and false
func (oc *OrderCache) GetOrder(orderId string) (*model.Order, bool, error) {
	order, ok := oc.cache[orderId]

	if ok {
		orderAsBytes, err := proto.Marshal(order)
		if err != nil {
			return nil, false, fmt.Errorf("failed to unmarshal order:%w", err)
		}
		orderCopy := &model.Order{}
		err = proto.Unmarshal(orderAsBytes, orderCopy)

		return orderCopy, true, nil
	} else {
		return nil, false, nil
	}
}

func (oc *OrderCache) Close() {
	oc.store.Close()
}
