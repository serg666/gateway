package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type BankChannel interface {
	DecodeSettings(settings *repository.AccountSettings) error
	SutableForInstrument(instrument *repository.Instrument) bool
	Authorize (c *gin.Context, transaction *repository.Transaction, instrumentInstance interface{})
	PreAuthorize(c *gin.Context)
	Confirm(c *gin.Context)
	Reverse(c *gin.Context)
	Refund(c *gin.Context)
	Complete3DS(c *gin.Context)
}

var BankChannelType int = 1
