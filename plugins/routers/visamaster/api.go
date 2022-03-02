package visamaster

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/routers"
	"github.com/serg666/repository"
)

var (
	Id  = 1
	Key = "visamaster"
	Registered = plugins.RegisterRouter(Id, Key, func(
		settings *repository.RouterSettings,
		logger repository.LoggerFunc,
	) routers.Router {
		return &VisaMasterRouter{
			logger:   logger,
			settings: settings,
		}
	})
)

type VisaMasterRouter struct {
	logger   repository.LoggerFunc
	settings *repository.RouterSettings
}

func (vmr *VisaMasterRouter) Route(c *gin.Context, account *repository.Account) error {
	vmr.logger(c).Print("some message")
	return nil
}
