package staticdata

import (
	"context"
	services "github.com/ettec/otp-common/api/staticdataservice"
	"github.com/ettec/otp-common/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"testing"
	"time"
)

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
	if listing, ok := s.listings[in.ListingId];ok {
		return listing, nil
	}


	panic("listing not found")
}

func (s sdsClient) GetListings(ctx context.Context, in *services.ListingIds, opts ...grpc.CallOption) (*services.Listings, error) {
	panic("implement me")
}

func Test_listingSource_GetListing(t *testing.T) {

	listings := map[int32]*model.Listing{}
	listing1 := &model.Listing{Id: 1}
	listings[listing1.Id] = listing1

	sds, _:= newStaticDataSource( func() (services.StaticDataServiceClient, GrpcConnection, error) {
		sdsClient := &sdsClient{listings: listings}
		conn := &grpcConn{}

		return sdsClient, conn, nil
	}, 1000)

	resultChan := make(chan *model.Listing, 100)

	sds.GetListing(1, resultChan)

	result1 := <-resultChan
	if result1.Id != 1 {
		t.FailNow()
	}

	time.Sleep(1*time.Second)

	sds.GetListing(1, resultChan)

	time.Sleep(1*time.Second)

	result2 := <-resultChan
	if result2.Id != 1 {
		t.FailNow()
	}

	if len(resultChan) > 0 {
		t.FailNow()
	}


}
