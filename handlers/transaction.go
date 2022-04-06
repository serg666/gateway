package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/config"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/channels"
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
	instrumentRequester plugins.InstrumentRequesterFunc,
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
		err, routerApi := plugins.RouterApi(route, th.accountStore, instrumentStore, instrumentRequester, th.loggerFunc)
		if err != nil {
			return fmt.Errorf("Can not get router: %v", err), nil
		}

		if err := routerApi.Route(c, route, request); err != nil {
			return fmt.Errorf("Can not get route: %v", err), nil
		}
	}

	return nil, route
}

func (th *transactionHandler) validate(c *gin.Context) (error, *repository.Transaction, channels.BankChannel) {
	pid, err := strconv.Atoi(c.Params.ByName("pid"))
	if err !=  nil {
		return fmt.Errorf("invalid profile id: %v", err), nil, nil
	}

	err, _, profiles := th.profileStore.Query(c, repository.NewProfileSpecificationByID(pid))

	if err != nil {
		return fmt.Errorf("faild to query profile store: %v", err), nil, nil
	}

	if len(profiles) == 0 {
		return fmt.Errorf("profile with id=%v not found", pid), nil, nil
	}

	profile := profiles[0]

	tid, err := strconv.Atoi(c.Params.ByName("tid"))
	if err !=  nil {
		return fmt.Errorf("invalid transaction id: %v", err), nil, nil
	}

	err, _, transactions := th.transactionStore.Query(c, repository.NewTransactionSpecificationByID(tid))

	if err != nil {
		return fmt.Errorf("faild to query transaction store: %v", err), nil, nil
	}

	if len(transactions) == 0 {
		return fmt.Errorf("transaction with id=%v not found", tid), nil, nil
	}

	transaction := transactions[0]

	if *transaction.Profile.Id != *profile.Id {
		return fmt.Errorf("incorrect transaction id: %v", tid), nil, nil
	}

	var instrumentStore interface{}

	// @note: depends on instrument type
	switch *transaction.Instrument.Key {
	case "card":
		instrumentStore = th.cardStore
	}

	err, bankApi := plugins.BankApi(th.cfg, transaction.Account, transaction.Instrument, instrumentStore, th.sessionStore, th.transactionStore, th.loggerFunc)
	if err != nil {
		return fmt.Errorf("faild to get bank api: %v", err), nil, nil
	}

	return nil, transaction, bankApi
}

func (th *transactionHandler) ProcessParesHandler(c *gin.Context) {
	var req validators.ProcessParesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, transaction, bankApi := th.validate(c)
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !transaction.Is3DSWaiting() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong state: %s", *transaction.Status),
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", transaction.Account)

	if err := bankApi.ProcessPares(c, transaction, req.Pares); err != nil {
		mess := err.Error()
		transaction.Declined(&mess)
	}

	if err, notfound := th.transactionStore.Update(c, transaction); err != nil {
		th.loggerFunc(c).Warningf("failed to update transaction: %v (notfound: %v)", err, notfound)
	}

	c.JSON(http.StatusOK, transaction)
}

func (th *transactionHandler) ProcessCresHandler(c *gin.Context) {
	var req validators.ProcessCresRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, transaction, bankApi := th.validate(c)
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !transaction.Is3DSWaiting() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong state: %s", *transaction.Status),
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", transaction.Account)

	if err := bankApi.ProcessCres(c, transaction, req.Cres); err != nil {
		mess := err.Error()
		transaction.Declined(&mess)
	}

	if err, notfound := th.transactionStore.Update(c, transaction); err != nil {
		th.loggerFunc(c).Warningf("failed to update transaction: %v (notfound: %v)", err, notfound)
	}

	c.JSON(http.StatusOK, transaction)
}

func (th *transactionHandler) CompleteMethodUrlHandler(c *gin.Context) {
	var req validators.CompleteMethodUrlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, transaction, bankApi := th.validate(c)
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !transaction.IsMethodUrlWaiting() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong state: %s", *transaction.Status),
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", transaction.Account)

	if err := bankApi.CompleteMethodUrl(c, transaction, *req.Completed); err != nil {
		mess := err.Error()
		transaction.Declined(&mess)
	}

	if err, notfound := th.transactionStore.Update(c, transaction); err != nil {
		th.loggerFunc(c).Warningf("failed to update transaction: %v (notfound: %v)", err, notfound)
	}

	c.JSON(http.StatusOK, transaction)
}

func (th *transactionHandler) ReverseHandler(c *gin.Context) {
	var req validators.ReversalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, transaction, bankApi := th.validate(c)
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !transaction.IsSuccess() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong state: %s", *transaction.Status),
		})
		return
	}

	if !transaction.IsPreAuth() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong type: %s", *transaction.Type),
		})
		return
	}

	if req.Amount > *transaction.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Incorrect amount: %d", req.Amount),
		})
		return
	}

	if req.Amount < *transaction.Amount && !*transaction.Account.PartialReversalEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Incorrect amount: %d, partial reversal not allowed", req.Amount),
		})
		return
	}

	if !*transaction.Account.ReversalEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Reversal not allowed",
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", transaction.Account)

	newTransaction := repository.NewTransaction(repository.REVERSAL,
		transaction.OrderId,
		transaction.Profile,
		transaction.Account,
		transaction.Instrument,
		transaction.InstrumentId,
		&req.Amount,
		transaction.Customer,
		transaction,
		nil,
	)

	if err := th.transactionStore.Add(c, newTransaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := bankApi.Reverse(c, newTransaction); err != nil {
		mess := err.Error()
		newTransaction.Declined(&mess)
	}

	if err, notfound := th.transactionStore.Update(c, newTransaction); err != nil {
		th.loggerFunc(c).Warningf("failed to update transaction: %v (notfound: %v)", err, notfound)
	}

	c.JSON(http.StatusOK, newTransaction)
}

