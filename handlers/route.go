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

type UpdateRouteRequest struct {
	Profile    *string                    `json:"profile_key" binding:"omitempty,required_without_all=Instrument Account Router Settings,notempty"`
	Instrument *string                    `json:"instrument_key" binding:"omitempty,required_without_all=Profile Account Router Settings,notempty"`
	Account    *int                       `json:"account_id" binding:"required_without_all=Profile Instrument Router Settings"`
	Router     *string                    `json:"router_key" binding:"omitempty,required_without_all=Profile Instrument Account Settings,notempty"`
	Settings   *repository.RouterSettings `json:"settings" binding:"required_without_all=Profile Instrument Account Router"`
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

	c.JSON(http.StatusOK, gin.H{
		"overall": overall,
		"routes": routes,
	})
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

	c.JSON(http.StatusOK, route)
}

func (rh *routeHandler) PatchRouteHandler(c *gin.Context) {
	var req UpdateRouteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var profile    *repository.Profile
	var instrument *repository.Instrument
	var account    *repository.Account
	var router     *repository.Router

	if req.Profile != nil {
		err, _, profiles := rh.profileStore.Query(c, repository.NewProfileSpecificationByKey(*req.Profile))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(profiles) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Profile with key=%v not found", *req.Profile),
			})
			return
		}
		profile = profiles[0]
	}

	if req.Instrument != nil {
		err, _, instruments := rh.instrumentStore.Query(c, repository.NewInstrumentSpecificationByKey(*req.Instrument))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(instruments) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Instrument with key=%v not found", *req.Instrument),
			})
			return
		}
		instrument = instruments[0]
	}

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
				"message": fmt.Sprintf("Account with ID=%v not found", *req.Account),
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
				"message": fmt.Sprintf("Router with key=%v not found", *req.Router),
			})
			return
		}
		router = routers[0]
	}

	route := &repository.Route{
		Id:         &id,
		Profile:    profile,
		Instrument: instrument,
		Account:    account,
		Router:     router,
		Settings:   req.Settings,
	}

	err, notfound := rh.routeStore.Update(c, route)

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
