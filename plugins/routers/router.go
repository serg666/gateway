package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type Router interface {
	Route(c *gin.Context, channel *repository.Channel, account *repository.Account) (error, *repository.Channel, *repository.Account)
}
