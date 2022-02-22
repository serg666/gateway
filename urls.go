package main

import (
	"fmt"
	"github.com/wk8/go-ordered-map"
	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/serg666/gateway/middlewares"
	"github.com/serg666/gateway/handlers"
	"github.com/serg666/gateway/config"
	"github.com/serg666/repository"
)

func MakeHandler(cfg *config.Config) (*gin.Engine, error) {
	loggerFunc := func (c interface{}) logrus.FieldLogger {
		return cfg.LogRusLogger(c.(*gin.Context))
	}

	pgPool, err := repository.MakePgPoolFromDSN(cfg.Databases.Default.Dsn)
	if err != nil {
		return nil, fmt.Errorf("Can not make pg pool due to: %v", err)
	}

	profileStore := repository.NewOrderedMapProfileStore(orderedmap.New(), loggerFunc)
	//currencyStore := repository.NewOrderedMapCurrencyStore(orderedmap.New(), loggerFunc)
	currencyStore := repository.NewPGPoolCurrencyStore(pgPool, loggerFunc)

	profileHandler := handlers.NewProfileHandler(cfg, profileStore, currencyStore)
	currencyHandler := handlers.NewCurrencyHandler(cfg, currencyStore)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notempty", func(fl validator.FieldLevel) bool {
			return fl.Field().Len() != 0
		})
	}

	gin.EnableJsonDecoderDisallowUnknownFields()

	handler := gin.New()

	handler.Use(
		requestid.New(),
		middlewares.Logger(cfg),
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

	return handler, nil
}
