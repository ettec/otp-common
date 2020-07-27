package executionvenue

import (
	"github.com/ettec/otp-common/model"
)

type OrderGateway interface {
	Send(order *model.Order, listing *model.Listing) error
	Cancel(order *model.Order) error
	Modify(order *model.Order, listing *model.Listing, Quantity *model.Decimal64, Price *model.Decimal64) error
}
