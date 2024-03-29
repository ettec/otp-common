package strategy

import (
	"context"
	api "github.com/ettec/otp-common/api/executionvenue"
	"github.com/ettec/otp-common/model"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"log"
	"testing"
	"time"
)

func Test_ParentOrderUpdatesOnlySentWhenParentOrderChanged(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parentOrderUpdatesChan, _, _, _, _, _, _, _ := setupStrategy(ctx,
		func(ctx context.Context, om *Strategy, sendChildQty chan *model.Decimal64, listing *model.Listing) {

			go func() {
				if om.ParentOrder.GetTargetStatus() == model.OrderStatus_LIVE {
					om.ParentOrder.SetStatus(model.OrderStatus_LIVE)
				}

				for i := 0; i < 6; i++ {
					om.CheckIfDone(ctx)
				}
			}()
		})

	go func() {
		time.Sleep(3 * time.Second)
		close(parentOrderUpdatesChan)
	}()

	numUpdates := 0
	for range parentOrderUpdatesChan {
		numUpdates++
	}

	if numUpdates != 1 {
		t.FailNow()
	}

}

func Test_SendingChildOrders(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupStrategyAndSendTwoChildOrders(ctx, t)
}

func Test_StrategyCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parentOrderUpdatesChan, _, cancelOrderOutboundParams, childOrdersIn, _,
		listing, om, doneChan, child1Id, child2Id := setupStrategyAndSendTwoChildOrders(ctx, t)

	om.CancelChan <- ""

	cp1 := <-cancelOrderOutboundParams

	if cp1.OrderId != child1Id && cp1.OrderId != child2Id {
		t.FailNow()
	}

	cp2 := <-cancelOrderOutboundParams

	if cp2.OrderId != child2Id && cp1.OrderId != child2Id {
		t.FailNow()
	}

	update := <-parentOrderUpdatesChan
	if update.GetTargetStatus() != model.OrderStatus_CANCELLED {
		t.FailNow()
	}

	childOrdersIn <- &model.Order{
		Id:                child1Id,
		Version:           2,
		ListingId:         listing.Id,
		Status:            model.OrderStatus_CANCELLED,
		RemainingQuantity: model.IasD(60),
	}

	update = <-parentOrderUpdatesChan
	if !update.GetExposedQuantity().Equal(model.IasD(40)) {
		t.FailNow()
	}

	childOrdersIn <- &model.Order{
		Id:                child2Id,
		Version:           2,
		ListingId:         listing.Id,
		Status:            model.OrderStatus_CANCELLED,
		RemainingQuantity: model.IasD(40),
	}

	update = <-parentOrderUpdatesChan
	if !update.GetExposedQuantity().Equal(model.IasD(0)) {
		t.FailNow()
	}

	if update.GetStatus() != model.OrderStatus_CANCELLED {
		t.FailNow()
	}

	id := <-doneChan
	if id != om.ParentOrder.Id {
		t.FailNow()
	}

}

func Test_cancelOfUnexposedOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parentOrderUpdatesChan, _, _, _, _, _, om, _ := setupStrategy(ctx, ExecuteAsDmaStrategy)

	order := <-parentOrderUpdatesChan

	if order.GetTargetStatus() != model.OrderStatus_NONE || order.GetStatus() != model.OrderStatus_LIVE {
		t.FailNow()
	}

	if !order.GetExposedQuantity().Equal(model.IasD(0)) {
		t.Fatalf("parent order should not be exposed")
	}

	om.CancelChan <- ""

	order = <-parentOrderUpdatesChan

	if order.GetTargetStatus() != model.OrderStatus_NONE || order.GetStatus() != model.OrderStatus_CANCELLED {
		t.FailNow()
	}

	if !order.GetExposedQuantity().Equal(model.IasD(0)) {
		t.Fatalf("parent order should be only partly exposed")
	}

}

func Test_cancelOfPartiallyExposedOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parentOrderUpdatesChan, childOrderOutboundParams, cancelOrderOutboundParams, childOrdersIn, sendChildQty, listing, om, doneChan := setupStrategy(ctx, ExecuteAsDmaStrategy)

	params1 := &api.CreateAndRouteOrderParams{
		OrderSide:     model.Side_BUY,
		Quantity:      model.IasD(10),
		Price:         model.IasD(200),
		ListingId:     listing.Id,
		Destination:   listing.Market.Mic,
		OriginatorId:  om.ExecVenueId,
		OriginatorRef: om.ParentOrder.Id,
	}

	<-parentOrderUpdatesChan

	sendChildQty <- model.IasD(10)
	pd := <-childOrderOutboundParams

	if !areParamsEqual(params1, pd.params) {
		t.FailNow()
	}

	order := <-parentOrderUpdatesChan

	if order.GetTargetStatus() != model.OrderStatus_NONE || order.GetStatus() != model.OrderStatus_LIVE {
		t.FailNow()
	}

	if !order.GetExposedQuantity().Equal(model.IasD(10)) {
		t.Fatalf("parent order should be only partly exposed")
	}

	childOrdersIn <- &model.Order{
		Id:                pd.id,
		Version:           1,
		ListingId:         listing.Id,
		Status:            model.OrderStatus_LIVE,
		RemainingQuantity: model.IasD(10),
	}

	order = <-parentOrderUpdatesChan

	if order.GetTargetStatus() != model.OrderStatus_NONE || order.GetStatus() != model.OrderStatus_LIVE {
		t.FailNow()
	}

	if !order.GetExposedQuantity().Equal(model.IasD(10)) {
		t.Fatalf("parent order should be only partly exposed")
	}

	om.CancelChan <- ""

	cp := <-cancelOrderOutboundParams

	if cp.OrderId != pd.id {
		t.FailNow()
	}

	order = <-parentOrderUpdatesChan

	if order.GetTargetStatus() != model.OrderStatus_CANCELLED || order.GetStatus() != model.OrderStatus_LIVE {
		t.FailNow()
	}

	if !order.GetExposedQuantity().Equal(model.IasD(10)) {
		t.Fatalf("parent order should be only partly exposed")
	}

	childOrdersIn <- &model.Order{
		Id:                pd.id,
		Version:           2,
		ListingId:         listing.Id,
		Status:            model.OrderStatus_CANCELLED,
		RemainingQuantity: model.IasD(10),
	}

	order = <-parentOrderUpdatesChan

	if order.GetTargetStatus() != model.OrderStatus_NONE || order.GetStatus() != model.OrderStatus_CANCELLED {
		t.FailNow()
	}

	if !order.GetExposedQuantity().Equal(model.IasD(0)) {
		t.Fatalf("parent order should be not be exposed")
	}

	id := <-doneChan
	if id != om.ParentOrder.Id {
		t.FailNow()
	}

}

func TestStrategyCompletesWhenChildOrdersFilled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parentOrderUpdatesChan, _, _, childOrdersIn, _,
		listing, om, doneChan, child1Id, child2Id := setupStrategyAndSendTwoChildOrders(ctx, t)

	childOrdersIn <- &model.Order{
		Id:                child1Id,
		Version:           2,
		Status:            model.OrderStatus_LIVE,
		LastExecQuantity:  model.IasD(60),
		LastExecPrice:     model.IasD(100),
		LastExecId:        "c1e1",
		RemainingQuantity: model.IasD(0),
		ListingId:         listing.GetId(),
	}

	order := <-parentOrderUpdatesChan

	if !order.GetTradedQuantity().Equal(model.IasD(60)) {
		t.FailNow()
	}

	childOrdersIn <- &model.Order{
		Id:                child2Id,
		Version:           2,
		Status:            model.OrderStatus_LIVE,
		LastExecQuantity:  model.IasD(40),
		LastExecPrice:     model.IasD(110),
		LastExecId:        "c2e1",
		ListingId:         listing.GetId(),
		RemainingQuantity: model.IasD(0),
	}

	order = <-parentOrderUpdatesChan

	if !order.GetTradedQuantity().Equal(model.IasD(100)) {
		t.FailNow()
	}

	if order.GetStatus() != model.OrderStatus_FILLED {
		t.FailNow()
	}

	doneId := <-doneChan

	if doneId != om.getStrategyOrderId() {
		t.FailNow()
	}

}

