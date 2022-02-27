package handlers

import (
	"fmt"
	//"strconv"
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

	account := &repository.Account{}
	router  := &repository.Router{}

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

func NewRouteHandler(
	routeStore repository.RouteRepository,
	profileStore repository.ProfileRepository,
	instrumentStore repository.InstrumentRepository,
	accountStore repository.AccountRepository,
	routerStore repository.RouterRepository,
	loggerFunc repository.LoggerFunc,
) *routeHandler {
	return &routeHandler{
		loggerFunc:      loggerFunc,
		routeStore:      routeStore,
		profileStore:    profileStore,
		instrumentStore: instrumentStore,
		accountStore:    accountStore,
		routerStore:     routerStore,
	}
}
