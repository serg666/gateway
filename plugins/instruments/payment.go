package instruments

import (
	"github.com/gin-gonic/gin"
)

type PaymentInstrument interface {
	FromContext(c *gin.Context) (error, interface{})
}
