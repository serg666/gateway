package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/instruments/card"
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
	cardStore       repository.CardRepository
}

func (th *transactionHandler) route(
	c *gin.Context,
	profile *repository.Profile,
	instrument *repository.Instrument,
) (error, *repository.Route) {
	err, _, routes := th.routeStore.Query(c, repository.NewRouteSpecificationByProfileAndInstrument(
		profile,
		instrument,
	))

	if err != nil {
		return fmt.Errorf("Error quering route by profile and instrument: %v", err), nil
	}

	if len(routes) == 0 {
		return fmt.Errorf("Route by profile and instrument not found: %v", err), nil
	}

	route := routes[0]

	if route.Router != nil {
		err, routerApi := plugins.RouterApi(route, th.loggerFunc)
		if err != nil {
			return fmt.Errorf("Can not get router: %v", err), nil
		}

		if err := routerApi.Route(c, route); err != nil {
			return fmt.Errorf("Can not get route: %v", err), nil
		}
	}

	return nil, route
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

	err, _, instruments := th.instrumentStore.Query(c, repository.NewInstrumentSpecificationByKey("card"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(instruments) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprint("Card Instrument not found"),
		})
		return
	}

	profile := profiles[0]
	instrument := instruments[0]

	var req bankcard.CardAuthorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	// @note: set some variables used in routers and channels due to instrument
	c.Set("cardAuthorizeRequest", req)
	c.Set("cardStore", th.cardStore)
	c.Set("accountStore", th.accountStore)

	err, route := th.route(c, profile, instrument)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, bankApi := plugins.BankApi(route, th.loggerFunc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", route.Account)

	err, transaction := bankApi.Authorize(c, instrument)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func NewTransactionHandler(
	routeStore repository.RouteRepository,
	routerStore repository.RouterRepository,
	instrumentStore repository.InstrumentRepository,
	profileStore repository.ProfileRepository,
	accountStore repository.AccountRepository,
	channelStore repository.ChannelRepository,
	currencyStore repository.CurrencyRepository,
	cardStore repository.CardRepository,
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
		cardStore:       cardStore,
	}
}
