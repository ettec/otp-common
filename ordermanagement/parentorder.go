package ordermanagement

import (
	"fmt"
	"github.com/ettec/otp-common/model"
)

var zero *model.Decimal64

func init() {
	zero = &model.Decimal64{}
}

// ParentOrder has a one-to-many relationship to its child orders.  The parent order aggregates and summarises state across
// its children.
type ParentOrder struct {
	model.Order
	ChildOrders          map[string]*model.Order
	executions           map[string]*model.Execution
	childOrderRefs       map[string]model.Ref
	childOrdersRecovered bool
}

func NewParentOrder(order model.Order) *ParentOrder {

	childOrderRefs := map[string]model.Ref{}
	for _, ref := range order.ChildOrdersRefs {
		childOrderRefs[ref.Id] = *ref
	}

	return &ParentOrder{
		order,
		map[string]*model.Order{},
		map[string]*model.Execution{},
		childOrderRefs,
		false,
	}
}

func (po *ParentOrder) OnChildOrderUpdate(childOrder *model.Order) error {

	po.ChildOrders[childOrder.Id] = childOrder

	var newExecution *model.Execution

	if childOrder.LastExecId != "" {
		if _, exists := po.executions[childOrder.LastExecId]; !exists {
			newExecution = &model.Execution{
				Id:    childOrder.LastExecId,
				Price: *childOrder.LastExecPrice,
				Qty:   *childOrder.LastExecQuantity,
			}

			po.executions[childOrder.LastExecId] = newExecution
		}
	}

	if !po.childOrdersRecovered {
		po.childOrdersRecovered = true
		for _, persistedRef := range po.ChildOrdersRefs {
			if order, exists := po.ChildOrders[persistedRef.Id]; !exists || persistedRef.Version > order.Version {
				po.childOrdersRecovered = false
				break
			}
		}
	}

	if ref, exists := po.childOrderRefs[childOrder.Id]; exists {
		if childOrder.Version <= ref.Version {
			// Ignore this child order update as it is before the latest referenced version
			return nil
		} else {
			newRef := model.Ref{Id: childOrder.Id, Version: childOrder.Version}
			po.childOrderRefs[childOrder.Id] = newRef
			foundIdx := -1
			for idx, existingRef := range po.ChildOrdersRefs {
				if existingRef.Id == newRef.Id {
					foundIdx = idx
					break
				}
			}

			po.ChildOrdersRefs[foundIdx] = &newRef
		}
	} else {
		newRef := model.Ref{Id: childOrder.Id, Version: childOrder.Version}
		po.childOrderRefs[childOrder.Id] = newRef
		po.ChildOrdersRefs = append(po.ChildOrdersRefs, &newRef)
	}

	if newExecution != nil {
		err := po.AddExecution(*newExecution)
		if err != nil {
			return fmt.Errorf("failed to add execution to parent order %s:%w", po.Id, err)
		}
	}

	exposedQnt := model.IasD(0)
	for _, order := range po.ChildOrders {
		if !order.IsTerminalState() {
			exposedQnt.Add(order.RemainingQuantity)
		}
	}

	if !po.ExposedQuantity.Equal(exposedQnt) {
		po.ExposedQuantity = exposedQnt
	}

	if po.GetTargetStatus() == model.OrderStatus_CANCELLED {
		if po.GetExposedQuantity().Equal(zero) {
			if err := po.SetStatus(model.OrderStatus_CANCELLED); err != nil {
				return fmt.Errorf("failed to set parent order %s status to cancelled: %w", po.Id, err)
			}
		}
	}

	return nil
}
