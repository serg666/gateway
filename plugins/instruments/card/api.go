package bankcard

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/instruments"
	"github.com/serg666/repository"
)

var (
	Id  = 1
	Key = "card"
	Registered = plugins.RegisterPaymentInstrument(Id, Key, func(
		logger repository.LoggerFunc,
	) instruments.PaymentInstrument {
		return &BankCard{
			logger: logger,
		}
	})
)

type BankCard struct {
	logger repository.LoggerFunc
}

func (bc *BankCard) FromContext(c *gin.Context) {
	bc.logger(c).Print("from context")
}
