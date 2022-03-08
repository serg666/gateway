package alfabank

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/instruments/card"
	"github.com/serg666/gateway/plugins/channels"
	"github.com/serg666/repository"
)

var (
	Id  = 2
	Key = "alfabank"
	Registered = plugins.RegisterBankChannel(Id, Key, func(
		account *repository.Account,
		logger repository.LoggerFunc,
	) channels.BankChannel {
		return &AlfaBankChannel{
			logger:  logger,
			account: account,
		}
	})
)

type AlfaBankChannel struct {
	logger  repository.LoggerFunc
	account *repository.Account
}

func (abc *AlfaBankChannel) SutableForInstrument(instrument *repository.Instrument) bool {
	return *instrument.Id == bankcard.Id
}

func (abc *AlfaBankChannel) Authorize(c *gin.Context, instrument *repository.Instrument) error {
	abc.logger(c).Print("authorize")
	cardStore, exists := c.Get("cardStore")
	abc.logger(c).Printf("store exists: %v", exists)
	abc.logger(c).Printf("store: %v (%T)", cardStore, cardStore)

	return nil
}

func (abc *AlfaBankChannel) PreAuthorize(c *gin.Context) {
	abc.logger(c).Print("preauthorize")
}

func (abc *AlfaBankChannel) Confirm(c *gin.Context) {
	abc.logger(c).Print("confirm")
}

func (abc *AlfaBankChannel) Reverse(c *gin.Context) {
	abc.logger(c).Print("reverse")
}

func (abc *AlfaBankChannel) Refund(c *gin.Context) {
	abc.logger(c).Print("refund")
}

func (abc *AlfaBankChannel) Complete3DS(c *gin.Context) {
	abc.logger(c).Print("complete3ds")
}
