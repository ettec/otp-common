package marketdata

import (
	"context"
	"fmt"
	"github.com/ettec/otp-common/api/marketdatasource"
	"github.com/ettec/otp-common/model"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"testing"
)

type testConnection struct {
	state        connectivity.State
	getStateChan chan connectivity.State
}

func (t testConnection) GetState() connectivity.State {
	t.state = <-t.getStateChan
	return t.state
}

func (t testConnection) WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool {

	for {
		if t.state != sourceState {
			return true
		}
		t.state = <-t.getStateChan
	}

}

type testClient struct {
	streamOutChan chan marketdatasource.MarketDataSource_ConnectClient
}

func (t testClient) Connect(ctx context.Context, opts ...grpc.CallOption) (marketdatasource.MarketDataSource_ConnectClient, error) {
	return <-t.streamOutChan, nil
}

type testClientStream struct {
	refreshChan    chan *model.ClobQuote
	refreshErrChan chan error
	subscribeChan  chan *marketdatasource.SubscribeRequest
}

func (t testClientStream) Send(request *marketdatasource.SubscribeRequest) error {
	t.subscribeChan <- request
	return nil
}

func (t testClientStream) Recv() (*model.ClobQuote, error) {
	select {
	case r := <-t.refreshChan:
		return r, nil
	case e := <-t.refreshErrChan:
		return nil, e
	}

	return <-t.refreshChan, nil
}

func (t testClientStream) Header() (metadata.MD, error) {
	panic("implement me")
}

func (t testClientStream) Trailer() metadata.MD {
	panic("implement me")
}

func (t testClientStream) CloseSend() error {
	panic("implement me")
}

func (t testClientStream) Context() context.Context {
	panic("implement me")
}

func (t testClientStream) SendMsg(m interface{}) error {
	panic("implement me")
}

func (t testClientStream) RecvMsg(m interface{}) error {
	panic("implement me")
}

func TestMarketDataGatewayClient_refreshesForwaredToOut(t *testing.T) {

	client, stream, conn, mdsQuoteStream := setup(t)

	conn.getStateChan <- connectivity.Ready

	client.streamOutChan <- stream

	stream.refreshChan <- &model.ClobQuote{}

	<-mdsQuoteStream.Chan()
}

func TestMarketDataGatewayClient_sendsEmptyQuoteForAllListingsOnConnectionError(t *testing.T) {

	client, stream, conn, mdc := setup(t)

	conn.getStateChan <- connectivity.Ready

	err := mdc.Subscribe(1)
	assert.NoError(t, err)
	err = mdc.Subscribe(2)
	assert.NoError(t, err)

	client.streamOutChan <- stream

	stream.refreshChan <- &model.ClobQuote{ListingId: 1}
	<-mdc.Chan()

	stream.refreshErrChan <- fmt.Errorf("testerror")
	r := <-mdc.Chan()

	if !r.StreamInterrupted {
		t.FailNow()
	}

	if r.ListingId != 1 {
		t.FailNow()
	}

	if len(r.Bids) != 0 || len(r.Offers) != 0 {
		t.FailNow()
	}

}

func TestMarketDataGatewayClient_testReconnectAfterError(t *testing.T) {

	client, stream, conn, toTest := setup(t)

	conn.getStateChan <- connectivity.Ready

	client.streamOutChan <- stream

	toTest.Subscribe(1)

	stream.refreshChan <- &model.ClobQuote{}
	<-toTest.Chan()

	stream.refreshErrChan <- fmt.Errorf("testerror")
	r := <-toTest.Chan()
	if r.StreamInterrupted != true {
		t.FailNow()
	}

	conn.getStateChan <- connectivity.TransientFailure
	conn.getStateChan <- connectivity.Ready
	client.streamOutChan <- stream

	<-stream.subscribeChan

}

func TestMarketDataGatewayClient_resubscribedOnConnect(t *testing.T) {

	client, stream, conn, toTest := setup(t)

	toTest.Subscribe(1)

	conn.getStateChan <- connectivity.Ready

	client.streamOutChan <- stream

	s := <-stream.subscribeChan

	if s.ListingId != 1 {
		t.FailNow()
	}

}

func setup(t *testing.T) (testClient, testClientStream, testConnection, QuoteStream) {

	client := testClient{

		streamOutChan: make(chan marketdatasource.MarketDataSource_ConnectClient),
	}

	stream := testClientStream{refreshChan: make(chan *model.ClobQuote),
		refreshErrChan: make(chan error),
		subscribeChan:  make(chan *marketdatasource.SubscribeRequest, 10)}

	conn := testConnection{
		getStateChan: make(chan connectivity.State),
	}

	c, err := NewMdsQuoteStreamFromFn(context.Background(), "testId", "testTarget", 0,
		func(targetAddress string) (commonMds, GrpcConnection, error) {

			return &sourceToCommonMds{client: client}, conn, nil
		})

	if err != nil {
		t.FailNow()
	}
	return client, stream, conn, c
}
