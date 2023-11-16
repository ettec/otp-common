package marketdata

import (
	"github.com/ettec/otp-common/model"
	"testing"
)

func TestQuotesAreConflated(t *testing.T) {

	in := make(chan *model.ClobQuote)

	out := conflateQuoteChan(in, 10)

	in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: 1}
	in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: 2}
	in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: 3}

	in <- &model.ClobQuote{ListingId: 2, XXX_sizecache: 6}
	in <- &model.ClobQuote{ListingId: 2, XXX_sizecache: 7}

	q := <-out
	if q.XXX_sizecache != 3 {
		t.Fatalf("expected last sent quote")
	}

}

func TestQuotesAreConflatedAndReceivedOrderIsMaintained(t *testing.T) {

	in := make(chan *model.ClobQuote)

	out := conflateQuoteChan(in, 10)

	in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: 1}
	in <- &model.ClobQuote{ListingId: 2, XXX_sizecache: 6}
	in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: 2}
	in <- &model.ClobQuote{ListingId: 3, XXX_sizecache: 11}
	in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: 3}
	in <- &model.ClobQuote{ListingId: 2, XXX_sizecache: 7}
	in <- &model.ClobQuote{ListingId: 3, XXX_sizecache: 12}

	q := <-out
	if q.ListingId != 1 || q.XXX_sizecache != 3 {
		t.FailNow()
	}

	q = <-out
	if q.ListingId != 2 || q.XXX_sizecache != 7 {
		t.FailNow()
	}

	q = <-out
	if q.ListingId != 3 || q.XXX_sizecache != 12 {
		t.FailNow()
	}

	in <- &model.ClobQuote{ListingId: 2, XXX_sizecache: 6}
	in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: 1}
	in <- &model.ClobQuote{ListingId: 3, XXX_sizecache: 11}
	in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: 2}
	in <- &model.ClobQuote{ListingId: 1, XXX_sizecache: 3}
	in <- &model.ClobQuote{ListingId: 3, XXX_sizecache: 12}
	in <- &model.ClobQuote{ListingId: 2, XXX_sizecache: 7}

	q = <-out
	if q.ListingId != 2 || q.XXX_sizecache != 7 {
		t.FailNow()
	}

	q = <-out
	if q.ListingId != 1 || q.XXX_sizecache != 3 {
		t.FailNow()
	}

	q = <-out
	if q.ListingId != 3 || q.XXX_sizecache != 12 {
		t.FailNow()
	}

}
