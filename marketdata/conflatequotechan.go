package marketdata

import (
	"errors"
	"github.com/ettec/otp-common/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log/slog"
)

var conflatorQuotesSent = promauto.NewCounter(prometheus.CounterOpts{
	Name: "conflator_quotes_sent",
	Help: "The number of quotes sent across all clients",
})

var conflatorQuotesReceived = promauto.NewCounter(prometheus.CounterOpts{
	Name: "conflator_quotes_received",
	Help: "The number of quotes received from all streams",
})

// conflateQuoteChan conflates the quotes on the inbound channel such that when the client reads
// from the outbound channel only the latest version of a quote for a given listing will be returned.
func conflateQuoteChan(inChan <-chan *model.ClobQuote, capacity int) <-chan *model.ClobQuote {

	pendingQuote := map[int32]*model.ClobQuote{}
	receivedOrder := newBoundedCircularBuffer[int32](capacity)
	log := slog.Default()

	outChan := make(chan *model.ClobQuote)
	go func() {
		defer close(outChan)

		for {
			var quote *model.ClobQuote

			if receivedOrder.len > 0 {
				listingId, _ := receivedOrder.getTail()
				quote = pendingQuote[listingId]
			}

			if quote != nil {
				select {
				case q, ok := <-inChan:
					if !ok {
						return
					}

					if err := conflate(q, pendingQuote, receivedOrder); err != nil {
						log.Error("failed to conflate quote, exiting: %v", err)
						return
					}

					conflatorQuotesReceived.Inc()

				case outChan <- quote:
					delete(pendingQuote, quote.ListingId)
					receivedOrder.removeTail()
					conflatorQuotesSent.Inc()
				}

			} else {
				select {
				case q, ok := <-inChan:
					if !ok {
						return
					}

					if err := conflate(q, pendingQuote, receivedOrder); err != nil {
						log.Error("failed to conflate quote due to error:%v", err)
						return
					}

					conflatorQuotesReceived.Inc()
				}

			}
		}

	}()

	return outChan
}

func conflate(q *model.ClobQuote, pendingQuote map[int32]*model.ClobQuote,
	receivedOrder *boundedCircularBuffer[int32]) error {

	if _, ok := pendingQuote[q.ListingId]; !ok {
		ok = receivedOrder.addHead(q.ListingId)
		if !ok {
			return errors.New("received order buffer is full")
		}
	}
	pendingQuote[q.ListingId] = q
	return nil
}
