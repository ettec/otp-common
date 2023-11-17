package ordermanagement

import (
	"context"
	"github.com/ettec/otp-common/model"
	"testing"
)

func TestChildOrderUpdatesDistribution(t *testing.T) {

	updates := make(chan ChildOrder, 100)

	d := NewChildOrderUpdatesDistributor(updates, 1000)

	aStream := d.GetChildOrderStream("a", 200)
	bStream := d.GetChildOrderStream("b", 200)

	updates <- ChildOrder{
		ParentOrderId: "a",
		Child:         &model.Order{Id: "a1"},
	}

	updates <- ChildOrder{
		ParentOrderId: "b",
		Child:         &model.Order{Id: "b1"},
	}

	d.Start(context.Background())

	u := <-aStream.Chan()
	if u.Id != "a1" {
		t.FailNow()
	}

	u = <-bStream.Chan()
	if u.Id != "b1" {
		t.FailNow()
	}

	cStream := d.GetChildOrderStream("c", 200)

	updates <- ChildOrder{
		ParentOrderId: "a",
		Child:         &model.Order{Id: "a1"},
	}

	updates <- ChildOrder{
		ParentOrderId: "c",
		Child:         &model.Order{Id: "c1"},
	}

	u = <-aStream.Chan()
	if u.Id != "a1" {
		t.FailNow()
	}

	u = <-cStream.Chan()
	if u.Id != "c1" {
		t.FailNow()
	}

}

func TestClosingChildOrderStream(t *testing.T) {

	updates := make(chan ChildOrder, 100)

	d := NewChildOrderUpdatesDistributor(updates, 1000)

	aStream := d.GetChildOrderStream("a", 200)
	bStream := d.GetChildOrderStream("b", 200)

	updates <- ChildOrder{
		ParentOrderId: "a",
		Child:         &model.Order{Id: "a1"},
	}

	updates <- ChildOrder{
		ParentOrderId: "b",
		Child:         &model.Order{Id: "b1"},
	}

	d.Start(context.Background())

	u := <-aStream.Chan()
	if u.Id != "a1" {
		t.FailNow()
	}

	u = <-bStream.Chan()
	if u.Id != "b1" {
		t.FailNow()
	}

	aStream.Close()

	updates <- ChildOrder{
		ParentOrderId: "a",
		Child:         &model.Order{Id: "a1"},
	}

	updates <- ChildOrder{
		ParentOrderId: "b",
		Child:         &model.Order{Id: "b1"},
	}

	_, ok := <-aStream.Chan()
	if ok {
		t.FailNow()
	}

	u = <-bStream.Chan()
	if u.Id != "b1" {
		t.FailNow()
	}

}

func TestBlockedStreamDoesNotStopOtherStreamEvents(t *testing.T) {

	updates := make(chan ChildOrder, 100)

	d := NewChildOrderUpdatesDistributor(updates, 1000)

	d.GetChildOrderStream("a", 1)
	bStream := d.GetChildOrderStream("b", 1)

	updates <- ChildOrder{
		ParentOrderId: "a",
		Child:         &model.Order{Id: "a1"},
	}

	updates <- ChildOrder{
		ParentOrderId: "b",
		Child:         &model.Order{Id: "b1"},
	}

	d.Start(context.Background())

	u := <-bStream.Chan()
	if u.Id != "b1" {
		t.FailNow()
	}

	updates <- ChildOrder{
		ParentOrderId: "a",
		Child:         &model.Order{Id: "a1"},
	}

	updates <- ChildOrder{
		ParentOrderId: "b",
		Child:         &model.Order{Id: "b1"},
	}

	u = <-bStream.Chan()
	if u.Id != "b1" {
		t.FailNow()
	}
}
