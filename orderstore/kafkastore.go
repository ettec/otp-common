// Package orderstore contains implementations of stores used to store order updates.
package orderstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/ettec/otp-common/model"
	"github.com/golang/protobuf/proto"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"time"
)

type KafkaStore struct {
	writer            *kafka.Writer
	topic             string
	kafkaBrokerUrls   []string
	ownerId           string
	kafkaReaderConfig kafka.ReaderConfig
}

type orderReader interface {
	Close() error
	ReadMessage(ctx context.Context) (kafka.Message, error)
}

func DefaultReaderConfig(topic string, kafkaBrokerUrls []string) kafka.ReaderConfig {
	return kafka.ReaderConfig{
		Brokers:        kafkaBrokerUrls,
		Topic:          topic,
		ReadBackoffMin: 100 * time.Millisecond,
		ReadBackoffMax: 200 * time.Millisecond,
		MaxWait:        150 * time.Millisecond,
	}
}

func DefaultWriterConfig(topic string, kafkaBrokerUrls []string) kafka.WriterConfig {
	return kafka.WriterConfig{
		Brokers:      kafkaBrokerUrls,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		Async:        true,
		BatchTimeout: 10 * time.Millisecond,
	}
}

func NewKafkaStore(kafkaReaderConfig kafka.ReaderConfig, kafkaWriterConfig kafka.WriterConfig,
	ownerId string) (*KafkaStore, error) {

	topic := "orders"

	result := KafkaStore{

		topic:             topic,
		kafkaReaderConfig: kafkaReaderConfig,
		ownerId:           ownerId,
	}

	result.writer = kafka.NewWriter(kafkaWriterConfig)

	return &result, nil
}

func (ks *KafkaStore) LoadOrders(ctx context.Context, filter func(order *model.Order) bool) (map[string]*model.Order, error) {

	slog.Info("restoring order state")
	reader := ks.getNewReader()
	defer func() {
		if err := reader.Close(); err != nil {
			slog.Error("error closing kafka reader", "error", err)
		}
	}()

	result, err := loadOrdersFromReader(ctx, reader, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to load orders from reader: %w", err)
	}

	return result, nil
}

func (ks *KafkaStore) getNewReader() *kafka.Reader {
	return kafka.NewReader(ks.kafkaReaderConfig)
}

func (ks *KafkaStore) SubscribeToAllOrders(ctx context.Context, createdAfter time.Time, bufferSize int) (map[string]*model.Order, <-chan *model.Order, error) {
	reader := ks.getNewReader()

	afterTimestamp := &model.Timestamp{Seconds: createdAfter.Unix()}
	timeFilter := func(order *model.Order) bool {
		return order.GetCreated().After(afterTimestamp)
	}

	initialState, err := loadOrdersFromReader(ctx, reader, timeFilter)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to restore order state from store: %w", err)
	}

	outChan := make(chan<- *model.Order, bufferSize)
	go func() {
		defer func() {
			close(outChan)
			if err := reader.Close(); err != nil {
				slog.Error("error closing kafka reader", "error", err)
			}
		}()
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				slog.Error("failed to read message", "error", err)
				return
			}

			order := &model.Order{}
			if err = proto.Unmarshal(msg.Value, order); err != nil {
				slog.Error("failed to unmarshal order", "error", err)
				return
			}

			outChan <- order
		}
	}()

	return initialState, nil, nil

}

const maxTimeBetweenMessages = 5 * time.Second

// loadOrdersFromReader loads all orders until either the order created time exceeds the start time or the time taken
// to read a message exceeds maxTimeBetweenMessages, this is taken an indication that the end of the order stream has
// been reached.  This is a workaround for the fact that kafka-go api does not provide a way to determine if the end of
// the stream has been reached.
func loadOrdersFromReader(ctx context.Context, reader orderReader, filter func(order *model.Order) bool) (map[string]*model.Order, error) {
	result := map[string]*model.Order{}
	startTime := time.Now()

	readMsgCount := 0

	for {

		deadline, cancel := context.WithDeadline(ctx, time.Now().Add(maxTimeBetweenMessages))
		msg, err := reader.ReadMessage(deadline)
		cancel()

		if err != nil {
			if !errors.Is(err, context.DeadlineExceeded) {
				return nil, fmt.Errorf("failed to restore order state from store: %w", err)
			} else {
				break
			}
		}

		readMsgCount++

		order := &model.Order{}
		if err = proto.Unmarshal(msg.Value, order); err != nil {
			return nil, fmt.Errorf("failed to unmarshal order whilst recovering state:%w", err)
		}

		if filter(order) {
			result[order.Id] = order
		}

		if msg.Time.After(startTime) {
			break
		}
	}

	slog.Info("order state restored", "ordersLoaded", len(result), "messagesRead", readMsgCount)
	return result, nil
}

func (ks *KafkaStore) Close() {
	if err := ks.writer.Close(); err != nil {
		slog.Error("error when closing kafka writer", "error", err)
	}
}

func (ks *KafkaStore) Write(ctx context.Context, order *model.Order) error {

	order.OwnerId = ks.ownerId
	orderBytes, err := proto.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(order.Id),
		Value: orderBytes,
	}

	if err = ks.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write order to kafka order store: %w", err)
	}

	return nil
}
