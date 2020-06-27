package orderstore

import (
	"context"
	"fmt"
	"github.com/ettec/otp-model"
	"github.com/golang/protobuf/proto"
	"github.com/segmentio/kafka-go"
	logger "log"
	"os"
	"time"
)

var log = logger.New(os.Stdout, "", logger.Ltime|logger.Lshortfile)
var errLog = logger.New(os.Stderr, "", logger.Ltime|logger.Lshortfile)

type KafkaStore struct {
	writer          *kafka.Writer
	topic           string
	kafkaBrokerUrls []string
	ownerId         string
}

type orderReader interface {
	Close() error
	ReadMessage(ctx context.Context) (kafka.Message, error)
}

func NewKafkaStore(kafkaBrokerUrls []string, ownerId string) (*KafkaStore, error) {

	topic := "orders"

	result := KafkaStore{

		topic:           topic,
		kafkaBrokerUrls: kafkaBrokerUrls,
		ownerId:         ownerId,
	}

	result.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      kafkaBrokerUrls,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		Async:        true,
		BatchTimeout: 10 * time.Millisecond,
	})

	return &result, nil
}

func (ks *KafkaStore) RecoverInitialCache() (map[string]*model.Order, error) {

	log.Println("restoring order state")
	reader := ks.getNewReader()
	defer reader.Close()

	owns := func(order *model.Order) bool {
		return ks.ownerId == order.GetOwnerId()
	}

	result, err := getInitialState(reader, owns)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ks *KafkaStore) getNewReader() *kafka.Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        ks.kafkaBrokerUrls,
		Topic:          ks.topic,
		ReadBackoffMin: 100 * time.Millisecond,
		ReadBackoffMax: 200 * time.Millisecond,
		MaxWait:        150 * time.Millisecond,
	})
	return reader
}

func (ks *KafkaStore) SubscribeToAllOrders(updatesChan chan<- *model.Order, after time.Time) (map[string]*model.Order, error) {
	reader := ks.getNewReader()

	afterTimestamp := &model.Timestamp{Seconds: after.Unix()}
	timeFilter := func(order *model.Order) bool {
		return order.GetCreated().After(afterTimestamp)
	}

	initialState, err := getInitialState(reader, timeFilter)

	if err != nil {
		return nil, err
	}

	go func() {
		defer reader.Close()
		for {
			msg, err := reader.ReadMessage(context.Background())

			if err != nil {
				errLog.Print("failed to read message:", err)
				break
			}

			order := &model.Order{}
			err = proto.Unmarshal(msg.Value, order)
			if err != nil {
				errLog.Print("failed to unmarshal order:", err)
				break
			}

			updatesChan <- order
		}
	}()

	return initialState, nil

}

func getInitialState(reader orderReader, filter func(order *model.Order) bool) (map[string]*model.Order, error) {
	result := map[string]*model.Order{}
	now := time.Now()

	readMsgCnt := 0

	for {

		deadline, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
		msg, err := reader.ReadMessage(deadline)
		cancel()

		if err != nil {
			if err != context.DeadlineExceeded {
				return nil, fmt.Errorf("failed to restore order state from store: %w", err)
			} else {
				break
			}
		}

		readMsgCnt++

		order := &model.Order{}
		err = proto.Unmarshal(msg.Value, order)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal order whilst recovering state:%w", err)
		}

		if filter(order) {
			result[order.Id] = order
		}

		if msg.Time.After(now) {
			break
		}
	}

	log.Printf("order state restored, %v orders reloaded from %v messages", len(result), readMsgCnt)
	return result, nil
}

func (ks *KafkaStore) Close() {
	ks.writer.Close()
}

func (ks *KafkaStore) Write(order *model.Order) error {

	order.OwnerId = ks.ownerId
	orderBytes, err := proto.Marshal(order)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(order.Id),
		Value: orderBytes,
	}

	err = ks.writer.WriteMessages(context.Background(), msg)

	if err != nil {
		return fmt.Errorf("failed to write order to kafka order store: %w", err)
	}

	return nil
}
