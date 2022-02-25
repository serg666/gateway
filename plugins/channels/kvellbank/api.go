package kvellbank

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/channels"
	"github.com/serg666/repository"
)

var (
	Id  = 1
	Key = "kvellbank"
	Registered = plugins.RegisterBankChannel(Id, Key, func(
		account *repository.Account,
		logger repository.LoggerFunc,
	) channels.BankChannel {
		return &KvellBankChannel{
			logger:  logger,
			account: account,
		}
	})
)

type KvellBankChannel struct {
	logger  repository.LoggerFunc
	account *repository.Account
}

func (kbc *KvellBankChannel) Authorize(c *gin.Context) {
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
