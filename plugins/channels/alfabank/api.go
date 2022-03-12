package alfabank

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
	Id  = 2
	Key = "alfabank"
	Registered = plugins.RegisterBankChannel(Id, Key, func(
		cfg     *config.Config,
		account *repository.Account,
		logger  repository.LoggerFunc,
	) channels.BankChannel {
		return &AlfaBankChannel{
			cfg:     cfg,
			logger:  logger,
			account: account,
		}
	})
)

type AlfaBankChannel struct {
	cfg     *config.Config
	logger  repository.LoggerFunc
	account *repository.Account
}

func (abc *AlfaBankChannel) SutableForInstrument(instrument *repository.Instrument) bool {
	return *instrument.Id == bankcard.Id
}

func (abc *AlfaBankChannel) Authorize(c *gin.Context, transaction *repository.Transaction, instrumentInstance interface{}) error {
	card, ok := instrumentInstance.(*repository.Card)
	if !ok {
		return fmt.Errorf("instrumentInstance has wrong type")
	}

	abc.logger(c).Printf("authorize card: %v", card)
	abc.logger(c).Printf("url: %v", abc.cfg.Alfabank.Ecom.Url)

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
