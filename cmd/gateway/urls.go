package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin/binding"
	"github.com/durango/go-credit-card"
	"github.com/go-playground/validator/v10"
	"github.com/serg666/gateway/config"
	"github.com/serg666/gateway/middlewares"
	"github.com/serg666/gateway/handlers"
	"github.com/serg666/repository"
)

func MakeHandler(
	routeStore repository.RouteRepository,
	routerStore repository.RouterRepository,
	instrumentStore repository.InstrumentRepository,
	accountStore repository.AccountRepository,
	channelStore repository.ChannelRepository,
	profileStore repository.ProfileRepository,
	currencyStore repository.CurrencyRepository,
	cardStore repository.CardRepository,
	transactionStore repository.TransactionRepository,
	sessionStore repository.SessionRepository,
	cfg *config.Config,
	loggerFunc repository.LoggerFunc,
) *gin.Engine {
	routeHandler := handlers.NewRouteHandler(
		routeStore,
		profileStore,
		instrumentStore,
		accountStore,
		routerStore,
		currencyStore,
		channelStore,
		loggerFunc,
	)
	accountHandler := handlers.NewAccountHandler(accountStore, currencyStore, channelStore, loggerFunc)
	profileHandler := handlers.NewProfileHandler(profileStore, currencyStore, loggerFunc)
	currencyHandler := handlers.NewCurrencyHandler(currencyStore, loggerFunc)
	transactionHandler := handlers.NewTransactionHandler(
		routeStore,
		routerStore,
		instrumentStore,
		profileStore,
		accountStore,
		channelStore,
		currencyStore,
		cardStore,
		transactionStore,
		sessionStore,
		cfg,
		loggerFunc,
	)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("iscvv", func(fl validator.FieldLevel) bool {
			card := creditcard.Card{Cvv: fl.Field().String()}
			err := card.ValidateCVV()
			return err == nil
		})
		v.RegisterValidation("luhncheck", func(fl validator.FieldLevel) bool {
			card := creditcard.Card{Number: fl.Field().String()}
			return card.ValidateNumber()
		})
		v.RegisterValidation("notempty", func(fl validator.FieldLevel) bool {
			return fl.Field().Len() != 0
		})
	}

	gin.EnableJsonDecoderDisallowUnknownFields()

	handler := gin.New()

	handler.Use(
		requestid.New(),
		middlewares.Logger(loggerFunc),
		gin.Recovery(),
	)

	// @note: payment interface
	handler.POST("/profiles/:pid/transactions/authorize/card", transactionHandler.CardAuthorizeHandler)
	handler.POST("/profiles/:pid/transactions/preauthorize/card", transactionHandler.CardPreAuthorizeHandler)
	handler.POST("/profiles/:pid/transactions/:tid/confirm", transactionHandler.ConfirmPreAuthHandler)
	handler.POST("/profiles/:pid/transactions/:tid/reverse", transactionHandler.ReverseHandler)
	handler.POST("/profiles/:pid/transactions/:tid/refund", transactionHandler.RefundHandler)
	handler.POST("/profiles/:pid/transactions/:tid/rebill", transactionHandler.RebillHandler)
	handler.POST("/profiles/:pid/transactions/:tid/completemethodurl", transactionHandler.CompleteMethodUrlHandler)
	handler.POST("/profiles/:pid/transactions/:tid/processcres", transactionHandler.ProcessCresHandler)
	handler.POST("/profiles/:pid/transactions/:tid/processpares", transactionHandler.ProcessParesHandler)
	handler.GET("/profiles/:pid/transactions/:tid", transactionHandler.GetTransactionHandler)

	// @note: admin interface (should be moved to another web service)
	handler.POST("/routes", routeHandler.CreateRouteHandler)
	handler.GET("/routes", routeHandler.GetRoutesHandler)
	handler.GET("/routes/:id", routeHandler.GetRouteHandler)
	handler.DELETE("/routes/:id", routeHandler.DeleteRouteHandler)
	handler.PATCH("/routes/:id", routeHandler.PatchRouteHandler)

	handler.POST("/accounts", accountHandler.CreateAccountHandler)
	handler.GET("/accounts", accountHandler.GetAccountsHandler)
	handler.GET("/accounts/:id", accountHandler.GetAccountHandler)
	handler.DELETE("/accounts/:id", accountHandler.DeleteAccountHandler)
	handler.PATCH("/accounts/:id", accountHandler.PatchAccountHandler)

	handler.POST("/profiles", profileHandler.CreateProfileHandler)
	handler.GET("/profiles", profileHandler.GetProfilesHandler)
	handler.DELETE("/profiles/:pid", profileHandler.DeleteProfileHandler)
	handler.GET("/profiles/:pid", profileHandler.GetProfileHandler)
	handler.PATCH("/profiles/:pid", profileHandler.PatchProfileHandler)

	handler.POST("/currencies", currencyHandler.CreateCurrencyHandler)
	handler.GET("/currencies", currencyHandler.GetCurrenciesHandler)
	handler.DELETE("/currencies/:id", currencyHandler.DeleteCurrencyHandler)
	handler.GET("/currencies/:id", currencyHandler.GetCurrencyHandler)
	handler.PATCH("/currencies/:id", currencyHandler.PatchCurrencyHandler)

	return handler
}
