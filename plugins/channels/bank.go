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
	ProcessCres(c *gin.Context, transaction *repository.Transaction, request interface{}) error
	ProcessPares(c *gin.Context, transaction *repository.Transaction, request interface{}) error
	CompleteMethodUrl(c *gin.Context, transaction *repository.Transaction, request interface{}) error
}

var BankChannelType int = 1
