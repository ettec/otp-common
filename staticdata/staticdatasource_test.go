package staticdata

import (
	"context"
	"errors"
	services "github.com/ettec/otp-common/api/staticdataservice"
	"github.com/ettec/otp-common/model"
	"github.com/ettec/otp-common/staticdata/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"testing"
	"time"
)

//go:generate go run github.com/golang/mock/mockgen -destination mocks/staticdataservice.go -package mocks github.com/ettec/otp-common/api/staticdataservice StaticDataServiceClient

func TestGetListing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listings := map[int32]*model.Listing{}
	listing1 := &model.Listing{Id: 1}
	listings[listing1.Id] = listing1

	sds, _ := newStaticDataSource(ctx, func() (services.StaticDataServiceClient, GrpcConnection, error) {
		sdsClient := &sdsClient{listings: listings}
		conn := &grpcConn{}

		return sdsClient, conn, nil
	}, 1000, 10)

	resultChan := make(chan ListingResult, 100)

	sds.GetListing(ctx, 1, resultChan)

	result1 := <-resultChan
	assert.Equal(t, int32(1), result1.Listing.Id)

	time.Sleep(1 * time.Second)

	sds.GetListing(ctx, 1, resultChan)

	time.Sleep(1 * time.Second)

	result2 := <-resultChan
	assert.Equal(t, int32(1), result2.Listing.Id)

	assert.GreaterOrEqual(t, 1, len(resultChan))
}

func TestGetListingsWithSameInstrument(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	listing1 := &model.Listing{Id: 1}
	listing2 := &model.Listing{Id: 2}

	listings := &services.Listings{Listings: []*model.Listing{listing1, listing2}}

	sdsClient := mocks.NewMockStaticDataServiceClient(mockCtrl)
	sdsClient.EXPECT().GetListingsWithSameInstrument(gomock.Any(), gomock.Any()).Return(listings, nil)

	sds, _ := newStaticDataSource(ctx, func() (services.StaticDataServiceClient, GrpcConnection, error) {
		conn := &grpcConn{}

		return sdsClient, conn, nil
	}, 1000, 10)

	resultChan := make(chan ListingsResult, 100)

	sds.GetListingsWithSameInstrument(ctx, 1, resultChan)

	result1 := <-resultChan
	assert.Equal(t, int32(1), result1.Listings[0].Id)
	assert.Equal(t, int32(2), result1.Listings[1].Id)

}

func TestRetryBehaviourWhenEventuallySucceeds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sdsClient := mocks.NewMockStaticDataServiceClient(mockCtrl)
	sdsClient.EXPECT().GetListing(gomock.Any(), gomock.Any()).Return(nil, errors.New("not ready"))
	sdsClient.EXPECT().GetListing(gomock.Any(), gomock.Any()).Return(nil, errors.New("not ready"))
	sdsClient.EXPECT().GetListing(gomock.Any(), gomock.Any()).Return(&model.Listing{Id: 1}, nil)

	sds, _ := newStaticDataSource(ctx, func() (services.StaticDataServiceClient, GrpcConnection, error) {
		conn := &grpcConn{}

		return sdsClient, conn, nil
	}, 1000, 10)

	resultChan := make(chan ListingResult, 100)

	sds.GetListing(ctx, 1, resultChan)

	result1 := <-resultChan
	assert.Equal(t, int32(1), result1.Listing.Id)

}

func TestRetryBehaviourWhenNeverSucceeds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sdsClient := mocks.NewMockStaticDataServiceClient(mockCtrl)
	sdsClient.EXPECT().GetListing(gomock.Any(), gomock.Any()).Return(nil, errors.New("not ready")).AnyTimes()

	sds, _ := newStaticDataSource(ctx, func() (services.StaticDataServiceClient, GrpcConnection, error) {
		conn := &grpcConn{}

		return sdsClient, conn, nil
	}, 1000, 2)

	resultChan := make(chan ListingResult, 100)

	sds.GetListing(ctx, 1, resultChan)

	result1 := <-resultChan
	assert.Nil(t, result1.Listing)
	assert.Error(t, result1.Err)
}

type grpcConn struct {
}

func (g grpcConn) GetState() connectivity.State {
	return connectivity.Ready
}

func (g grpcConn) WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool {
	panic("implement me")
}

type sdsClient struct {
	listings map[int32]*model.Listing
}

func (s sdsClient) GetListingsWithSameInstrument(ctx context.Context, in *services.ListingId, opts ...grpc.CallOption) (*services.Listings, error) {
	panic("implement me")
}

func (s sdsClient) GetListingMatching(ctx context.Context, in *services.ExactMatchParameters, opts ...grpc.CallOption) (*model.Listing, error) {
	panic("implement me")
}

func (s sdsClient) GetListingsMatching(ctx context.Context, in *services.MatchParameters, opts ...grpc.CallOption) (*services.Listings, error) {
	panic("implement me")
}

func (s sdsClient) GetListing(ctx context.Context, in *services.ListingId, opts ...grpc.CallOption) (*model.Listing, error) {
	if listing, ok := s.listings[in.ListingId]; ok {
		return listing, nil
	}

	panic("listing not found")
}

func (s sdsClient) GetListings(ctx context.Context, in *services.ListingIds, opts ...grpc.CallOption) (*services.Listings, error) {
	panic("implement me")
}