func (th *transactionHandler) RefundHandler(c *gin.Context) {
	var req validators.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, transaction, bankApi := th.validate(c)
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !transaction.IsSuccess() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong state: %s", *transaction.Status),
		})
		return
	}

	if !transaction.IsAuth() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong type: %s", *transaction.Type),
		})
		return
	}

	if req.Amount > *transaction.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Incorrect amount: %d", req.Amount),
		})
		return
	}

	if req.Amount < *transaction.Amount && !*transaction.Account.PartialRefundEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Incorrect amount: %d, partial refund not allowed", req.Amount),
		})
		return
	}

	if !*transaction.Account.RefundEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Refund not allowed",
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", transaction.Account)

	newTransaction := repository.NewTransaction(repository.REFUND,
		transaction.OrderId,
		transaction.Profile,
		transaction.Account,
		transaction.Instrument,
		transaction.InstrumentId,
		&req.Amount,
		transaction.Customer,
		transaction,
		nil,
	)

	if err := th.transactionStore.Add(c, newTransaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := bankApi.Refund(c, newTransaction); err != nil {
		mess := err.Error()
		newTransaction.Declined(&mess)
	}

	if err, notfound := th.transactionStore.Update(c, newTransaction); err != nil {
		th.loggerFunc(c).Warningf("failed to update transaction: %v (notfound: %v)", err, notfound)
	}

	c.JSON(http.StatusOK, newTransaction)
}

func (th *transactionHandler) RebillHandler(c *gin.Context) {
	var req validators.RebillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, transaction, bankApi := th.validate(c)
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !transaction.IsSuccess() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong state: %s", *transaction.Status),
		})
		return
	}

	if !(transaction.IsAuth() || transaction.IsPreAuth()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong type: %s", *transaction.Type),
		})
		return
	}

	if !*transaction.Account.RebillEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Rebill not allowed",
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", transaction.Account)

	newTransaction := repository.NewTransaction(repository.REBILL,
		transaction.OrderId,
		transaction.Profile,
		transaction.Account,
		transaction.Instrument,
		transaction.InstrumentId,
		&req.Amount,
		transaction.Customer,
		transaction,
		nil,
	)

	if err := th.transactionStore.Add(c, newTransaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := bankApi.Rebill(c, newTransaction); err != nil {
		mess := err.Error()
		newTransaction.Declined(&mess)
	}

	if err, notfound := th.transactionStore.Update(c, newTransaction); err != nil {
		th.loggerFunc(c).Warningf("failed to update transaction: %v (notfound: %v)", err, notfound)
	}

	c.JSON(http.StatusOK, newTransaction)
}

func (th *transactionHandler) ConfirmPreAuthHandler(c *gin.Context) {
	var req validators.ConfirmPreAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, transaction, bankApi := th.validate(c)
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !transaction.IsSuccess() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong state: %s", *transaction.Status),
		})
		return
	}

	if !transaction.IsPreAuth() {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Transaction has wrong type: %s", *transaction.Type),
		})
		return
	}

	if req.Amount > *transaction.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Incorrect amount: %d", req.Amount),
		})
		return
	}

	if req.Amount < *transaction.Amount && !*transaction.Account.PartialConfirmEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Incorrect amount: %d, partial confirm not allowed", req.Amount),
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", transaction.Account)

	newTransaction := repository.NewTransaction(repository.CONFIRMAUTH,
		transaction.OrderId,
		transaction.Profile,
		transaction.Account,
		transaction.Instrument,
		transaction.InstrumentId,
		&req.Amount,
		transaction.Customer,
		transaction,
		nil,
	)

	if err := th.transactionStore.Add(c, newTransaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := bankApi.Confirm(c, newTransaction); err != nil {
		mess := err.Error()
		newTransaction.Declined(&mess)
	}

	if err, notfound := th.transactionStore.Update(c, newTransaction); err != nil {
		th.loggerFunc(c).Warningf("failed to update transaction: %v (notfound: %v)", err, notfound)
	}

	c.JSON(http.StatusOK, newTransaction)
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

	err, route := th.route(c, profile, instrument, th.cardStore, validators.CardAuthorizationInstrumentRequester, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, bankApi := plugins.BankApi(th.cfg, route.Account, route.Instrument, th.cardStore, th.sessionStore, th.transactionStore,th.loggerFunc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", route.Account)

	transaction := repository.NewTransaction(
		repository.AUTH,
		&req.OrderId,
		profile,
		route.Account,
		instrument,
		card.Id,
		&req.Amount,
		&req.Customer,
		nil,
		&req.BrowserInfo,
	)

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

func (th *transactionHandler) CardPreAuthorizeHandler(c *gin.Context) {
	var req validators.CardPreAuthorizeRequest
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
		validators.CardPreAuthorizationInstrumentRequester,
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

	err, route := th.route(c, profile, instrument, th.cardStore, validators.CardPreAuthorizationInstrumentRequester, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, bankApi := plugins.BankApi(th.cfg, route.Account, route.Instrument, th.cardStore, th.sessionStore, th.transactionStore, th.loggerFunc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	th.loggerFunc(c).Printf("using account: %v", route.Account)

	transaction := repository.NewTransaction(
		repository.PREAUTH,
		&req.OrderId,
		profile,
		route.Account,
		instrument,
		card.Id,
		&req.Amount,
		&req.Customer,
		nil,
		&req.BrowserInfo,
	)

	if err := th.transactionStore.Add(c, transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := bankApi.PreAuthorize(c, transaction, req); err != nil {
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
