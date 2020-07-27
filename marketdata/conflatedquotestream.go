package marketdata

import (
	"fmt"
	"github.com/ettec/otp-model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"os"
)

var conflatorQuotesSent = promauto.NewCounter(prometheus.CounterOpts{
	Name: "conflator_quotes_sent",
	Help: "The number of quotes sent across all clients",
})

var conflatorQuotesReceived = promauto.NewCounter(prometheus.CounterOpts{
	Name: "conflator_quotes_received",
	Help: "The number of quotes received from all streams",
})

type MdsQuoteStream interface {
	Subscribe(listingId int32)
	GetStream() <-chan *model.ClobQuote
	Close()
}

type conflatedQuoteStream struct {
	stream        MdsQuoteStream
	outChan       chan<- *model.ClobQuote
	closeChan     chan bool
	pendingQuote  map[int32]*model.ClobQuote
	receivedOrder *boundedCircularInt32Buffer
	errLog        *log.Logger
}

func (c *conflatedQuoteStream) Subscribe(listingId int32) {
	c.stream.Subscribe(listingId)
}

func (c *conflatedQuoteStream) Close() {
	c.closeChan <- true
	c.stream.Close()
}

func NewConflatedQuoteStream(stream MdsQuoteStream, out chan<- *model.ClobQuote, capacity int) *conflatedQuoteStream {
	c := &conflatedQuoteStream{
		stream: stream, outChan: out, closeChan: make(chan bool),
		pendingQuote: map[int32]*model.ClobQuote{}, receivedOrder: newBoundedCircularIntBuffer(capacity),
		errLog: log.New(os.Stderr, "", log.Lshortfile|log.Ltime)}

	inChan := stream.GetStream()

	go func() {

		for {
			var eq *model.ClobQuote

			if c.receivedOrder.len > 0 {
				listingId, _ := c.receivedOrder.getTail()
				eq = c.pendingQuote[listingId]
			}

			if eq != nil {
				select {
				case q, ok := <-inChan:
					if !ok {
						c.errLog.Printf("inbound quote channel has closed, exiting")
						return
					}

					if err := c.conflate(q); err != nil {
						c.errLog.Println("exiting:", err)
						return
					}

					conflatorQuotesReceived.Inc()

				case c.outChan <- eq:
					delete(c.pendingQuote, eq.ListingId)
					c.receivedOrder.removeTail()
					conflatorQuotesSent.Inc()

				case <-c.closeChan:
					return
				}

			} else {
				select {
				case q := <-inChan:
					if err := c.conflate(q); err != nil {
						c.errLog.Println("exiting:", err)
						return
					}

					conflatorQuotesReceived.Inc()
				case <-c.closeChan:
					return
				}

			}
		}

	}()

	return c
}

func (c *conflatedQuoteStream) conflate(q *model.ClobQuote) error {

	if _, ok := c.pendingQuote[q.ListingId]; !ok {
		ok = c.receivedOrder.addHead(q.ListingId)
		if !ok {
			return fmt.Errorf("unable to handle inbound quote as quote received order buffer size exceeded")
		}
	}
	c.pendingQuote[q.ListingId] = q
	return nil
}

type boundedCircularInt32Buffer struct {
	buffer   []int32
	capacity int
	len      int
	readPtr  int
	writePtr int
}

func newBoundedCircularIntBuffer(capacity int) *boundedCircularInt32Buffer {
	b := &boundedCircularInt32Buffer{buffer: make([]int32, capacity, capacity), capacity: capacity}

	return b
}

// true if the buffer is not full and the value is added
func (b *boundedCircularInt32Buffer) addHead(i int32) bool {

	if b.len == b.capacity {
		return false
	}

	b.buffer[b.writePtr] = i
	b.len++

	if b.writePtr == b.capacity-1 {
		b.writePtr = 0
	} else {
		b.writePtr++
	}

	return true

}

func (b *boundedCircularInt32Buffer) getTail() (int32, bool) {
	if b.len == 0 {
		return 0, false
	}

	res := b.buffer[b.readPtr]
	return res, true
}

// returns the value and true if a value is available
func (b *boundedCircularInt32Buffer) removeTail() (int32, bool) {
	if b.len == 0 {
		return 0, false
	}

	res := b.buffer[b.readPtr]
	b.len--
	b.readPtr++
	if b.readPtr == b.capacity {
		b.readPtr = 0
	}

	return res, true

}
