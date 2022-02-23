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
	profileStore repository.ProfileRepository,
	currencyStore repository.CurrencyRepository,
	loggerFunc repository.LoggerFunc,
) *gin.Engine {
	profileHandler := handlers.NewProfileHandler(profileStore, currencyStore)
	currencyHandler := handlers.NewCurrencyHandler(currencyStore)

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
