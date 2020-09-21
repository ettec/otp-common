package executionvenue

import (
	"github.com/ettec/otp-common/orderstore"
	"github.com/ettec/otp-common/model"
)

type OrderCache struct {
	store orderstore.OrderStore
	cache map[string]*model.Order
}

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

	existingOrder, exists := oc.cache[order.Id]
	if exists {
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
func (oc *OrderCache) GetOrder(orderId string) (*model.Order, bool) {
	order, ok := oc.cache[orderId]

	if ok {
		return order, true
	} else {
		return nil, false
	}
}

func (oc *OrderCache) Close() {
	oc.store.Close()
}
