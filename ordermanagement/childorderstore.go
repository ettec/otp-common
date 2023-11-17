// Types and functions to be used by services that manage orders.
package ordermanagement

import (
	"context"
	"github.com/ettec/otp-common/model"
	"github.com/golang/protobuf/proto"
	"github.com/segmentio/kafka-go"
	"log/slog"
)

type ChildOrder struct {
	ParentOrderId string
	Child         *model.Order
}

type orderReader interface {
	Close() error
	ReadMessage(ctx context.Context) (kafka.Message, error)
}

// GetChildOrders returns a channel of orders whose originator id matches the given order id and are thus children of the order.
func GetChildOrders(ctx context.Context, id string, kafkaReaderConfig kafka.ReaderConfig, bufferSize int) (<-chan ChildOrder, error) {
	return getChildOrdersFromReader(ctx, id, kafka.NewReader(kafkaReaderConfig), bufferSize)
}

func getChildOrdersFromReader(ctx context.Context, parentOriginatorId string, reader orderReader, bufferSize int) (<-chan ChildOrder, error) {
	isChildOrder := func(order *model.Order) bool {
		return parentOriginatorId == order.GetOriginatorId()
	}

	getParentOrderId := func(order *model.Order) string {
		return order.OriginatorRef
	}

	out := make(chan ChildOrder, bufferSize)

	go func() {
		defer func() {
			close(out)
			if err := reader.Close(); err != nil {
				slog.Error("failed to close kafka reader", "error", err)
			}
		}()

		for {

			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				slog.Error("failed to read child orders from kafka store", "error", err)
				return
			}

			order := &model.Order{}
			err = proto.Unmarshal(msg.Value, order)
			if err != nil {
				slog.Error("failed to unmarshal order", "error", err)
				return
			}

			if isChildOrder(order) {
				out <- ChildOrder{
					ParentOrderId: getParentOrderId(order),
					Child:         order,
				}
			}
		}
	}()

	return out, nil
}
