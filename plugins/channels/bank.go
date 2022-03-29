package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type BankChannel interface {
	Authorize(c *gin.Context, transaction *repository.Transaction, request interface{}) error
	PreAuthorize(c *gin.Context)
	Confirm(c *gin.Context)
	Reverse(c *gin.Context)
	Refund(c *gin.Context)
	Complete3DS(c *gin.Context)
	CompleteMethodUrl(c *gin.Context, transaction *repository.Transaction, request interface{}) error
}

var BankChannelType int = 1
