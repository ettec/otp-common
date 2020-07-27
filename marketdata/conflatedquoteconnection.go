package marketdata

import (
	"fmt"
	"github.com/ettec/otp-model"
	"log"
	"os"
)

type ConflatedQuoteConnection interface {
	GetId() string
	Subscribe(listingId int32) error
	Close()
}

type conflatedQuoteConnection struct {
	id               string
	maxSubscriptions int
	stream           Stream
	subscriptions    map[int32]bool
	log              *log.Logger
	errLog           *log.Logger
}

func (c *conflatedQuoteConnection) Close() {
	c.stream.Close()
}

func (c *conflatedQuoteConnection) Subscribe(listingId int32) error {

	if len(c.subscriptions) == c.maxSubscriptions {
		return fmt.Errorf("max number of subscriptions for this connection has been reached: %v", c.maxSubscriptions)
	}

	if c.subscriptions[listingId] {
		return fmt.Errorf("already subscribed to listing id: %v", listingId)
	}

	c.stream.Subscribe(listingId)
	c.log.Println("subscribed to listing id:", listingId)

	return nil
}

func (c *conflatedQuoteConnection) GetId() string {
	return c.id
}

func NewConflatedQuoteConnection(id string, stream MdsQuoteStream, out chan<- *model.ClobQuote,
	maxSubscriptions int) *conflatedQuoteConnection {

	conflatedStream := NewConflatedQuoteStream(stream, out, maxSubscriptions)

	c := &conflatedQuoteConnection{id: id,
		maxSubscriptions: maxSubscriptions,
		subscriptions:    map[int32]bool{},
		stream:           conflatedStream,
		log:              log.New(os.Stdout, "conflatedQuoteConnection:"+id, log.LstdFlags),
		errLog:           log.New(os.Stderr, "conflatedQuoteConnection:"+id, log.LstdFlags)}

	return c
}

type Stream interface {
	Subscribe(listingId int32)
	Close()
}
