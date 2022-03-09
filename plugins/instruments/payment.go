package instruments

import (
	"github.com/gin-gonic/gin"
)

type PaymentInstrument interface {
	FromRequest(c *gin.Context, request interface{}, instrumentStore interface{}) (error, interface{})
}
