package marketdata

import (
	"github.com/ettec/otp-common/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testMdsQuoteStream struct {
	subscribe func(listingId int32)
	stream    chan *model.ClobQuote
}

func (t testMdsQuoteStream) Close() {
	panic("implement me")
}

func (t testMdsQuoteStream) Subscribe(listingId int32) error {
	t.subscribe(listingId)
	return nil
}

func (t testMdsQuoteStream) Chan() <-chan *model.ClobQuote {
	return t.stream
}

func TestClientConnectionSubscribe(t *testing.T) {

	in := make(chan *model.ClobQuote, 100)

	c := NewConflatedQuoteStream("8testId", &testMdsQuoteStream{
		func(listingId int32) {

		}, in}, 100)

	err := c.Subscribe(1)
	assert.NoError(t, err)
	err = c.Subscribe(2)
	assert.NoError(t, err)

	in <- &model.ClobQuote{ListingId: 1}
	in <- &model.ClobQuote{ListingId: 2}

	if q := <-c.Chan(); q.ListingId != 1 {
		t.Errorf("expected quote with listing id 1")
	}
	if q := <-c.Chan(); q.ListingId != 2 {
		t.Errorf("expected quote with listing id 2")
	}

	select {
	case <-c.Chan():
		t.Errorf("no more quotes expected")
	default:
	}

}

func TestSlowConnectionDoesNotBlockDownstreamSender(t *testing.T) {

	in := make(chan *model.ClobQuote)

	c := NewConflatedQuoteStream("testId",
		&testMdsQuoteStream{
			func(listingId int32) {
			}, in}, 100)

	err := c.Subscribe(1)
	assert.NoError(t, err)
	err = c.Subscribe(2)
	assert.NoError(t, err)

	for i := 0; i < 2000; i++ {
		in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: int32(i)}
		in <- &model.ClobQuote{ListingId: 2, XXX_sizecache: int32(i)}
	}

	if q := <-c.Chan(); q.ListingId != 1 && q.XXX_sizecache != 1999 {
		t.Errorf("expected quote with listing id 1")
	}
	if q := <-c.Chan(); q.ListingId != 2 && q.XXX_sizecache != 1999 {
		t.Errorf("expected quote with listing id 2")
	}

}
