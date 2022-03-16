package alfabank

import (
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"errors"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/validators"
	"github.com/serg666/gateway/client"
	"github.com/serg666/gateway/config"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/instruments/card"
	"github.com/serg666/gateway/plugins/channels"
	"github.com/serg666/repository"
)

var (
	Id  = 2
	Key = "alfabank"
	Registered = plugins.RegisterBankChannel(Id, Key, func(
		cfg             *config.Config,
		route           *repository.Route,
		instrumentStore interface{},
		logger          repository.LoggerFunc,
	) (error, channels.BankChannel) {
		if *route.Instrument.Id != bankcard.Id {
			return fmt.Errorf("alfabank channel not sutable for instrument <%d>", *route.Instrument.Id), nil
		}

		jsonbody, err := json.Marshal(route.Account.Settings)
		if err != nil {
			return fmt.Errorf("can not marshal alfabank account settings: %v", err), nil
		}

		d := json.NewDecoder(bytes.NewReader(jsonbody))
		d.DisallowUnknownFields()

		var abs AlfaBankSettings

		if err := d.Decode(&abs); err != nil {
			return fmt.Errorf("can not decode alfabank account settings: %v", err), nil
		}

		return nil, &AlfaBankChannel{
			cfg:             cfg,
			logger:          logger,
			instrumentStore: instrumentStore,
			settings:        &abs,
		}
	})
)

type AlfaBankSettings struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AlfaBankChannel struct {
	cfg             *config.Config
	logger          repository.LoggerFunc
	instrumentStore interface{}
	settings        *AlfaBankSettings
}

func (abc *AlfaBankChannel) makeRequest(c *gin.Context, method string, url string, data url.Values) (error, *map[string]interface{}) {
	data.Set("userName", abc.settings.Login)
	data.Set("password", abc.settings.Password)

	r, err := http.NewRequest(method, fmt.Sprintf("%s/%s", abc.cfg.Alfabank.Ecom.Url, url), strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("can not make new request: %v", err), nil
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Client.Do(r)
	if err != nil {
		return fmt.Errorf("can not do request: %v", err), nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("can not read body: %v", err), nil
	}

	abc.logger(c).Printf("response body: %s", string(body))

	var jsonResp map[string]interface{}
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return fmt.Errorf("can not unmarshal body: %v", err), nil
	}

	return nil, &jsonResp
}

func (abc *AlfaBankChannel) parseError(c *gin.Context, response *map[string]interface{}) (*string, *string) {
	errCode := "-1"
	errMess := "unknown error"

	if errorCode, ok := (*response)["errorCode"]; ok {
		switch rc := errorCode.(type) {
		case float64:
			errCode = strconv.FormatFloat(rc, 'f', -1, 32)
		case int:
			errCode = strconv.Itoa(rc)
		case string:
			errCode = rc
		default:
			abc.logger(c).Printf("rc type: %v (%T)", rc, rc)
		}
	}

	if errorMessage, ok := (*response)["errorMessage"]; ok {
		if mess, ok := errorMessage.(string); ok {
			errMess = mess
		}
	}

	return &errCode, &errMess
}

func (abc *AlfaBankChannel) Authorize(c *gin.Context, transaction *repository.Transaction, request interface{}) error {
	req, ok := request.(validators.CardAuthorizeRequest)
	if !ok {
		return fmt.Errorf("request has wrong type")
	}

	data := url.Values{}
	data.Set("orderNumber", strconv.Itoa(*transaction.Id))
	data.Set("currency", strconv.Itoa(*transaction.CurrencyConverted.NumericCode))
	data.Set("amount", strconv.Itoa(int(*transaction.AmountConverted)))
	data.Set("returnUrl", "1")

	err, jsonResp := abc.makeRequest(c, "POST", "ab/rest/register.do", data)
	if err != nil {
		return fmt.Errorf("can not make register order request: %v", err)
	}

	if orderId, ok := (*jsonResp)["orderId"]; ok {
		if remoteId, ok := orderId.(string); ok {
			transaction.RemoteId = &remoteId

			data := url.Values{}
			data.Set("MDORDER", remoteId)
			data.Set("$PAN", req.Card.Pan)
			data.Set("$CVC", req.Card.Cvv)
			data.Set("YYYY", strconv.Itoa(req.Card.ExpDate.Year()))
			data.Set("MM", fmt.Sprintf("%02d", int(req.Card.ExpDate.Month())))
			data.Set("TEXT", req.Card.Holder)

			err, jsonResp := abc.makeRequest(c, "POST", "ab/rest/paymentorder.do", data)
			if err != nil {
				return fmt.Errorf("can not make payment order request: %v", err)
			}

			rc, mess := abc.parseError(c, jsonResp)
			transaction.ResponseCode = rc
			if *rc != "0" {
				return errors.New(*mess)
			}
			// @note: success request
			// @todo: process transaction
		} else {
			return errors.New("orderId has wrong type")
		}
	} else {
		rc, mess := abc.parseError(c, jsonResp)

		transaction.ResponseCode = rc
		return errors.New(*mess)
	}

	return nil
}

func (abc *AlfaBankChannel) PreAuthorize(c *gin.Context) {
	abc.logger(c).Print("preauthorize")
}

func (abc *AlfaBankChannel) Confirm(c *gin.Context) {
	abc.logger(c).Print("confirm")
}

func (abc *AlfaBankChannel) Reverse(c *gin.Context) {
	abc.logger(c).Print("reverse")
}

func (abc *AlfaBankChannel) Refund(c *gin.Context) {
	abc.logger(c).Print("refund")
}

func (abc *AlfaBankChannel) Complete3DS(c *gin.Context) {
	abc.logger(c).Print("complete3ds")
}
