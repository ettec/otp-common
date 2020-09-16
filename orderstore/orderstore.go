package orderstore

import (
	"github.com/ettec/otp-common/model"
)

type OrderStore interface {
	Write(order *model.Order) error
	RecoverInitialCache(loadOrder func(order *model.Order) bool) (map[string]*model.Order,  error)
	Close()
}