func setupStrategyAndSendTwoChildOrders(ctx context.Context, t *testing.T) (parentOrderUpdatesChan chan model.Order, childOrderOutboundParams chan paramsAndId,
	childOrderCancelParams chan *api.CancelOrderParams, childOrdersIn chan *model.Order,
	sendChildQty chan *model.Decimal64, listing *model.Listing,
	om *Strategy, doneChan chan string, child1Id string, child2Id string) {

	parentOrderUpdatesChan, childOrderOutboundParams, childOrderCancelParams, childOrdersIn, sendChildQty, listing,
		om, doneChan = setupStrategy(ctx, ExecuteAsDmaStrategy)

	<-parentOrderUpdatesChan

	sendChildQty <- &model.Decimal64{Mantissa: 60}

	params1 := &api.CreateAndRouteOrderParams{
		OrderSide:     model.Side_BUY,
		Quantity:      model.IasD(60),
		Price:         model.IasD(200),
		ListingId:     listing.Id,
		Destination:   listing.Market.Mic,
		OriginatorId:  om.ExecVenueId,
		OriginatorRef: om.ParentOrder.Id,
	}

	pd := <-childOrderOutboundParams
	child1Id = pd.id

	if !areParamsEqual(params1, pd.params) {
		t.FailNow()
	}

	<-parentOrderUpdatesChan

	sendChildQty <- &model.Decimal64{Mantissa: 40}
	params2 := &api.CreateAndRouteOrderParams{
		OrderSide:     model.Side_BUY,
		Quantity:      model.IasD(40),
		Price:         model.IasD(200),
		ListingId:     listing.Id,
		Destination:   listing.Market.Mic,
		OriginatorId:  om.ExecVenueId,
		OriginatorRef: om.ParentOrder.Id,
	}

	pd = <-childOrderOutboundParams
	child2Id = pd.id

	if !areParamsEqual(params2, pd.params) {
		t.FailNow()
	}

	order := <-parentOrderUpdatesChan

	if order.GetTargetStatus() != model.OrderStatus_NONE || order.GetStatus() != model.OrderStatus_LIVE {
		t.FailNow()
	}

	if order.GetAvailableQty().GreaterThan(model.IasD(0)) {
		t.Fatalf("no quantity should be left to trade")
	}

	childOrdersIn <- &model.Order{
		Id:                child1Id,
		Version:           0,
		Status:            model.OrderStatus_LIVE,
		ListingId:         listing.GetId(),
		RemainingQuantity: model.IasD(60),
	}

	order = <-parentOrderUpdatesChan
	if !order.GetExposedQuantity().Equal(model.IasD(100)) {
		t.FailNow()
	}

	childOrdersIn <- &model.Order{
		Id:                child2Id,
		Version:           0,
		Status:            model.OrderStatus_LIVE,
		ListingId:         listing.Id,
		RemainingQuantity: model.IasD(40),
	}

	order = <-parentOrderUpdatesChan
	if !order.GetExposedQuantity().Equal(model.IasD(100)) {
		t.FailNow()
	}

	return parentOrderUpdatesChan, childOrderOutboundParams, childOrderCancelParams, childOrdersIn, sendChildQty,
		listing, om, doneChan, child1Id, child2Id
}

func setupStrategy(ctx context.Context, strategy func(ctx context.Context, om *Strategy, sendChildQty chan *model.Decimal64, listing *model.Listing)) (parentOrderUpdatesChan chan model.Order, childOrderOutboundParamsChan chan paramsAndId,
	childOrderCancelParamsChan chan *api.CancelOrderParams, childOrdersIn chan *model.Order,
	sendChildQty chan *model.Decimal64, listing *model.Listing,
	om *Strategy, doneChan chan string) {

	listing = &model.Listing{
		Version: 0,
		Id:      1,
		Market:  &model.Market{Mic: "XNAS"},
	}

	parentOrderUpdatesChan = make(chan model.Order)

	childOrderOutboundParamsChan = make(chan paramsAndId)
	childOrderCancelParamsChan = make(chan *api.CancelOrderParams)
	orderRouter := &testOmClient{
		croParamsChan:    childOrderOutboundParamsChan,
		cancelParamsChan: childOrderCancelParamsChan,
	}

	childOrdersIn = make(chan *model.Order)
	childOrderStream := testChildOrderStream{stream: childOrdersIn}

	doneChan = make(chan string)

	om, err := NewStrategyFromCreateParams("p1", &api.CreateAndRouteOrderParams{
		OrderSide:          model.Side_BUY,
		Quantity:           &model.Decimal64{Mantissa: 100},
		Price:              &model.Decimal64{Mantissa: 200},
		ListingId:          listing.Id,
		Destination:        listing.Market.Mic,
		OriginatorId:       "",
		OriginatorRef:      "",
		RootOriginatorId:   "",
		RootOriginatorRef:  "",
		ExecParametersJson: "",
	}, "e1", func(ctx context.Context, o *model.Order) error {
		parentOrderUpdatesChan <- *o
		return nil
	}, orderRouter, childOrderStream, doneChan)
	if err != nil {
		panic(err)
	}

	sendChildQty = make(chan *model.Decimal64)
	strategy(ctx, om, sendChildQty, listing)
	return parentOrderUpdatesChan, childOrderOutboundParamsChan, childOrderCancelParamsChan, childOrdersIn, sendChildQty, listing, om, doneChan
}

