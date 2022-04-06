package validators

import (
	"fmt"
	"github.com/serg666/gateway/plugins/instruments/card"
	"github.com/serg666/repository"
)

type CardAuthorizeRequest struct {
	OrderId            string                 `json:"order_id" binding:"required"`
	Amount             uint                   `json:"amount" binding:"required,min=1"`
	Customer           string                 `json:"customer" binding:"required"`
	Card               bankcard.Card          `json:"card" binding:"required"`
	ThreeDSVer2TermUrl string                 `json:"threedsver2termurl" binding:"required"`
	BrowserInfo        repository.BrowserInfo `json:"browser_info" binding:"required"`
}

func CardAuthorizationInstrumentRequester(request interface{}) (error, interface{}) {
	cardAuthorizeRequest, ok := request.(CardAuthorizeRequest)
	if !ok {
		return fmt.Errorf("card authorization request has wrong type"), nil
	}

	return nil, cardAuthorizeRequest.Card
}

type CardPreAuthorizeRequest struct {
	CardAuthorizeRequest
}

func CardPreAuthorizationInstrumentRequester(request interface{}) (error, interface{}) {
	cardPreAuthorizeRequest, ok := request.(CardPreAuthorizeRequest)
	if !ok {
		return fmt.Errorf("card preauthorization request has wrong type"), nil
	}

	return nil, cardPreAuthorizeRequest.Card
}

type CompleteMethodUrlRequest struct {
	Completed *bool `json:"completed" binding:"required"`
}

type ProcessCresRequest struct {
	Cres string `json:"cres" binding:"required"`
}

type ProcessParesRequest struct {
	Pares string `json:"pares" binding:"required"`
}

type ConfirmPreAuthRequest struct {
	Amount uint `json:"amount" binding:"required,min=1"`
}

type ReversalRequest struct {
	ConfirmPreAuthRequest
}

type RefundRequest struct {
	ConfirmPreAuthRequest
}

type RebillRequest struct {
	ConfirmPreAuthRequest
}
