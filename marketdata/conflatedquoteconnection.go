// Package marketdata contains code to assist in building services that handle market data.
package marketdata

import (
	"fmt"
	"github.com/ettec/otp-common/model"
	"log/slog"
)

// ConflatedQuoteStream conflates quotes from a quote stream such that the most recent quote for a listing is read from
// the stream even if the client is reading quotes at a slower rate than they are being published.
type ConflatedQuoteStream struct {
	id                 string
	maxSubscriptions   int
	quotesIn           QuoteStream
	conflatedQuoteChan <-chan *model.ClobQuote
	subscriptions      map[int32]bool
	log                *slog.Logger
}

func NewConflatedQuoteStream(id string, stream QuoteStream,
	maxSubscriptions int) *ConflatedQuoteStream {

	conflatedQuoteChan := conflateQuoteChan(stream.Chan(), maxSubscriptions)
	logger := slog.Default().With("conflatedQuoteConnectionId", id)

	c := &ConflatedQuoteStream{id: id,
		maxSubscriptions:   maxSubscriptions,
		subscriptions:      map[int32]bool{},
		quotesIn:           stream,
		conflatedQuoteChan: conflatedQuoteChan,
		log:                logger,
	}

	return c
}

func (c *ConflatedQuoteStream) Subscribe(listingId int32) error {

	if len(c.subscriptions) == c.maxSubscriptions {
		return fmt.Errorf("max number of subscriptions, %v, for this connection has been reached", c.maxSubscriptions)
	}

	if c.subscriptions[listingId] {
		return nil
	}

	err := c.quotesIn.Subscribe(listingId)
	if err != nil {
		return fmt.Errorf("failed to subscribe to listing %v: %w", listingId, err)
	}

	c.log.Info("Subscribed to listing", "listingId", listingId)

	return nil
}

func (c *ConflatedQuoteStream) Chan() <-chan *model.ClobQuote {
	return c.conflatedQuoteChan
}

func (c *ConflatedQuoteStream) Close() {
	c.quotesIn.Close()
}
