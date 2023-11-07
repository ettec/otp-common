package marketdata

import (
	"context"
	"github.com/ettec/otp-common/model"
	"log/slog"
)

type QuoteSource interface {
	Subscribe(listingId int)
}

type subscription struct {
	listingId int32
	out       chan<- *model.ClobQuote
}

type DistributorQuoteStream struct {
	out         chan *model.ClobQuote
	distributor *QuoteDistributor
}

func (q *DistributorQuoteStream) Subscribe(listingId int32) error {
	q.distributor.subscriptionChan <- subscription{
		listingId: listingId,
		out:       q.out,
	}
	return nil
}

func (q *DistributorQuoteStream) Chan() <-chan *model.ClobQuote {
	return q.out
}

func (q *DistributorQuoteStream) Close() {
	q.distributor.removeOutChan <- q.out
}

type subscribeToListing = func(listingId int32) error

// QuoteDistributor fans out quote data sourced from a quote stream to multiple clients.
type QuoteDistributor struct {
	listingToStreams    map[int32][]chan<- *model.ClobQuote
	streamToListings    map[chan<- *model.ClobQuote][]int32
	removeOutChan       chan chan<- *model.ClobQuote
	subscriptionChan    chan subscription
	lastQuote           map[int32]*model.ClobQuote
	subscribedFn        subscribeToListing
	subscribedToListing map[int32]bool
	sendBufferSize      int
}

func NewQuoteDistributor(ctx context.Context, stream QuoteStream, sendBufferSize int) *QuoteDistributor {

	q := &QuoteDistributor{
		listingToStreams:    map[int32][]chan<- *model.ClobQuote{},
		streamToListings:    map[chan<- *model.ClobQuote][]int32{},
		removeOutChan:       make(chan chan<- *model.ClobQuote),
		subscriptionChan:    make(chan subscription),
		lastQuote:           map[int32]*model.ClobQuote{},
		subscribedFn:        stream.Subscribe,
		subscribedToListing: map[int32]bool{},
		sendBufferSize:      sendBufferSize,
	}

	log := slog.Default()
	streamChan := stream.Chan()

	go func() {

		for {
			select {
			case <-ctx.Done():
				return
			case s := <-q.subscriptionChan:

				subscribeToQuotes := q.listingToStreams[s.listingId] == nil
				if subscribeToQuotes {
					if err := q.subscribedFn(s.listingId); err != nil {
						log.Error("failed to subscribe to listing", "listingId", s.listingId, "error", err)
					}

				}

				streamSubscribed := false
				for _, stream := range q.listingToStreams[s.listingId] {
					if stream == s.out {
						streamSubscribed = true
						break
					}
				}

				if !streamSubscribed {
					q.listingToStreams[s.listingId] = append(q.listingToStreams[s.listingId], s.out)
					q.streamToListings[s.out] = append(q.streamToListings[s.out], s.listingId)
					if lastQuote, exists := q.lastQuote[s.listingId]; exists {
						s.out <- lastQuote
					}
				}

			case cq := <-streamChan:
				q.lastQuote[cq.ListingId] = cq

				for _, stream := range q.listingToStreams[cq.ListingId] {
					stream <- cq
				}
			case s := <-q.removeOutChan:
				subscribedListings := q.streamToListings[s]
				for _, listingId := range subscribedListings {
					for idx, o := range q.listingToStreams[listingId] {
						if o == s {
							q.listingToStreams[listingId] = append(q.listingToStreams[listingId][:idx], q.listingToStreams[listingId][idx+1:]...)
							break
						}
					}
				}

				delete(q.streamToListings, s)
				close(s)
			}
		}

	}()

	return q
}

func (q *QuoteDistributor) NewQuoteStream() *DistributorQuoteStream {
	result := &DistributorQuoteStream{make(chan *model.ClobQuote, q.sendBufferSize),
		q}
	return result
}
