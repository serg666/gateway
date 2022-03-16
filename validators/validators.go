package validators

import (
	"fmt"
	"github.com/serg666/gateway/plugins/instruments/card"
)

type CardAuthorizeRequest struct {
	OrderId string        `json:"order_id" binding:"required"`
	Amount  uint          `json:"amount" binding:"required,min=1"`
	Card    bankcard.Card `json:"card" binding:"required"`
}

func CardAuthorizationInstrumentRequester(request interface{}) (error, interface{}) {
	cardAuthorizeRequest, ok := request.(CardAuthorizeRequest)
	if !ok {
		return fmt.Errorf("card authorization request has wrong type"), nil
	}

	return nil, cardAuthorizeRequest.Card
}