func ExecuteAsDmaStrategy(ctx context.Context, om *Strategy, sendChildQty chan *model.Decimal64, listing *model.Listing) {

	if om.ParentOrder.GetTargetStatus() == model.OrderStatus_LIVE {
		om.ParentOrder.SetStatus(model.OrderStatus_LIVE)
	}
	go func() {
		for {
			done, err := om.CheckIfDone(ctx)
			if err != nil {
				om.Log.Error("failed to check if done, cancelling order", "error", err)
				om.CancelChan <- ""
			}

			if done {
				return
			}

			select {
			case <-ctx.Done():
				return
			case errMsg := <-om.CancelChan:
				if errMsg != "" {
					om.ParentOrder.ErrorMessage = errMsg
				}

				err := om.CancelChildOrdersAndStrategyOrder()
				if err != nil {
					log.Panicf("failed to Cancel order:%v", err)
				}
			case co, ok := <-om.ChildOrderUpdateChan:
				if ok {
					om.OnChildOrderUpdate(ok, co)
				} else {
					msg := "child order update chan unexpectedly closed, cancelling order"
					om.Log.Info(msg)
					om.CancelChan <- msg
				}
			case q := <-sendChildQty:
				om.SendChildOrder(om.ParentOrder.Side, q, om.ParentOrder.Price, listing.Id, listing.GetMarket().Mic, "")
			}
		}
	}()

}

func areParamsEqual(p1 *api.CreateAndRouteOrderParams, p2 *api.CreateAndRouteOrderParams) bool {
	return p1.Quantity.Equal(p2.Quantity) && p1.ListingId == p2.ListingId && p1.Price.Equal(p2.Price) && p1.OrderSide == p2.OrderSide &&
		p1.OriginatorRef == p2.OriginatorRef && p1.OriginatorId == p2.OriginatorId

}

type testEvClient struct {
	params []*api.CreateAndRouteOrderParams
}

func (t *testEvClient) GetExecutionParametersMetaData(ctx context.Context, empty *model.Empty, opts ...grpc.CallOption) (*api.ExecParamsMetaDataJson, error) {
	panic("implement me")
}

func (t *testEvClient) CreateAndRouteOrder(ctx context.Context, in *api.CreateAndRouteOrderParams, opts ...grpc.CallOption) (*api.OrderId, error) {
	t.params = append(t.params, in)
	id, _ := uuid.NewUUID()
	return &api.OrderId{
		OrderId: id.String(),
	}, nil
}

type paramsAndId struct {
	params *api.CreateAndRouteOrderParams
	id     string
}

type testOmClient struct {
	croParamsChan    chan paramsAndId
	cancelParamsChan chan *api.CancelOrderParams
}

func (t *testOmClient) GetExecutionParametersMetaData(ctx context.Context, empty *model.Empty, opts ...grpc.CallOption) (*api.ExecParamsMetaDataJson, error) {
	panic("implement me")
}

func (t *testOmClient) CreateAndRouteOrder(ctx context.Context, in *api.CreateAndRouteOrderParams, opts ...grpc.CallOption) (*api.OrderId, error) {

	id, _ := uuid.NewUUID()

	t.croParamsChan <- paramsAndId{in, id.String()}

	return &api.OrderId{
		OrderId: id.String(),
	}, nil
}

func (t *testOmClient) CancelOrder(ctx context.Context, in *api.CancelOrderParams, opts ...grpc.CallOption) (*model.Empty, error) {
	t.cancelParamsChan <- in
	return &model.Empty{}, nil
}

func (t *testOmClient) ModifyOrder(ctx context.Context, in *api.ModifyOrderParams, opts ...grpc.CallOption) (*model.Empty, error) {
	panic("implement me")
}

func (t *testEvClient) CancelOrder(ctx context.Context, in *api.CancelOrderParams, opts ...grpc.CallOption) (*model.Empty, error) {
	panic("implement me")
}

func (t *testEvClient) ModifyOrder(ctx context.Context, in *api.ModifyOrderParams, opts ...grpc.CallOption) (*model.Empty, error) {
	panic("implement me")
}

type testChildOrderStream struct {
	stream chan *model.Order
}

func (t testChildOrderStream) Chan() <-chan *model.Order {
	return t.stream
}

func (t testChildOrderStream) Close() {
}
