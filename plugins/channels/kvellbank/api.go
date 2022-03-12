package kvellbank

import (
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
		cfg     *config.Config,
		logger  repository.LoggerFunc,
	) channels.BankChannel {
		return &KvellBankChannel{
			cfg:     cfg,
			logger:  logger,
		}
	})
)

type KvellBankChannel struct {
	cfg     *config.Config
	logger  repository.LoggerFunc
}

func (kbc *KvellBankChannel) SutableForInstrument(instrument *repository.Instrument) bool {
	return *instrument.Id == bankcard.Id
}

func (kbc *KvellBankChannel) DecodeSettings(settings *repository.AccountSettings) error {
	return nil
}

func (kbc *KvellBankChannel) Authorize(c *gin.Context, transaction *repository.Transaction, instrumentInstance interface{}) {
	kbc.logger(c).Print("authorize int")
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
