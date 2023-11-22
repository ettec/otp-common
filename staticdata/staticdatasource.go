package staticdata

import (
	"context"
	"fmt"
	services "github.com/ettec/otp-common/api/staticdataservice"
	"github.com/ettec/otp-common/bootstrap"
	"github.com/ettec/otp-common/k8s"
	"github.com/ettec/otp-common/model"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
	"log/slog"
	"sync"
	"time"
)

type GrpcConnection interface {
	GetState() connectivity.State
	WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool
}

// Source is a source of static data that fetches data asynchronously from the static data service and has in-built
// retry behaviour
type Source struct {
	cache       *sync.Map
	sdcTaskChan chan *task
}

func NewStaticDataSource(ctx context.Context) (*Source, error) {

	responseBufSize := bootstrap.GetOptionalIntEnvVar("STATICDATASOURCE_RESPONSE_BUFFER_SIZE", 10000)
	taskRetryLimit := bootstrap.GetOptionalIntEnvVar("STATICDATASOURCE_TASK_RETRY_LIMIT", 10)
	grpcMaxReconnectDelay := time.Duration(bootstrap.GetOptionalIntEnvVar("STATICDATASOURCE_GRPC_MAX_RECONNECT_DELAY_SECS",
		120)) * time.Second

	appLabel := "static-data-service"

	targetAddress, err := k8s.GetServiceAddress(appLabel)
	if err != nil {
		return nil, err
	}

	return newStaticDataSource(ctx, func() (client services.StaticDataServiceClient, connection GrpcConnection, err error) {
		slog.Info("creating new static data source", "targetAddress", targetAddress)
		conn, err := grpc.Dial(targetAddress, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(grpcMaxReconnectDelay))
		if err != nil {
			return nil, nil, err
		}

		sdc := services.NewStaticDataServiceClient(conn)

		return sdc, conn, nil
	}, responseBufSize, taskRetryLimit)
}

func newStaticDataSource(ctx context.Context,
	getConnection func() (services.StaticDataServiceClient, GrpcConnection, error), sdsResponseBufSize int,
	taskRetryLimit int) (*Source, error) {
	s := &Source{
		sdcTaskChan: make(chan *task, sdsResponseBufSize),
		cache:       &sync.Map{},
	}

	sdc, conn, err := getConnection()
	if err != nil {
		return nil, err
	}

	go func() {

		retryTasks := map[uuid.UUID]*task{}
		timer := time.NewTicker(1 * time.Second)
		defer timer.Stop()
		for {
			state := conn.GetState()
			if state != connectivity.Ready {
				slog.Info("connecting to static data service....")
				for state != connectivity.Ready {
					conn.WaitForStateChange(context.Background(), state)
					state = conn.GetState()
				}
				slog.Info("static data service connected")
			}

			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				for _, task := range retryTasks {
					if time.Since(task.lastRetry) > time.Duration(task.retry)*time.Second {
						delete(retryTasks, task.id)
						s.sdcTaskChan <- task
					}
				}

			case task := <-s.sdcTaskChan:
				err = task.action(sdc)
				if err != nil {
					if task.retry < taskRetryLimit {
						task.retry = task.retry + 1
						task.lastRetry = time.Now()
						retryTasks[task.id] = task
						slog.Warn("failed to execute static data service task, retry scheduled", "task", task, "error", err)
					} else {
						delete(retryTasks, task.id)
						task.onUnrecoverableError(fmt.Errorf("failed to execute action after %d retries: %w", task.retry, err))
						slog.Error("failed to execute static data service task", "retries", task.retry, "task", task, "error", err)
					}

				}
			}
		}
	}()

	return s, nil
}

type task struct {
	id                   uuid.UUID
	description          string
	retry                int
	lastRetry            time.Time
	action               func(sdc services.StaticDataServiceClient) error
	onUnrecoverableError func(err error)
}

func newTask(description string, action func(sdc services.StaticDataServiceClient) error, onUnrecoverableError func(err error)) *task {
	return &task{
		id:                   uuid.New(),
		description:          description,
		action:               action,
		onUnrecoverableError: onUnrecoverableError,
	}
}

type ListingResult struct {
	Listing *model.Listing
	Err     error
}

func (s *Source) GetListing(ctx context.Context, listingId int32, resultChan chan<- ListingResult) {
	if value, ok := s.cache.Load(listingId); ok {
		listing := value.(*model.Listing)
		go func() {
			resultChan <- ListingResult{listing, nil}
		}()
		return
	}

	s.sdcTaskChan <- newTask(fmt.Sprintf("GetListing for id %d", listingId), func(sdc services.StaticDataServiceClient) error {
		listing, err := sdc.GetListing(ctx, &services.ListingId{
			ListingId: listingId,
		})

		if err != nil {
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.NotFound {
				return fmt.Errorf("error retrieving listing:%w", err)
			} else {
				slog.Warn("no listing found", "listingId", listingId)
			}
		} else {
			s.cache.Store(listing.GetId(), listing)
			resultChan <- ListingResult{listing, nil}
		}

		return nil
	}, func(err error) {
		resultChan <- ListingResult{nil, err}
	})
}

func (s *Source) GetListingMatching(ctx context.Context, matchParams *services.ExactMatchParameters, resultChan chan<- ListingResult) {
	s.sdcTaskChan <- newTask(fmt.Sprintf("GetListing matching %v", matchParams), func(sdc services.StaticDataServiceClient) error {
		listing, err := sdc.GetListingMatching(ctx, matchParams)

		if err != nil {
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.NotFound {
				return fmt.Errorf("error retrieving listing:%w", err)
			} else {
				slog.Warn("no listing found for match params", "params", matchParams)
			}
		} else {
			slog.Info("received listing for match params", "listing", listing, "params", matchParams)
			resultChan <- ListingResult{listing, nil}
		}

		return nil
	}, func(err error) {
		resultChan <- ListingResult{nil, err}
	})
}

type ListingsResult struct {
	Listings []*model.Listing
	Err      error
}

func (s *Source) GetListingsWithSameInstrument(ctx context.Context, listingId int32, resultChan chan<- ListingsResult) {

	s.sdcTaskChan <- newTask(fmt.Sprintf("GetListingsWithSameInstrument for listing id %d", listingId),
		func(sdc services.StaticDataServiceClient) error {

			listings, err := sdc.GetListingsWithSameInstrument(ctx, &services.ListingId{
				ListingId: listingId,
			})

			if err != nil {
				st, ok := status.FromError(err)
				if !ok || st.Code() != codes.NotFound {
					return fmt.Errorf("error retrieving listings :%v", err)
				} else {
					slog.Error("no listings with same instrument", "listingId", listingId)
				}
			} else {
				resultChan <- ListingsResult{listings.Listings, nil}
			}

			return nil
		}, func(err error) {
			resultChan <- ListingsResult{nil, err}
		})

}
