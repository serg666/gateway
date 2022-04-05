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
		cfg              *config.Config,
		account          *repository.Account,
		instrument       *repository.Instrument,
		sessionStore     repository.SessionRepository,
		transactionStore repository.TransactionRepository,
		logger           repository.LoggerFunc,
	) (error, channels.BankChannel) {
		if *instrument.Id != bankcard.Id {
			return fmt.Errorf("kvellbank channel not sutable for instrument <%d>", *instrument.Id), nil
		}

		return nil, &KvellBankChannel{
			cfg:              cfg,
			sessionStore:     sessionStore,
			transactionStore: transactionStore,
			logger:           logger,
		}
	})
)

type KvellBankChannel struct {
	cfg              *config.Config
	sessionStore     repository.SessionRepository
	transactionStore repository.TransactionRepository
	logger           repository.LoggerFunc
}

func (kbc *KvellBankChannel) Authorize(c *gin.Context, transaction *repository.Transaction, request interface{}) error {
	kbc.logger(c).Print("authorize int")
	return nil
}

func (kbc *KvellBankChannel) PreAuthorize(c *gin.Context, transaction *repository.Transaction, request interface{}) error {
	kbc.logger(c).Print("preauthorize int")
	return nil
}

func (kbc *KvellBankChannel) Confirm(c *gin.Context, transaction *repository.Transaction) error {
	kbc.logger(c).Print("confirm int")
	return nil
}

func (kbc *KvellBankChannel) Reverse(c *gin.Context, transaction *repository.Transaction) error {
	kbc.logger(c).Print("reverse int")
	return nil
}

func (kbc *KvellBankChannel) Refund(c *gin.Context, transaction *repository.Transaction) error {
	kbc.logger(c).Print("refund int")
	return nil
}

func (kbc *KvellBankChannel) ProcessCres(c *gin.Context, transaction *repository.Transaction, cres string) error {
	kbc.logger(c).Print("process cres")
	return nil
}

func (kbc *KvellBankChannel) ProcessPares(c *gin.Context, transaction *repository.Transaction, pares string) error {
	kbc.logger(c).Print("process pares")
	return nil
}

func (kbc *KvellBankChannel) CompleteMethodUrl(c *gin.Context, transaction *repository.Transaction, completed bool) error {
	kbc.logger(c).Print("complete method url")
	return nil
}
