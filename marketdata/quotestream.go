package marketdata

import "github.com/ettec/otp-common/model"

// QuoteStream is a standardised interface for quote stream interaction.
type QuoteStream interface {
	// Subscribe to quotes for the given listing id.
	Subscribe(listingId int32) error

	// Chan returns a channel that will receive quotes for the subscribed listings.
	Chan() <-chan *model.ClobQuote

	// Close the quote stream.
	Close()
}
