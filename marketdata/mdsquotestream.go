package marketdata

import (
	"context"
	"fmt"
	api "github.com/ettec/otp-common/api/marketdataservice"
	"github.com/ettec/otp-common/api/marketdatasource"
	"github.com/ettec/otp-common/bootstrap"
	"github.com/ettec/otp-common/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
)

var quotesReceived = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quotes_received",
	Help: "The number of quotes received from all streams",
})

type GetMdsClientFn = func(targetAddress string) (commonMds, GrpcConnection, error)

type GrpcConnection interface {
	GetState() connectivity.State
	WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool
}

type commonMds interface {
	Connect(ctx context.Context, opts ...grpc.CallOption) (mdsClient, error)
}

type mdsClient interface {
	Subscribe(int32) error
	Recv() (*model.ClobQuote, error)
}

type serviceToCommonMdsClient struct {
	connectClient api.MarketDataService_ConnectClient
	client        api.MarketDataServiceClient
	subscriberId  string
}

func (s *serviceToCommonMdsClient) Subscribe(listingId int32) error {
	_, err := s.client.Subscribe(context.Background(), &api.MdsSubscribeRequest{
		SubscriberId: s.subscriberId,
		ListingId:    listingId,
	})

	return err
}
func (s *serviceToCommonMdsClient) Recv() (*model.ClobQuote, error) {
	return s.connectClient.Recv()
}

type serviceToCommonMds struct {
	subscriberId string
	client       api.MarketDataServiceClient
}

func (s *serviceToCommonMds) Connect(ctx context.Context, opts ...grpc.CallOption) (mdsClient, error) {
	cc, err := s.client.Connect(ctx, &api.MdsConnectRequest{SubscriberId: s.subscriberId}, opts...)
	if err != nil {
		return nil, err
	}

	return &serviceToCommonMdsClient{
		connectClient: cc,
		client:        s.client,
		subscriberId:  s.subscriberId,
	}, nil
}

// MdsQuoteStream wraps the API of a market data service such that it conforms to the QuoteStream interface.
type MdsQuoteStream struct {
	conn             *grpc.ClientConn
	subscriptionLock sync.Mutex
	client           mdsClient
	subscriptions    map[int32]bool
	out              chan *model.ClobQuote
	cancelContext    func()
	log              *slog.Logger
}

// NewQuoteStreamFromMarketDataService returns a quote stream that sources quote data from a market data service.
func NewQuoteStreamFromMarketDataService(ctx context.Context, subscriberId string, targetAddress string, maxReconnectInterval time.Duration,
	quoteBufferSize int) (QuoteStream, error) {

	mdcFn := func(targetAddress string) (commonMds, GrpcConnection, error) {
		conn, err := grpc.Dial(targetAddress, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(maxReconnectInterval))
		if err != nil {
			return nil, nil, err
		}

		client := api.NewMarketDataServiceClient(conn)
		return &serviceToCommonMds{
			subscriberId: subscriberId,
			client:       client,
		}, conn, nil
	}

	return NewMdsQuoteStreamFromFn(ctx, subscriberId, targetAddress, quoteBufferSize, mdcFn)
}

type sourceToCommonMdsClient struct {
	cc marketdatasource.MarketDataSource_ConnectClient
}

func (s *sourceToCommonMdsClient) Subscribe(listingId int32) error {
	err := s.cc.Send(&marketdatasource.SubscribeRequest{
		ListingId: listingId,
	})

	return err
}
func (s *sourceToCommonMdsClient) Recv() (*model.ClobQuote, error) {
	return s.cc.Recv()
}

type sourceToCommonMds struct {
	client marketdatasource.MarketDataSourceClient
}

