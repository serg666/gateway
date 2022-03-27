package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/config"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/validators"
	"github.com/serg666/repository"
)

type transactionHandler struct {
	cfg              *config.Config
	loggerFunc       repository.LoggerFunc
	profileStore     repository.ProfileRepository
	accountStore     repository.AccountRepository
	currencyStore    repository.CurrencyRepository
	channelStore     repository.ChannelRepository
	instrumentStore  repository.InstrumentRepository
	routeStore       repository.RouteRepository
	routerStore      repository.RouterRepository
	cardStore        repository.CardRepository
	transactionStore repository.TransactionRepository
	sessionStore     repository.SessionRepository
}

func (th *transactionHandler) route(
	c *gin.Context,
	profile *repository.Profile,
	instrument *repository.Instrument,
	instrumentStore interface{},
	request interface{},
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
		err, routerApi := plugins.RouterApi(route, th.accountStore, instrumentStore, th.loggerFunc)
		if err != nil {
			return fmt.Errorf("Can not get router: %v", err), nil
		}

		if err := routerApi.Route(c, route, request); err != nil {
			return fmt.Errorf("Can not get route: %v", err), nil
		}
	}

	return nil, route
}

func (th *transactionHandler) CompleteMethodUrlHandler(c *gin.Context) {
	var req validators.CompleteMethodUrlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	pid, err := strconv.Atoi(c.Params.ByName("pid"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, _, profiles := th.profileStore.Query(c, repository.NewProfileSpecificationByID(pid))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Profile with id=%v not found", pid),
		})
		return
	}

	profile := profiles[0]

	tid, err := strconv.Atoi(c.Params.ByName("tid"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, _, transactions := th.transactionStore.Query(c, repository.NewTransactionSpecificationByID(tid))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(transactions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Transaction with id=%v not found", tid),
		})
		return
	}

	transaction := transactions[0]

	if *transaction.Profile.Id != *profile.Id {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Transaction profile does not match profile: %d!=%d", *transaction.Profile.Id, *profile.Id),
		})
		return
	}

	// @note: BankApi should not accept route. It should accept account and instrument instead of route

	c.JSON(http.StatusOK, transaction)
}

func (th *transactionHandler) CardAuthorizeHandler(c *gin.Context) {
	var req validators.CardAuthorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	id, err := strconv.Atoi(c.Params.ByName("pid"))
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

	err, instrumentApi := plugins.InstrumentApi(
		instrument,
		th.cardStore,
		th.loggerFunc,
		validators.CardAuthorizationInstrumentRequester,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, instrumentInstance := instrumentApi.FromRequest(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	card, ok := instrumentInstance.(*repository.Card)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "instrumentInstance has wrong type",
		})
		return
	}

	err, route := th.route(c, profile, instrument, th.cardStore, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// @todo: BankApi should not accept route. It should accept account and instrument instead of route
	err, bankApi := plugins.BankApi(th.cfg, route, th.cardStore, th.sessionStore, th.loggerFunc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", route.Account)

	transaction := repository.NewTransaction("authorize", &req.OrderId, profile, route.Account, instrument, card.Id, &req.Amount, nil)

	if err := th.transactionStore.Add(c, transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := bankApi.Authorize(c, transaction, req); err != nil {
		mess := err.Error()
		transaction.Declined(&mess)
	}

	if err, notfound := th.transactionStore.Update(c, transaction); err != nil {
		th.loggerFunc(c).Warningf("failed to update transaction: %v (notfound: %v)", err, notfound)
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
	transactionStore repository.TransactionRepository,
	sessionStore repository.SessionRepository,
	cfg *config.Config,
	loggerFunc repository.LoggerFunc,
) *transactionHandler {
	return &transactionHandler{
		cfg:              cfg,
		loggerFunc:       loggerFunc,
		profileStore:     profileStore,
		accountStore:     accountStore,
		currencyStore:    currencyStore,
		channelStore:     channelStore,
		instrumentStore:  instrumentStore,
		routeStore:       routeStore,
		routerStore:      routerStore,
		cardStore:        cardStore,
		transactionStore: transactionStore,
		sessionStore:     sessionStore,
	}
}
