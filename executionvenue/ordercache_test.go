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

func TestOrderCache_Store_DoesNotStoreDuplicateOrder(t *testing.T) {
	testStore := &testOrderStore{}
	cache, _:= NewOrderCache( testStore, "testid")

	order := &model.Order{Id: "1"}
	cache.Store(order)
	cache.Store(order)

	if len(testStore.writes) != 1 {
		t.FailNow()
	}

}
