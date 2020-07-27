package marketdata

import (
	"github.com/ettec/otp-model"
	"testing"
)

type testMdsQuoteStream struct {
	subscribe func(listingId int32)
	stream    chan *model.ClobQuote
}

func (t testMdsQuoteStream) Close() {
	panic("implement me")
}

func (t testMdsQuoteStream) Subscribe(listingId int32) {
	t.subscribe(listingId)
}

func (t testMdsQuoteStream) GetStream() <-chan *model.ClobQuote {
	return t.stream
}

func Test_clientConnection_Subscribe(t *testing.T) {

	in := make(chan *model.ClobQuote, 100)
	out := make(chan *model.ClobQuote, 100)

	c := NewConflatedQuoteConnection("8testId", &testMdsQuoteStream{
		func(listingId int32) {

		}, in}, out, 100)

	c.Subscribe(1)
	c.Subscribe(2)

	in <- &model.ClobQuote{ListingId: 1}
	in <- &model.ClobQuote{ListingId: 2}

	if q := <-out; q.ListingId != 1 {
		t.Errorf("expected quote with listing id 1")
	}
	if q := <-out; q.ListingId != 2 {
		t.Errorf("expected quote with listing id 2")
	}

	select {
	case <-out:
		t.Errorf("no more quotes expected")
	default:
	}

}

func Test_slowConnectionDoesNotBlockDownstreamSender(t *testing.T) {

	in := make(chan *model.ClobQuote)
	out := make(chan *model.ClobQuote, 100)

	c := NewConflatedQuoteConnection("testId",
		&testMdsQuoteStream{
			func(listingId int32) {
			}, in}, out, 100)

	c.Subscribe(1)
	c.Subscribe(2)

	for i := 0; i < 2000; i++ {
		in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: int32(i)}
		in <- &model.ClobQuote{ListingId: 2, XXX_sizecache: int32(i)}
	}

	if q := <-out; q.ListingId != 1 && q.XXX_sizecache != 1999 {
		t.Errorf("expected quote with listing id 1")
	}
	if q := <-out; q.ListingId != 2 && q.XXX_sizecache != 1999 {
		t.Errorf("expected quote with listing id 2")
	}

}
