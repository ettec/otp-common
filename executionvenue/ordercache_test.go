package executionvenue

import (
	"github.com/ettec/otp-common/model"
	"testing"
)

type  testOrderStore struct {
	writes []*model.Order
}

func (t *testOrderStore) Write(order *model.Order) error {
	t.writes = append(t.writes, order)
	return nil
}

func (t *testOrderStore) RecoverInitialCache(loadOrder func(order *model.Order) bool) (map[string]*model.Order, error) {
	return map[string]*model.Order{}, nil
}

func (t *testOrderStore) Close() {
	panic("implement me")
}


func TestModifyingStoredOrderDoesNotModifyCacheOrder(t *testing.T) {

	testStore := &testOrderStore{}
	cache, _:= NewOrderCache( testStore, "testid")

	order := &model.Order{Id: "1", Status: model.OrderStatus_NONE}
	cache.Store(order)

	order.Status = model.OrderStatus_FILLED

	order2, _,_ :=cache.GetOrder("1")
	if order2.Status != model.OrderStatus_NONE {
		t.FailNow()
	}

}


func TestModifyingRetrievedOrderDoesNotModifyCacheOrder(t *testing.T) {

	testStore := &testOrderStore{}
	cache, _:= NewOrderCache( testStore, "testid")

	order := &model.Order{Id: "1", Status: model.OrderStatus_NONE}
	cache.Store(order)

	order, _,_ =cache.GetOrder("1")

	order.Status = model.OrderStatus_FILLED

	order2, _,_ :=cache.GetOrder("1")
	if order2.Status != model.OrderStatus_NONE {
		t.FailNow()
	}

}

func TestOrderCache_Store_DoesNotStoreOrderIfIdentical(t *testing.T) {
	testStore := &testOrderStore{}
	cache, _:= NewOrderCache( testStore, "testid")

	order := &model.Order{Id: "1", Status: model.OrderStatus_NONE}
	cache.Store(order)

	order, _,_ =cache.GetOrder("1")
	cache.Store(order)

	if len(testStore.writes) != 1 {
		t.FailNow()
	}



	order.Status =  model.OrderStatus_LIVE

	cache.Store(order)

	if len(testStore.writes) != 2 {
		t.FailNow()
	}

}
