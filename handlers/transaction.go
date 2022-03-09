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
}

func (th *transactionHandler) route(
	c *gin.Context,
	profile *repository.Profile,
	instrument *repository.Instrument,
	instrumentInstance interface{},
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
		err, routerApi := plugins.RouterApi(route, th.accountStore, th.loggerFunc)
		if err != nil {
			return fmt.Errorf("Can not get router: %v", err), nil
		}

		if err := routerApi.Route(c, route, instrumentInstance); err != nil {
			return fmt.Errorf("Can not get route: %v", err), nil
		}
	}

	return nil, route
}

func (th *transactionHandler) CardAuthorizeHandler(c *gin.Context) {
	var req bankcard.CardAuthorizeRequest
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

	err, instrumentApi := plugins.InstrumentApi(instrument, th.loggerFunc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, instrumentInstance := instrumentApi.FromRequest(c, req, th.cardStore)
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

	err, route := th.route(c, profile, instrument, card)
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

	transaction := NewTransaction("authorize", &req.OrderId, profile, route.Account, instrument, card.Id, &req.Amount, nil)

	if err := th.transactionStore.Add(c, transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := bankApi.Authorize(c, transaction, card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func NewTransaction(
	txType string,
	orderId *string,
	profile *repository.Profile,
	account *repository.Account,
	instrument *repository.Instrument,
	instrumentId *int,
	amount *uint,
	reference *repository.Transaction,
) *repository.Transaction {
	txStatus := "new"
	return &repository.Transaction{
		Type: &txType,
		Status: &txStatus,
		Profile: profile,
		Account: account,
		Instrument: instrument,
		InstrumentId: instrumentId,
		Currency: profile.Currency,
		Amount: amount,
		AmountConverted: amount, // @todo: convert amount to account currency from profile currency
		CurrencyConverted: account.Currency,
		OrderId: orderId,
		Reference: reference,
	}
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
	loggerFunc repository.LoggerFunc,
) *transactionHandler {
	return &transactionHandler{
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
	}
}
