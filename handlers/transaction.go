package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/repository"
)

type transactionHandler struct {
	loggerFunc    repository.LoggerFunc
	profileStore  repository.ProfileRepository
	accountStore  repository.AccountRepository
	currencyStore repository.CurrencyRepository
	channelStore  repository.ChannelRepository
}

func (th *transactionHandler) CardAuthorizeHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, _, profiles := th.profileStore.Query(c, repository.NewProfileSpecificationByID(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Profile with id=%v not found", id),
		})
		return
	}

	// @todo: somehow find channel and account within the channel to make authorize request to bank api (routing) 
	err, overall, channels := th.channelStore.Query(nil, repository.NewChannelSpecificationByID(2))
	th.loggerFunc(c).Printf("err: %v, overall: %v, channels: %v", err, overall, channels[0])

	err, overall, accounts := th.accountStore.Query(nil, repository.NewAccountSpecificationByID(26))
	th.loggerFunc(c).Printf("err: %v, overall: %v, accounts: %v", err, overall, accounts[0])

	err, bankApi := plugins.BankApi(channels[0], accounts[0], th.loggerFunc)
	th.loggerFunc(c).Printf("err: %v", err)
	if err == nil {
		th.loggerFunc(c).Printf("bank api: %v %T", bankApi, bankApi)
		bankApi.Authorize(c)
	}

	c.JSON(http.StatusOK, gin.H{"message":"ok"})
}

func NewTransactionHandler(
	profileStore repository.ProfileRepository,
	accountStore repository.AccountRepository,
	channelStore repository.ChannelRepository,
	currencyStore repository.CurrencyRepository,
	loggerFunc repository.LoggerFunc,
) *transactionHandler {
	return &transactionHandler{
		loggerFunc:    loggerFunc,
		profileStore:  profileStore,
		accountStore:  accountStore,
		currencyStore: currencyStore,
		channelStore:  channelStore,
	}
}
