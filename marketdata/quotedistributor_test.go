package marketdata

import (
	"context"
	"github.com/ettec/otp-common/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_quoteDistributor_Send(t *testing.T) {

	in := make(chan *model.ClobQuote, 10)

	d := NewQuoteDistributor(context.Background(), testMdsQuoteStream{
		func(listingId int32) {
		}, in}, 100)

	s1 := d.NewQuoteStream()

	s2 := d.NewQuoteStream()

	err := s1.Subscribe(1)
	assert.NoError(t, err)
	err = s2.Subscribe(1)
	assert.NoError(t, err)

	in <- &model.ClobQuote{ListingId: 1}

	if q := <-s1.Chan(); q.ListingId != 1 {
		t.Errorf("expected quote not received")
	}

	if q := <-s2.Chan(); q.ListingId != 1 {
		t.Errorf("expected quote not received")
	}

}

func Test_subscriptionReceivesLastSentQuote(t *testing.T) {

	in := make(chan *model.ClobQuote, 10)

	d := NewQuoteDistributor(context.Background(), testMdsQuoteStream{
		func(listingId int32) {
		}, in}, 100)

	s1 := d.NewQuoteStream()
	s2 := d.NewQuoteStream()

	err := s1.Subscribe(1)
	assert.NoError(t, err)

	in <- &model.ClobQuote{ListingId: 1}

	if q := <-s1.Chan(); q.ListingId != 1 {
		t.Errorf("expected quote note received")
	}

	err = s2.Subscribe(1)
	assert.NoError(t, err)

	if q := <-s2.Chan(); q.ListingId != 1 {
		t.Errorf("expected quote note received")
	}

}

func Test_subscribeCalledOnceForAGivenListing(t *testing.T) {

	in := make(chan *model.ClobQuote)

	subscribeCalls := make(chan int32, 10)
	d := NewQuoteDistributor(context.Background(), testMdsQuoteStream{
		func(listingId int32) {
			subscribeCalls <- listingId
		}, in}, 100)

	s1 := d.NewQuoteStream()
	d.NewQuoteStream()

	err := s1.Subscribe(1)
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	if len(subscribeCalls) != 1 {
		t.FailNow()
	}

}

func Test_subscribeOnlyCalledOnceForAGivenListing(t *testing.T) {

	in := make(chan *model.ClobQuote)

	subscribeCalls := make(chan int32, 10)
	d := NewQuoteDistributor(context.Background(), testMdsQuoteStream{
		func(listingId int32) {
			subscribeCalls <- listingId
		}, in}, 100)

	s1 := d.NewQuoteStream()
	s2 := d.NewQuoteStream()

	err := s1.Subscribe(1)
	assert.NoError(t, err)
	err = s2.Subscribe(1)
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	if len(subscribeCalls) != 1 {
		t.FailNow()
	}

}

func Test_onlySubscribedQuotesReceived(t *testing.T) {

	in := make(chan *model.ClobQuote)

	d := NewQuoteDistributor(context.Background(), testMdsQuoteStream{
		func(listingId int32) {
		}, in}, 100)

	s1 := d.NewQuoteStream()

	err := s1.Subscribe(1)
	assert.NoError(t, err)
	err = s1.Subscribe(2)
	assert.NoError(t, err)

	in <- &model.ClobQuote{ListingId: 3}
	in <- &model.ClobQuote{ListingId: 1}
	in <- &model.ClobQuote{ListingId: 4}
	in <- &model.ClobQuote{ListingId: 2}

	q := <-s1.Chan()
	if q.ListingId != 1 {
		t.Errorf("unexpected quote")
	}

	q = <-s1.Chan()
	if q.ListingId != 2 {
		t.Errorf("unexpected quote")
	}

}
