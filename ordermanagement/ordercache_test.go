package ordermanagement

import (
	"context"
	"github.com/ettec/otp-common/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testOrderStore struct {
	writes []*model.Order
}

func (t *testOrderStore) Write(ctx context.Context, order *model.Order) error {
	t.writes = append(t.writes, order)
	return nil
}

func (t *testOrderStore) LoadOrders(ctx context.Context, loadOrder func(order *model.Order) bool) (map[string]*model.Order, error) {
	return map[string]*model.Order{}, nil
}

func (t *testOrderStore) Close() {
	panic("implement me")
}

func TestModifyingStoredOrderDoesNotModifyCacheOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testStore := &testOrderStore{}
	cache, err := NewOwnerOrderCache(ctx, "testid", testStore)
	assert.NoError(t, err)

	order := &model.Order{Id: "1", Status: model.OrderStatus_NONE}
	err = cache.Store(ctx, order)
	assert.NoError(t, err)

	order.Status = model.OrderStatus_FILLED

	order2, _, _ := cache.GetOrder("1")
	assert.Equal(t, model.OrderStatus_NONE, order2.Status)
}

func TestModifyingRetrievedOrderDoesNotModifyCacheOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testStore := &testOrderStore{}
	cache, err := NewOwnerOrderCache(ctx, "testid", testStore)
	assert.NoError(t, err)

	order := &model.Order{Id: "1", Status: model.OrderStatus_NONE}
	err = cache.Store(ctx, order)
	assert.NoError(t, err)

	order, _, _ = cache.GetOrder("1")

	order.Status = model.OrderStatus_FILLED

	order2, _, _ := cache.GetOrder("1")
	assert.Equal(t, model.OrderStatus_NONE, order2.Status)
}

func TestOrderIsNotStoredIfIdenticalToExistingOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testStore := &testOrderStore{}
	cache, _ := NewOwnerOrderCache(ctx, "testid", testStore)

	order := &model.Order{Id: "1", Status: model.OrderStatus_NONE}
	err := cache.Store(ctx, order)
	assert.NoError(t, err)

	order, _, _ = cache.GetOrder("1")
	err = cache.Store(ctx, order)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(testStore.writes))

	order.Status = model.OrderStatus_LIVE

	err = cache.Store(ctx, order)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(testStore.writes))
}
