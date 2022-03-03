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

func (th *transactionHandler) route(
	c *gin.Context,
	profile *repository.Profile,
	instrument *repository.Instrument,
) (error, *repository.Account) {
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
		err, routerApi := plugins.RouterApi(route.Router, route.Settings, th.loggerFunc)
		if err != nil {
			return fmt.Errorf("Can not get router: %v", err), nil
		}

		if err := routerApi.Route(c, route.Account); err != nil {
			return fmt.Errorf("Can not get route: %v", err), nil
		}
	}

	return nil, route.Account
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

	profile := profiles[0]
	instrument := instruments[0]

	err, account := th.route(c, profile, instrument)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, bankApi := plugins.BankApi(account, th.loggerFunc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, instrumentApi := plugins.InstrumentApi(instrument, th.loggerFunc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	th.loggerFunc(c).Printf("bank api: %v (%T)", bankApi, bankApi)
	th.loggerFunc(c).Printf("instrument: %v (%T)", instrumentApi, instrumentApi)
	th.loggerFunc(c).Printf("acc: %v (%T)", account, account)
	bankApi.Authorize(c, instrumentApi)

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
