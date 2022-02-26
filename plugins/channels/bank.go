package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/plugins/instruments"
)

type BankChannel interface {
	Authorize (c *gin.Context, instrument instruments.PaymentInstrument)
	PreAuthorize(c *gin.Context)
	Confirm(c *gin.Context)
	Reverse(c *gin.Context)
	Refund(c *gin.Context)
	Complete3DS(c *gin.Context)
}

var BankChannelType int = 1
