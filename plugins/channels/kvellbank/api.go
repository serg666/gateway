package kvellbank

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/config"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/instruments/card"
	"github.com/serg666/gateway/plugins/channels"
	"github.com/serg666/repository"
)

var (
	Id  = 1
	Key = "kvellbank"
	Registered = plugins.RegisterBankChannel(Id, Key, func(
		cfg             *config.Config,
		route           *repository.Route,
		instrumentStore interface{},
		logger          repository.LoggerFunc,
	) (error, channels.BankChannel) {
		if *route.Instrument.Id != bankcard.Id {
			return fmt.Errorf("kvellbank channel not sutable for instrument <%d>", *route.Instrument.Id), nil
		}

		return nil, &KvellBankChannel{
			cfg:             cfg,
			instrumentStore: instrumentStore,
			logger:          logger,
		}
	})
)

type KvellBankChannel struct {
	cfg             *config.Config
	instrumentStore interface{}
	logger          repository.LoggerFunc
}

func (kbc *KvellBankChannel) Authorize(c *gin.Context, transaction *repository.Transaction, request interface{}) error {
	kbc.logger(c).Print("authorize int")
	return nil
}

func (kbc *KvellBankChannel) PreAuthorize(c *gin.Context) {
	kbc.logger(c).Print("preauthorize int")
}

func (kbc *KvellBankChannel) Confirm(c *gin.Context) {
	kbc.logger(c).Print("confirm int")
}

func (kbc *KvellBankChannel) Reverse(c *gin.Context) {
	kbc.logger(c).Print("reverse int")
}

func (kbc *KvellBankChannel) Refund(c *gin.Context) {
	kbc.logger(c).Print("refund int")
}

func (kbc *KvellBankChannel) Complete3DS(c *gin.Context) {
	kbc.logger(c).Print("complete3ds int")
}
