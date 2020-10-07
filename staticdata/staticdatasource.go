package staticdata

import (
	"context"
	"fmt"
	services "github.com/ettec/otp-common/api/staticdataservice"
	"github.com/ettec/otp-common/bootstrap"
	"github.com/ettec/otp-common/k8s"
	"github.com/ettec/otp-common/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"sync"
	"time"
)

type SubscriptionHandler interface {
	Subscribe(listingId int32)
}

type SubscriptionClient interface {
	Subscribe(symbol string)
}

type GetListingFn = func(listingId int32, onSymbol chan<- *model.Listing)

type ListingSource interface {
	GetListing(listingId int32, result chan<- *model.Listing)
}

type listingSource struct {
	cache *sync.Map
	sdcTaskChan chan staticDataServiceTask
	errLog      *log.Logger
}

type GrpcConnection interface {
	GetState() connectivity.State
	WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool
}

type GetStaticDataServiceClientFn = func() (services.StaticDataServiceClient, GrpcConnection, error)


func NewStaticDataSource(external bool) (*listingSource, error) {

	responseBufSize := bootstrap.GetOptionalIntEnvVar("STATICDATASOURCE_RESPONSE_BUFFER_SIZE", 10000)
	grpcMaxReconnectDelay := time.Duration(bootstrap.GetOptionalIntEnvVar("STATICDATASOURCE_GRPC_MAX_RECONNECT_DELAY_SECS",
		120))*time.Second

	appLabel := "static-data-service"

	targetAddress, err := k8s.GetServiceAddress(appLabel)
	if err != nil {
		return nil, err
	}

	return newStaticDataSource(func() (client services.StaticDataServiceClient, connection GrpcConnection, err error) {
		log.Println("static data service address:" + targetAddress)
		conn, err := grpc.Dial(targetAddress, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(grpcMaxReconnectDelay))
		if err != nil {
			return nil, nil, err
		}

		sdc := services.NewStaticDataServiceClient(conn)

		return sdc, conn, nil
	}, responseBufSize)
}

func newStaticDataSource(getConnection GetStaticDataServiceClientFn, sdsResponseBufSize int) (*listingSource, error) {
	s := &listingSource{
		sdcTaskChan: make(chan staticDataServiceTask, sdsResponseBufSize),
		errLog:      log.New(os.Stdout, log.Prefix(), log.Flags()),
		cache: &sync.Map{},
	}

	sdc, conn, err := getConnection()
	if err != nil {
		return nil, err
	}

	go func() {

		for {
			state := conn.GetState()
			if state != connectivity.Ready {
				log.Printf("connecting to static data service....")
				for state != connectivity.Ready {
					conn.WaitForStateChange(context.Background(), state)
					state = conn.GetState()
				}
				log.Printf("static data service connected")
			}

			select {
			case t := <-s.sdcTaskChan:
				err := t(sdc)
				if err != nil {
					s.sdcTaskChan <- t
					s.errLog.Printf("error executing static data service task, retry schduled.  Error:%v", err)
				}
			}

		}
	}()

	return s, nil
}

type staticDataServiceTask func(sdc services.StaticDataServiceClient) error

func (s *listingSource) GetListing(listingId int32, resultChan chan<- *model.Listing) {
	if value, ok := s.cache.Load(listingId); ok {
		listing := value.(*model.Listing)
		go func() {
			resultChan <-listing
		}()
	}

	s.sdcTaskChan <- func(sdc services.StaticDataServiceClient) error {
		listing, err := sdc.GetListing(context.Background(), &services.ListingId{
			ListingId: listingId,
		})

		if err != nil {
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.NotFound {
				return fmt.Errorf("error retrieving listing:%v", err)
			} else {
				s.errLog.Printf("no listing found for id:%v", listingId)
			}
		} else {
			s.cache.Store(listing.GetId(), listing)
			resultChan <- listing
		}

		return nil
	}
}

func (s *listingSource) GetListingMatching(matchParams *services.ExactMatchParameters, result chan<- *model.Listing) {
	s.sdcTaskChan <- func(sdc services.StaticDataServiceClient) error {
		listing, err := sdc.GetListingMatching(context.Background(), matchParams)

		if err != nil {
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.NotFound {
				return fmt.Errorf("error retrieving listing:%v", err)
			} else {
				s.errLog.Printf("no listing found for match params:%v", matchParams)
			}
		} else {
			log.Printf("received listing:%v for symbol matching:%v and mic:%v", listing, matchParams.Symbol,
				matchParams.Mic)
			result <- listing
		}

		return nil
	}
}

func (s *listingSource) GetListingsWithSameInstrument(listingId int32, listingGroupsIn chan<- []*model.Listing) {

	s.sdcTaskChan <- func(sdc services.StaticDataServiceClient) error {

		listings, err := sdc.GetListingsWithSameInstrument(context.Background(), &services.ListingId{
			ListingId: listingId,
		})

		if err != nil {
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.NotFound {
				return fmt.Errorf("error retrieving listings :%v", err)
			} else {
				s.errLog.Printf("no listings found for same instrument, listing id:%v", listingId)
			}
		} else {
			listingGroupsIn <- listings.Listings
		}

		return nil
	}

}