func (s *sourceToCommonMds) Connect(ctx context.Context, opts ...grpc.CallOption) (mdsClient, error) {
	cc, err := s.client.Connect(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &sourceToCommonMdsClient{
		cc: cc,
	}, nil
}

func NewQuoteStreamFromMdSource(ctx context.Context, id string, targetAddress string, maxReconnectInterval time.Duration,
	quoteBufferSize int) (QuoteStream, error) {

	mdcFn := func(targetAddress string) (commonMds, GrpcConnection, error) {
		conn, err := grpc.Dial(targetAddress, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(maxReconnectInterval))
		if err != nil {
			return nil, nil, err
		}

		client := marketdatasource.NewMarketDataSourceClient(conn)
		return &sourceToCommonMds{client}, conn, nil
	}

	return NewMdsQuoteStreamFromFn(ctx, id, targetAddress, quoteBufferSize, mdcFn)
}

func NewMdsQuoteStreamFromFn(parentCtx context.Context, id string, targetAddress string, outBufferSize int,
	getConnection GetMdsClientFn) (QuoteStream, error) {

	maxReconnectWaitTime := time.Duration(bootstrap.GetOptionalIntEnvVar("MDSQUOTESTREAM_MAX_RECONNECT_WAIT_SECS", 30)) * time.Second

	ctx, cancel := context.WithCancel(parentCtx)
	out := make(chan *model.ClobQuote, outBufferSize)

	log := slog.With("target", targetAddress)

	mqs := &MdsQuoteStream{
		out:           out,
		subscriptions: map[int32]bool{},
		cancelContext: cancel,
		log:           log,
	}

	log.Info("connecting to market data source", "targetAddress", targetAddress)

	client, conn, err := getConnection(targetAddress)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(out)

		sentQuoteListingIds := map[int32]bool{}
		for {
			state := conn.GetState()
			retryWait := time.Duration(1) * time.Second
			for state != connectivity.Ready {
				log.Info(fmt.Sprintf("waiting %v before checking market data source connection state", retryWait))

				select {
				case <-ctx.Done():
				case <-time.After(retryWait):
				}

				retryWait = retryWait * 2

				if retryWait > maxReconnectWaitTime {
					retryWait = maxReconnectWaitTime
				}

				conn.WaitForStateChange(ctx, state)

				state = conn.GetState()
				log.Info("market data source connection state", "state", state)
			}

			mdsClient, err := client.Connect(metadata.AppendToOutgoingContext(ctx, "subscriber_id", id))
			if err != nil {
				mqs.log.Error("failed to connect to quote stream", "error", err)
				continue
			}

			log.Info("connected to quote stream, resubscribing to all listings")

			mqs.setMdsClient(mdsClient)

			for {
				clobQuote, err := mdsClient.Recv()
				if err != nil {
					log.Error("inbound stream error", "error", err)
					break
				}
				out <- clobQuote
				sentQuoteListingIds[clobQuote.ListingId] = true
				quotesReceived.Inc()
			}

			log.Info("inbound stream closed, resetting all quotes")
			for listingId := range sentQuoteListingIds {
				out <- &model.ClobQuote{
					ListingId:         listingId,
					Bids:              []*model.ClobLine{},
					Offers:            []*model.ClobLine{},
					StreamInterrupted: true,
					StreamStatusMsg:   "market data source client stream interrupted",
				}
			}
			sentQuoteListingIds = map[int32]bool{}

		}
	}()

	return mqs, nil
}

func (m *MdsQuoteStream) setMdsClient(mdsClient mdsClient) {
	m.subscriptionLock.Lock()
	defer m.subscriptionLock.Unlock()
	m.client = mdsClient
	for listingId := range m.subscriptions {
		err := mdsClient.Subscribe(listingId)
		if err != nil {
			m.log.Error("failed to subscribe to quote for listing", "listingId", listingId, "error", err)
			break
		}
	}
}

func (m *MdsQuoteStream) Chan() <-chan *model.ClobQuote {
	return m.out
}

func (m *MdsQuoteStream) Subscribe(listingId int32) error {
	m.subscriptionLock.Lock()
	defer m.subscriptionLock.Unlock()
	m.subscriptions[listingId] = true
	if m.client != nil {
		err := m.client.Subscribe(listingId)
		if err != nil {
			return fmt.Errorf("failed to subscribe to quote for listing %v: %w", listingId, err)

		}
	}

	return nil
}

func (m *MdsQuoteStream) Close() {
	m.cancelContext()
}
