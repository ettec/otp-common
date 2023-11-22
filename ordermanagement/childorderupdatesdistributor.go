package ordermanagement

import (
	"context"
	"github.com/ettec/otp-common/model"
	"log/slog"
)

type childOrderSubscriptionRequest struct {
	parentOrderId string
	orderChan     chan *model.Order
}

type OrderStream interface {
	Chan() <-chan *model.Order
	Close()
}

// childOrderUpdatesDistributor is used to subscribe to child order updates for a given parent order id.
type childOrderUpdatesDistributor struct {
	subscribeChan         chan childOrderSubscriptionRequest
	unsubscribeChan       chan string
	childOrderUpdatesChan <-chan ChildOrder
}

func NewChildOrderUpdatesDistributor(updates <-chan ChildOrder, initialSubscriptionsBufferSize int) *childOrderUpdatesDistributor {

	d := &childOrderUpdatesDistributor{
		subscribeChan:         make(chan childOrderSubscriptionRequest, initialSubscriptionsBufferSize),
		unsubscribeChan:       make(chan string, initialSubscriptionsBufferSize),
		childOrderUpdatesChan: updates,
	}

	return d
}

func (d *childOrderUpdatesDistributor) GetChildOrderStream(parentOrderId string, bufferSize int) OrderStream {
	return newChildOrderStream(parentOrderId, bufferSize, d)
}

func (d *childOrderUpdatesDistributor) Start(ctx context.Context) {
	parentIdToChan := map[string]chan *model.Order{}

	go func() {
		defer func() {
			for _, c := range parentIdToChan {
				close(c)
			}
		}()

	initialSubscriptionsLoop:
		for {
			select {
			case <-ctx.Done():
				return
			case subscriptionRequest := <-d.subscribeChan:
				parentIdToChan[subscriptionRequest.parentOrderId] = subscriptionRequest.orderChan
			default:
				break initialSubscriptionsLoop
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case u := <-d.childOrderUpdatesChan:
				if orderChan, ok := parentIdToChan[u.ParentOrderId]; ok {
					select {
					case orderChan <- u.Child:
					default:
						slog.Warn("slow consumer, closing child order update channel", "parentOrderId", u.ParentOrderId)
						close(orderChan)
						delete(parentIdToChan, u.ParentOrderId)
					}

				}
			case subscriptionRequest := <-d.subscribeChan:
				parentIdToChan[subscriptionRequest.parentOrderId] = subscriptionRequest.orderChan
			case parentId := <-d.unsubscribeChan:
				updateChan := parentIdToChan[parentId]
				delete(parentIdToChan, parentId)
				close(updateChan)
			}
		}

	}()
}

type childOrderStream struct {
	parentOrderId string
	orderChan     chan *model.Order
	distributor   *childOrderUpdatesDistributor
}

func newChildOrderStream(parentOrderId string, bufferSize int, d *childOrderUpdatesDistributor) *childOrderStream {
	stream := &childOrderStream{parentOrderId: parentOrderId, orderChan: make(chan *model.Order, bufferSize), distributor: d}
	d.subscribeChan <- childOrderSubscriptionRequest{
		parentOrderId: parentOrderId,
		orderChan:     stream.orderChan,
	}
	return stream
}

func (c *childOrderStream) Chan() <-chan *model.Order {
	return c.orderChan
}

func (c *childOrderStream) Close() {
	c.distributor.unsubscribeChan <- c.parentOrderId
}
