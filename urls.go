package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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
		loggerFunc,
	)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
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
	handler.POST("/profiles/:id/transactions/authorize/card", transactionHandler.CardAuthorizeHandler)

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
	handler.DELETE("/profiles/:id", profileHandler.DeleteProfileHandler)
	handler.GET("/profiles/:id", profileHandler.GetProfileHandler)
	handler.PATCH("/profiles/:id", profileHandler.PatchProfileHandler)

	handler.POST("/currencies", currencyHandler.CreateCurrencyHandler)
	handler.GET("/currencies", currencyHandler.GetCurrenciesHandler)
	handler.DELETE("/currencies/:id", currencyHandler.DeleteCurrencyHandler)
	handler.GET("/currencies/:id", currencyHandler.GetCurrencyHandler)
	handler.PATCH("/currencies/:id", currencyHandler.PatchCurrencyHandler)

	return handler
}
