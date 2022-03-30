package validators

import (
	"fmt"
	"github.com/serg666/gateway/plugins/instruments/card"
)

type BrowserInfo struct {
	UserAgent     string `json:"user_agent" binding:"required"`
	AcceptHeader  string `json:"accept_header" binding:"required"`
	ColorDepth    *int   `json:"color_depth" binding:"required"`
	IP            string `json:"ip" binding:"required"`
	Language      string `json:"language" binding:"required"`
	ScreenHeight  *int   `json:"screen_height" binding:"required"`
	ScreenWidth   *int   `json:"screen_width" binding:"required"`
	ScreenPrint   string `json:"screen_print" binding:"required"`
	TZ            *int   `json:"tz" binding:"required"`
	TimeZone      string `json:"time_zone" binding:"required"`
	JavaEnabled   *bool  `json:"java_enabled" binding:"required"`
	DeviceChannel string `json:"device_channel" binding:"required"`
}

type CardAuthorizeRequest struct {
	OrderId            string        `json:"order_id" binding:"required"`
	Amount             uint          `json:"amount" binding:"required,min=1"`
	Customer           string        `json:"customer" binding:"required"`
	Card               bankcard.Card `json:"card" binding:"required"`
	ThreeDSVer2TermUrl string        `json:"threedsver2termurl" binding:"required"`
	BrowserInfo        BrowserInfo   `json:"browser_info" binding:"required"`
}

func CardAuthorizationInstrumentRequester(request interface{}) (error, interface{}) {
	cardAuthorizeRequest, ok := request.(CardAuthorizeRequest)
	if !ok {
		return fmt.Errorf("card authorization request has wrong type"), nil
	}

	return nil, cardAuthorizeRequest.Card
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
