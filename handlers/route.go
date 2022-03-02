package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type CreateRouteRequest struct {
	Profile    *string                    `json:"profile_key" binding:"required,notempty"`
	Instrument *string                    `json:"instrument_key" binding:"required,notempty"`
	Account    *int                       `json:"account_id" binding:"required_without=Router"`
	Router     *string                    `json:"router_key" binding:"omitempty,required_without=Account,notempty"`
	Settings   *repository.RouterSettings `json:"settings"`
}

type routeHandler struct {
	loggerFunc      repository.LoggerFunc
	routeStore      repository.RouteRepository
	profileStore    repository.ProfileRepository
	instrumentStore repository.InstrumentRepository
	accountStore    repository.AccountRepository
	routerStore     repository.RouterRepository
	currencyStore   repository.CurrencyRepository
	channelStore    repository.ChannelRepository
}

func (rh *routeHandler) CreateRouteHandler(c *gin.Context) {
	var req CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, _, profiles := rh.profileStore.Query(c, repository.NewProfileSpecificationByKey(*req.Profile))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(profiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Profile with key %s not found", *req.Profile),
		})
		return
	}

	err, _, instruments := rh.instrumentStore.Query(c, repository.NewInstrumentSpecificationByKey(*req.Instrument))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(instruments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Instrument with key %s not found", *req.Instrument),
		})
		return
	}

	var account *repository.Account
	var router  *repository.Router

	if req.Account != nil {
		err, _, accounts := rh.accountStore.Query(c, repository.NewAccountSpecificationByID(*req.Account))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(accounts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Account with ID=%d not found", *req.Account),
			})
			return
		}

		account = accounts[0]
	}

	if req.Router != nil {
		err, _, routers := rh.routerStore.Query(c, repository.NewRouterSpecificationByKey(*req.Router))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(routers) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Router with key=%s not found", *req.Router),
			})
			return
		}

		router = routers[0]
	}

	route := &repository.Route{
		Profile:    profiles[0],
		Instrument: instruments[0],
		Account:    account,
		Router:     router,
		Settings:   req.Settings,
	}
	if err := rh.routeStore.Add(c, route); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, route)
}

func RefreshRouteProfile(c *gin.Context, route *repository.Route, profileStore repository.ProfileRepository) error {
	if !(route.Profile != nil && route.Profile.Id != nil) {
		return nil
	}

	err, _, profiles := profileStore.Query(c, repository.NewProfileSpecificationByID(
		*route.Profile.Id,
	))

	if err != nil {
		return fmt.Errorf("Can not update route profile: %v", err)
	}

	for _, profile := range profiles {
		route.Profile = profile
	}

	return nil
}

func RefreshRouteRouter(c *gin.Context, route *repository.Route, routerStore repository.RouterRepository) error {
	if !(route.Router != nil && route.Router.Id != nil) {
		return nil
	}

	err, _, routers := routerStore.Query(c, repository.NewRouterSpecificationByID(
		*route.Router.Id,
	))

	if err != nil {
		return fmt.Errorf("Can not update route router: %v", err)
	}

	for _, router := range routers {
		route.Router = router
	}

	return nil
}

func RefreshRouteInstrument(c *gin.Context, route *repository.Route, instrumentStore repository.InstrumentRepository) error {
	if !(route.Instrument != nil && route.Instrument.Id != nil) {
		return nil
	}

	err, _, instruments := instrumentStore.Query(c, repository.NewInstrumentSpecificationByID(
		*route.Instrument.Id,
	))

	if err != nil {
		return fmt.Errorf("Can not update route instrument: %v", err)
	}

	for _, instrument := range instruments {
		route.Instrument = instrument
	}

	return nil
}

func RefreshRouteAccount(c *gin.Context, route *repository.Route, accountStore repository.AccountRepository) error {
	if !(route.Account != nil && route.Account.Id != nil) {
		return nil
	}

	err, _, accounts := accountStore.Query(c, repository.NewAccountSpecificationByID(
		*route.Account.Id,
	))

	if err != nil {
		return fmt.Errorf("Can not update route account: %v", err)
	}

	for _, account := range accounts {
		route.Account = account
	}

	return nil
}

func (rh *routeHandler) GetRoutesHandler(c *gin.Context) {
	var req LimitAndOffsetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, overall, routes := rh.routeStore.Query(c, repository.NewRouteSpecificationWithLimitAndOffset(
		req.Limit,
		req.Offset,
	))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	for _, route := range routes {
		if err := rh.refreshRouteForeigns(c, route); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"overall": overall,
		"routes": routes,
	})
}

func (rh *routeHandler) refreshRouteForeigns(c *gin.Context, route *repository.Route) error {
	if err := RefreshRouteProfile(c, route, rh.profileStore); err != nil {
		return err
	}

	if err := RefreshProfileCurrency(c, route.Profile, rh.currencyStore); err != nil {
		return err
	}

	if err := RefreshRouteInstrument(c, route, rh.instrumentStore); err != nil {
		return err
	}

	if err := RefreshRouteAccount(c, route, rh.accountStore); err != nil {
		return err
	}

	if err := RefreshAccountCurrency(c, route.Account, rh.currencyStore); err != nil {
		return err
	}

	if err := RefreshAccountChannel(c, route.Account, rh.channelStore); err != nil {
		return err
	}

	if err := RefreshRouteRouter(c, route, rh.routerStore); err != nil {
		return err
	}

	return nil
}

func (rh *routeHandler) GetRouteHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, _, routes := rh.routeStore.Query(c, repository.NewRouteSpecificationByID(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(routes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Account with id=%v not found", id),
		})
		return
	}

	for _, route := range routes {
		if err := rh.refreshRouteForeigns(c, route); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, routes[0])
}

func (rh *routeHandler) DeleteRouteHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	route := &repository.Route{
		Id: &id,
	}

	err, notfound := rh.routeStore.Delete(c, route)

	if notfound {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err !=  nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := rh.refreshRouteForeigns(c, route); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, route)
}

func NewRouteHandler(
	routeStore repository.RouteRepository,
	profileStore repository.ProfileRepository,
	instrumentStore repository.InstrumentRepository,
	accountStore repository.AccountRepository,
	routerStore repository.RouterRepository,
	currencyStore repository.CurrencyRepository,
	channelStore repository.ChannelRepository,
	loggerFunc repository.LoggerFunc,
) *routeHandler {
	return &routeHandler{
		loggerFunc:      loggerFunc,
		routeStore:      routeStore,
		profileStore:    profileStore,
		instrumentStore: instrumentStore,
		accountStore:    accountStore,
		routerStore:     routerStore,
		currencyStore:   currencyStore,
		channelStore:    channelStore,
	}
}
