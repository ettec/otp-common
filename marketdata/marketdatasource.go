package marketdata

import (
	"fmt"
	"github.com/ettec/otp-common/api/marketdatasource"
	"github.com/ettec/otp-common/bootstrap"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc/metadata"
	"log/slog"
)

var connections = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "active_connections",
	Help: "The number of active connections",
})

var quotesSent = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quotes_sent",
	Help: "The number of quotes sent across all clients",
})

type marketDataSourceServer struct {
	quoteDistributor quoteDistributor
	maxSubscriptions int
}

type quoteDistributor interface {
	NewQuoteStream() *DistributorQuoteStream
}

func NewMarketDataSource(quoteDistributor quoteDistributor) marketdatasource.MarketDataSourceServer {

	maxSubscriptions := bootstrap.GetOptionalIntEnvVar("MARKETDATASOURCE_MAX_SUBSCRIPTIONS", 10000)

	return &marketDataSourceServer{quoteDistributor, maxSubscriptions}
}

const SubscriberIdKey = "subscriber_id"

func (s *marketDataSourceServer) Connect(stream marketdatasource.MarketDataSource_ConnectServer) error {

	metaData, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return fmt.Errorf("failed to get metadata from incoming context")
	}

	values := metaData.Get(SubscriberIdKey)
	if len(values) != 1 {
		return fmt.Errorf("meta data does not contain an entry for required subscriber id key %v", SubscriberIdKey)
	}

	log := slog.Default().With("subscriberId", values[0])

	fromClientId := values[0]
	subscriberId := fromClientId + ":" + uuid.New().String()

	log.Info("connect request received", "subscriber", fromClientId, "uniqueConnectionId", subscriberId)

	quoteStream := NewConflatedQuoteStream(subscriberId, s.quoteDistributor.NewQuoteStream(),
		s.maxSubscriptions)
	defer quoteStream.Close()

	go func() {
		for {
			subscription, err := stream.Recv()
			if err != nil {
				log.Error("error receiving from grpc stream", "error", err)
				break
			} else {
				log.Info("subscribing to listing id", "subscriberId", subscriberId,
					"listingId", subscription.ListingId)
				err := quoteStream.Subscribe(subscription.ListingId)
				if err != nil {
					log.Error("error subscribing to listing", "listingId", subscription.ListingId, "error", err)
				}
			}
		}
	}()

	connections.Inc()

	for quote := range quoteStream.Chan() {
		if err := stream.Send(quote); err != nil {
			log.Error("failed to send quote, closing connection", "subscriberId", subscriberId, "error", err)
			break
		}

		quotesSent.Inc()
	}

	connections.Dec()

	return nil
}
