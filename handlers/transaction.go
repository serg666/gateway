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
	loggerFunc      repository.LoggerFunc
	profileStore    repository.ProfileRepository
	accountStore    repository.AccountRepository
	currencyStore   repository.CurrencyRepository
	channelStore    repository.ChannelRepository
	instrumentStore repository.InstrumentRepository
	routeStore      repository.RouteRepository
	routerStore     repository.RouterRepository
}

func (th *transactionHandler) AuthorizeHandler(c *gin.Context) {
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

	instrumentKey := c.Params.ByName("instrument")
	err, _, instruments := th.instrumentStore.Query(c, repository.NewInstrumentSpecificationByKey(
		instrumentKey,
	))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(instruments) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Instrument with key=%s not found", instrumentKey),
		})
		return
	}

	err, instrumentApi := plugins.InstrumentApi(instruments[0], th.loggerFunc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
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
		th.loggerFunc(c).Printf("instrument: %v, %v (%T)", instrumentApi, instrumentApi)
		bankApi.Authorize(c, instrumentApi)
	}

	c.JSON(http.StatusOK, gin.H{"message":"ok"})
}

func NewTransactionHandler(
	routeStore repository.RouteRepository,
	routerStore repository.RouterRepository,
	instrumentStore repository.InstrumentRepository,
	profileStore repository.ProfileRepository,
	accountStore repository.AccountRepository,
	channelStore repository.ChannelRepository,
	currencyStore repository.CurrencyRepository,
	loggerFunc repository.LoggerFunc,
) *transactionHandler {
	return &transactionHandler{
		loggerFunc:      loggerFunc,
		profileStore:    profileStore,
		accountStore:    accountStore,
		currencyStore:   currencyStore,
		channelStore:    channelStore,
		instrumentStore: instrumentStore,
		routeStore:      routeStore,
		routerStore:     routerStore,
	}
}
