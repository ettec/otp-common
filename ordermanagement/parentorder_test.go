package ordermanagement

import (
	"github.com/ettec/otp-common/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func IasD(i int) *model.Decimal64 {
	return &model.Decimal64{
		Mantissa: int64(i),
		Exponent: 0,
	}
}

func TestParentOrderCancel(t *testing.T) {

	po := NewParentOrder(*model.NewOrder("a", model.Side_BUY, IasD(20), IasD(50), 1, "oi",
		"or", "ri", "rr", "XNAS"))
	err := po.SetTargetStatus(model.OrderStatus_LIVE)
	assert.NoError(t, err)
	err = po.SetStatus(model.OrderStatus_LIVE)
	assert.NoError(t, err)

	err = po.OnChildOrderUpdate(&model.Order{Id: "a1", Version: 1, TargetStatus: model.OrderStatus_LIVE, Quantity: IasD(15), RemainingQuantity: IasD(15)})
	assert.NoError(t, err)

	if !po.ExposedQuantity.Equal(IasD(15)) {
		t.FailNow()
	}

	if !po.GetAvailableQty().Equal(IasD(5)) {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a1", Version: 2, TargetStatus: model.OrderStatus_NONE, Status: model.OrderStatus_LIVE, Quantity: IasD(15), RemainingQuantity: IasD(15)})
	assert.NoError(t, err)

	if !po.ExposedQuantity.Equal(IasD(15)) || !po.GetAvailableQty().Equal(IasD(5)) {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a2", Version: 1, TargetStatus: model.OrderStatus_LIVE, Quantity: IasD(5), RemainingQuantity: IasD(5)})
	assert.NoError(t, err)

	if !po.TradedQuantity.Equal(IasD(0)) || !po.ExposedQuantity.Equal(IasD(20)) || !po.GetAvailableQty().Equal(IasD(0)) {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a2", Version: 2, TargetStatus: model.OrderStatus_NONE, Status: model.OrderStatus_LIVE, Quantity: IasD(5), RemainingQuantity: IasD(5)})
	assert.NoError(t, err)

	if !po.TradedQuantity.Equal(IasD(0)) || !po.ExposedQuantity.Equal(IasD(20)) || !po.GetAvailableQty().Equal(IasD(0)) {
		t.FailNow()
	}

	err = po.SetTargetStatus(model.OrderStatus_CANCELLED)
	assert.NoError(t, err)

	err = po.OnChildOrderUpdate(&model.Order{Id: "a1", Version: 3, TargetStatus: model.OrderStatus_CANCELLED, Status: model.OrderStatus_LIVE, Quantity: IasD(15), RemainingQuantity: IasD(15)})
	assert.NoError(t, err)

	if !po.TradedQuantity.Equal(IasD(0)) || !po.ExposedQuantity.Equal(IasD(20)) ||
		!po.GetAvailableQty().Equal(IasD(0)) || po.GetStatus() != model.OrderStatus_LIVE {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a2", Version: 3, TargetStatus: model.OrderStatus_CANCELLED, Status: model.OrderStatus_LIVE, Quantity: IasD(5), RemainingQuantity: IasD(5)})
	assert.NoError(t, err)

	if !po.TradedQuantity.Equal(IasD(0)) || !po.ExposedQuantity.Equal(IasD(20)) ||
		!po.GetAvailableQty().Equal(IasD(0)) || po.GetStatus() != model.OrderStatus_LIVE {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a1", Version: 4, TargetStatus: model.OrderStatus_NONE, Status: model.OrderStatus_CANCELLED, Quantity: IasD(15), RemainingQuantity: IasD(15)})
	assert.NoError(t, err)

	if !po.TradedQuantity.Equal(IasD(0)) || !po.ExposedQuantity.Equal(IasD(5)) ||
		!po.GetAvailableQty().Equal(IasD(15)) || po.GetStatus() != model.OrderStatus_LIVE {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a2", Version: 4, TargetStatus: model.OrderStatus_NONE, Status: model.OrderStatus_CANCELLED, Quantity: IasD(5), RemainingQuantity: IasD(5)})
	assert.NoError(t, err)

	if !po.TradedQuantity.Equal(IasD(0)) || !po.ExposedQuantity.Equal(IasD(0)) ||
		!po.GetAvailableQty().Equal(IasD(20)) || po.GetStatus() != model.OrderStatus_CANCELLED {
		t.FailNow()
	}

}

func TestParentOrderUpdatesWhenChildOrdersFilled(t *testing.T) {

	po := NewParentOrder(*model.NewOrder("a", model.Side_BUY, IasD(20), IasD(50), 1, "oi", "or", "ri",
		"rr", "XNAS"))
	err := po.SetTargetStatus(model.OrderStatus_LIVE)
	assert.NoError(t, err)

	err = po.OnChildOrderUpdate(&model.Order{Id: "a1", Version: 1, TargetStatus: model.OrderStatus_LIVE, Quantity: IasD(15), RemainingQuantity: IasD(15)})
	assert.NoError(t, err)

	if !po.ExposedQuantity.Equal(IasD(15)) {
		t.FailNow()
	}

	if !po.GetAvailableQty().Equal(IasD(5)) {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a2", Version: 1, TargetStatus: model.OrderStatus_LIVE, Quantity: IasD(5), RemainingQuantity: IasD(5)})
	assert.NoError(t, err)

	if !po.ExposedQuantity.Equal(IasD(20)) {
		t.FailNow()
	}

	if !po.GetAvailableQty().Equal(IasD(0)) {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a1", Version: 2, LastExecPrice: IasD(50), LastExecQuantity: IasD(5),
		LastExecId: "a1e1", RemainingQuantity: IasD(10)})
	assert.NoError(t, err)

	if !po.TradedQuantity.Equal(IasD(5)) {
		t.FailNow()
	}

	if !po.ExposedQuantity.Equal(IasD(15)) {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a2", Version: 2, TargetStatus: model.OrderStatus_LIVE, Quantity: IasD(5), RemainingQuantity: IasD(0),
		LastExecPrice: IasD(50), LastExecQuantity: IasD(5),
		LastExecId: "a2e1"})
	assert.NoError(t, err)

	if !po.TradedQuantity.Equal(IasD(10)) {
		t.FailNow()
	}

	if !po.ExposedQuantity.Equal(IasD(10)) {
		t.FailNow()
	}

	if !po.GetAvailableQty().Equal(IasD(0)) {
		t.FailNow()
	}

	err = po.OnChildOrderUpdate(&model.Order{Id: "a1", Version: 3, LastExecPrice: IasD(50), LastExecQuantity: IasD(10),
		LastExecId: "a1e2", RemainingQuantity: IasD(0)})
	assert.NoError(t, err)

	if !po.TradedQuantity.Equal(IasD(20)) {
		t.FailNow()
	}

	if !po.ExposedQuantity.Equal(IasD(0)) {
		t.FailNow()
	}

	if !po.GetAvailableQty().Equal(IasD(0)) {
		t.FailNow()
	}

	if po.GetStatus() != model.OrderStatus_FILLED {
		t.FailNow()
	}

}

func TestParentOrderRecovery(t *testing.T) {

	po := NewParentOrder(*model.NewOrder("a", model.Side_BUY, IasD(20), IasD(50), 1, "oi", "or", "ri",
		"rr", "XNAS"))

	preFailureUpdates := []*model.Order{
		{Id: "a1", Version: 1, TargetStatus: model.OrderStatus_LIVE, Quantity: IasD(15), RemainingQuantity: IasD(15)},
		{Id: "a2", Version: 1, TargetStatus: model.OrderStatus_LIVE, Quantity: IasD(5), RemainingQuantity: IasD(5)},
		{Id: "a1", Version: 2, LastExecPrice: IasD(50), LastExecQuantity: IasD(5),
			LastExecId: "a1e1", RemainingQuantity: IasD(10)},
		{Id: "a2", Version: 2, TargetStatus: model.OrderStatus_LIVE, Quantity: IasD(5), RemainingQuantity: IasD(0),
			LastExecPrice: IasD(50), LastExecQuantity: IasD(5),
			LastExecId: "a2e1"},
	}

	for _, update := range preFailureUpdates {
		err := po.OnChildOrderUpdate(update)
		assert.NoError(t, err)
	}

	if !po.TradedQuantity.Equal(IasD(10)) {
		t.FailNow()
	}

	if !po.ExposedQuantity.Equal(IasD(10)) {
		t.FailNow()
	}

	if !po.GetAvailableQty().Equal(IasD(0)) {
		t.FailNow()
	}

	var preFailureState model.Order
	preFailureState = po.Order

	recoveredOrder := NewParentOrder(preFailureState)

	if !recoveredOrder.TradedQuantity.Equal(IasD(10)) {
		t.FailNow()
	}

	if !recoveredOrder.ExposedQuantity.Equal(IasD(10)) {
		t.FailNow()
	}

	if !recoveredOrder.GetAvailableQty().Equal(IasD(0)) {
		t.FailNow()
	}

	numUpdates := len(preFailureUpdates)
	for idx, update := range preFailureUpdates {
		err := recoveredOrder.OnChildOrderUpdate(update)
		assert.NoError(t, err)
		if idx < numUpdates-1 {
			if recoveredOrder.childOrdersRecovered {
				t.FailNow()
			}
		} else {
			if !recoveredOrder.childOrdersRecovered {
				t.FailNow()
			}
		}

	}

	if !recoveredOrder.TradedQuantity.Equal(IasD(10)) {
		t.FailNow()
	}

	if !recoveredOrder.ExposedQuantity.Equal(IasD(10)) {
		t.FailNow()
	}

	if !recoveredOrder.GetAvailableQty().Equal(IasD(0)) {
		t.FailNow()
	}

	if len(recoveredOrder.executions) != 2 {
		t.FailNow()
	}

	err := recoveredOrder.OnChildOrderUpdate(&model.Order{Id: "a1", Version: 3, LastExecPrice: IasD(50), LastExecQuantity: IasD(10),
		LastExecId: "a1e2", RemainingQuantity: IasD(0)})
	assert.NoError(t, err)

	if !recoveredOrder.TradedQuantity.Equal(IasD(20)) {
		t.FailNow()
	}

	if !recoveredOrder.ExposedQuantity.Equal(IasD(0)) {
		t.FailNow()
	}

	if !recoveredOrder.GetAvailableQty().Equal(IasD(0)) {
		t.FailNow()
	}

	if recoveredOrder.GetStatus() != model.OrderStatus_FILLED {
		t.FailNow()
	}

	if len(recoveredOrder.executions) != 3 {
		t.FailNow()
	}

}
