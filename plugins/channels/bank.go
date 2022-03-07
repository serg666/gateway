package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type BankChannel interface {
	SutableForInstrument(instrument *repository.Instrument) bool
	Authorize (c *gin.Context, instrument *repository.Instrument) (error, error)
	PreAuthorize(c *gin.Context)
	Confirm(c *gin.Context)
	Reverse(c *gin.Context)
	Refund(c *gin.Context)
	Complete3DS(c *gin.Context)
}

var BankChannelType int = 1
