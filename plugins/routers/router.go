package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type Router interface {
	SutableForInstrument(instrument *repository.Instrument) bool
	Route(c *gin.Context, route *repository.Route, instrumentInstance interface{}) error
}
