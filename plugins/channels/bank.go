package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type BankChannel interface {
	Authorize(c *gin.Context, transaction *repository.Transaction, request interface{}) error
	PreAuthorize(c *gin.Context, transaction *repository.Transaction, request interface{}) error
	Confirm(c *gin.Context)
	Reverse(c *gin.Context)
	Refund(c *gin.Context)
	ProcessCres(c *gin.Context, transaction *repository.Transaction, cres string) error
	ProcessPares(c *gin.Context, transaction *repository.Transaction, pares string) error
	CompleteMethodUrl(c *gin.Context, transaction *repository.Transaction, completed bool) error
}

var BankChannelType int = 1
