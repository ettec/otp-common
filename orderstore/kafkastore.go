// Implementations of stores used to store order updates.
package orderstore

import (
	"context"
	"fmt"
	"github.com/ettec/otp-common/model"
	"github.com/golang/protobuf/proto"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"time"
)


var errLog = log.New(os.Stderr, log.Prefix(), log.Flags())

type KafkaStore struct {
	writer          *kafka.Writer
	topic           string
	kafkaBrokerUrls []string
	ownerId string
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

		topic:           topic,
		kafkaReaderConfig: kafkaReaderConfig,
		ownerId: ownerId,
	}

	result.writer = kafka.NewWriter(kafkaWriterConfig)

	return &result, nil
}

func (ks *KafkaStore) RecoverInitialCache( loadOrder func(order *model.Order) bool) (map[string]*model.Order, error) {

	log.Println("restoring order state")
	reader := ks.getNewReader()
	defer func() {
		err := reader.Close()
		if err != nil {
			log.Printf("error on closing kafka reader:%v", err)
		}
	} ()

	result, err := getInitialState(reader, loadOrder)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ks *KafkaStore) getNewReader() *kafka.Reader {
	reader := kafka.NewReader(ks.kafkaReaderConfig)
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
		defer func() {
			err := reader.Close()
			if err != nil {
				log.Printf("error on closing kafka reader:%v", err)
			}
		} ()
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
	err := ks.writer.Close()
	if err != nil {
		errLog.Printf("error when closing kafka writer:%v", err)
	}
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
