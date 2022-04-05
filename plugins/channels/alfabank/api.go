package alfabank

import (
	"fmt"
	"regexp"
	"bytes"
	"strconv"
	"strings"
	"errors"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/mileusna/useragent"
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
		cfg              *config.Config,
		account          *repository.Account,
		instrument       *repository.Instrument,
		sessionStore     repository.SessionRepository,
		transactionStore repository.TransactionRepository,
		logger           repository.LoggerFunc,
	) (error, channels.BankChannel) {
		if *instrument.Id != bankcard.Id {
			return fmt.Errorf("alfabank channel not sutable for instrument <%d>", *instrument.Id), nil
		}

		jsonbody, err := json.Marshal(account.Settings)
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
			cfg:              cfg,
			logger:           logger,
			sessionStore:     sessionStore,
			transactionStore: transactionStore,
			settings:         &abs,
		}
	})
)

type ClientInfo struct {
	UserAgent             string `json:"userAgent"`
	OS                    string `json:"os"`
	OSVersion             string `json:"osversion"`
	Device                string `json:"device"`
	Mobile                bool   `json:"mobile"`
	ScreenPrint           string `json:"screenPrint"`
	ColorDepth            int    `json:"colorDepth"`
	ScreenHeight          int    `json:"screenHeight"`
	ScreenWidth           int    `json:"screenWidth"`
	JavaEnabled           bool   `json:"javaEnabled"`
	JavascriptEnabled     bool   `json:"javascriptEnabled"`
	BrowserLanguage       string `json:"browserLanguage"`
	BrowserTimeZone       string `json:"browserTimeZone"`
	BrowserTimeZoneOffset int    `json:"browserTimeZoneOffset"`
}

type AlfaBankSettings struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AlfaBankChannel struct {
	cfg              *config.Config
	logger           repository.LoggerFunc
	sessionStore     repository.SessionRepository
	transactionStore repository.TransactionRepository
	settings         *AlfaBankSettings
}

func (abc *AlfaBankChannel) maskParams(data string) string {
	sampleRegexp := regexp.MustCompile(`PAN=[^&]+([^&]{4})`)
	result := sampleRegexp.ReplaceAllString(data, "PAN=******$1")
	sampleRegexp = regexp.MustCompile(`CVC=[^&]+`)
	result = sampleRegexp.ReplaceAllString(result, "CVC=***")
	sampleRegexp = regexp.MustCompile(`password=[^&]+`)
	result = sampleRegexp.ReplaceAllString(result, "password=******")
	return result
}

func (abc *AlfaBankChannel) makeRequest(
	c *gin.Context,
	method string,
	url string,
	data string,
) (error, *map[string]interface{}) {
	uri := fmt.Sprintf("%s/%s", abc.cfg.Alfabank.Ecom.Url, url)
	abc.logger(c).Printf("Requesting: %s", uri)
	abc.logger(c).Printf("Params: %s", abc.maskParams(data))
	r, err := http.NewRequest(method, uri, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("can not make new request: %v", err), nil
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data)))

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

func (abc *AlfaBankChannel) putBrowserInfo(
	c *gin.Context,
	browserInfo validators.BrowserInfo,
	serverUrl *string,
	transId *string,
) error {
	abc.logger(c).Debugf("putting browserInfo: %v", browserInfo)

	if serverUrl == nil {
		return errors.New("server URL is null")
	}

	abc.logger(c).Debugf("requesting: %s", *serverUrl)
	data := url.Values{}

	r, err := http.NewRequest("POST", *serverUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("can not make new request: %v", err)
	}

	r.Header.Add("User-Agent", browserInfo.UserAgent)
	r.Header.Add("Accept", browserInfo.AcceptHeader)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Client.Do(r)
	if err != nil {
		return fmt.Errorf("can not do request: %v", err)
	}
	defer res.Body.Close()

	abc.logger(c).Debugf("response code: %d", res.StatusCode)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("can not read response body: %v", err)
	}

	abc.logger(c).Debugf("response body: %s", string(body))
	re := regexp.MustCompile(`(https?://[^\"\s>]+)`)
	links := re.FindAll(body, -1)
	abc.logger(c).Printf("links: %q", links)
	if len(links) == 0 {
		return errors.New("no links has been found")
	}

	clientUrl := string(links[len(links)-1])
	abc.logger(c).Printf("link: %s", clientUrl)
	abc.logger(c).Debugf("requesting: %s", clientUrl)
	ua := ua.Parse(browserInfo.UserAgent)
	clientInfo := ClientInfo{
		UserAgent: ua.String,
		OS: ua.OS,
		OSVersion: ua.OSVersion,
		Device: ua.Device,
		Mobile: ua.Mobile,
		ScreenPrint: browserInfo.ScreenPrint,
		ColorDepth: *browserInfo.ColorDepth,
		ScreenHeight: *browserInfo.ScreenHeight,
		ScreenWidth: *browserInfo.ScreenWidth,
		JavaEnabled: *browserInfo.JavaEnabled,
		JavascriptEnabled: true,
		BrowserLanguage: browserInfo.Language,
		BrowserTimeZone: browserInfo.TimeZone,
		BrowserTimeZoneOffset: *browserInfo.TZ,
	}

	jsonbody, err := json.Marshal(clientInfo)
	if err != nil {
		return fmt.Errorf("can not marshal client info: %v", err)
	}

	abc.logger(c).Printf("json body: %s", string(jsonbody))
	postData := url.Values{}
	postData.Set("threeDSServerTransID", *transId)
	postData.Set("clientInfo", string(jsonbody))
	nr, err := http.NewRequest("POST", clientUrl, strings.NewReader(postData.Encode()))
	if err != nil {
		return fmt.Errorf("can not make client info request: %v", err)
	}

	nr.Header.Add("User-Agent", browserInfo.UserAgent)
	nr.Header.Add("Accept", browserInfo.AcceptHeader)
	nr.Header.Add("Referer", *serverUrl)
	nr.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	nr.Header.Add("Content-Length", strconv.Itoa(len(postData.Encode())))

	nres, err := client.Client.Do(nr)
	if err != nil {
		return fmt.Errorf("can not do client info request: %v", err)
	}
	defer nres.Body.Close()

	abc.logger(c).Debugf("client response code: %d", nres.StatusCode)
	nbody, err := ioutil.ReadAll(nres.Body)
	if err != nil {
		return fmt.Errorf("can not read client info response body: %v", err)
	}

	abc.logger(c).Debugf("client response body: %s", string(nbody))

	return nil
}

func (abc *AlfaBankChannel) isCREQ(c *gin.Context, response *map[string]interface{}) (bool, *string, *string) {
	var acsURL *string
	var cReq  *string

	if acsUrl, ok := (*response)["acsUrl"]; ok {
		if acs, ok := acsUrl.(string); ok {
			acsURL = &acs
		} else {
			return false, nil, nil
		}
	} else {
		return false, nil, nil
	}

	if packedCReq, ok := (*response)["packedCReq"]; ok {
		if creq, ok := packedCReq.(string); ok {
			cReq = &creq
		} else {
			return false, nil, nil
		}
	} else {
		return false, nil, nil
	}

	return true, acsURL, cReq
}

func (abc *AlfaBankChannel) is3DS20(
	c *gin.Context,
	response *map[string]interface{},
) (bool, *string, *string, *string, *string) {
	var threeDSServerTransId *string
	var threeDSMethodURLServer *string
	var threeDSMethodURL *string
	var threeDSMethodDataPacked *string
	var is3DS bool

	is3DS = false
	if is3DSVer2, ok := (*response)["is3DSVer2"]; ok {
		switch is3DS2 := is3DSVer2.(type) {
		case bool:
			is3DS = is3DS2
		case string:
			is3DS = is3DS2 == "true"
		default:
			abc.logger(c).Printf("is3DS2 type: %v (%T)", is3DS2, is3DS2)
		}
	}

	if serverTransId, ok := (*response)["threeDSServerTransId"]; ok {
		if transId, ok := serverTransId.(string); ok {
			threeDSServerTransId = &transId
		}
	}

	if methodURLServer, ok := (*response)["threeDSMethodURLServer"]; ok {
		if serverURL, ok := methodURLServer.(string); ok {
			threeDSMethodURLServer = &serverURL
		}
	}

	if methodURL, ok := (*response)["threeDSMethodURL"]; ok {
		if mURL, ok := methodURL.(string); ok {
			threeDSMethodURL = &mURL
		}
	}

	if methodDataPacked, ok := (*response)["threeDSMethodDataPacked"]; ok {
		if methodData, ok := methodDataPacked.(string); ok {
			threeDSMethodDataPacked = &methodData
		}
	}

	return is3DS, threeDSServerTransId, threeDSMethodURLServer, threeDSMethodURL, threeDSMethodDataPacked
}

func (abc *AlfaBankChannel) is3DS10(c *gin.Context, response *map[string]interface{}) (bool, *string, *string) {
	var acsURL *string
	var paREQ  *string

	if acsUrl, ok := (*response)["acsUrl"]; ok {
		if acs, ok := acsUrl.(string); ok {
			acsURL = &acs
		} else {
			return false, nil, nil
		}
	} else {
		return false, nil, nil
	}

	if paReq, ok := (*response)["paReq"]; ok {
		if pareq, ok := paReq.(string); ok {
			paREQ = &pareq
		} else {
			return false, nil, nil
		}
	} else {
		return false, nil, nil
	}

	return true, acsURL, paREQ
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

func (abc *AlfaBankChannel) parseState(
	c *gin.Context,
	response *map[string]interface{},
) (string, *string, *string, *string, *string, *string) {
	var actionCode  *string
	var actionCodeDescription  *string
	var rrn  *string
	var authcode  *string
	var bindingId *string

	state := "-1"

	if orderStatus, ok := (*response)["orderStatus"]; ok {
		switch status := orderStatus.(type) {
		case float64:
			state = strconv.FormatFloat(status, 'f', -1, 32)
		case int:
			state = strconv.Itoa(status)
		case string:
			state = status
		default:
			abc.logger(c).Warningf("orderStatus type: %v (%T)", status, status)
		}
	}

	if ac, ok := (*response)["actionCode"]; ok {
		switch rc := ac.(type) {
		case float64:
			errCode := strconv.FormatFloat(rc, 'f', -1, 32)
			actionCode = &errCode
		case int:
			errCode := strconv.Itoa(rc)
			actionCode = &errCode
		case string:
			actionCode = &rc
		default:
			abc.logger(c).Warningf("actionCode type: %v (%T)", rc, rc)
		}
	}

	if actionDescr, ok := (*response)["actionCodeDescription"]; ok {
		switch mess := actionDescr.(type) {
		case string:
			actionCodeDescription = &mess
		default:
			abc.logger(c).Warningf("actionCodeDescription type: %v (%T)", mess, mess)
		}
	}

	if authRefNum, ok := (*response)["authRefNum"]; ok {
		switch refNum := authRefNum.(type) {
		case float64:
			ref := strconv.FormatFloat(refNum, 'f', -1, 32)
			rrn = &ref
		case int:
			ref := strconv.Itoa(refNum)
			rrn = &ref
		case string:
			rrn = &refNum
		default:
			abc.logger(c).Warningf("authRefNum type: %v (%T)", refNum, refNum)
		}
	}

	if cardAuthInfo, ok := (*response)["cardAuthInfo"]; ok {
		if authInfo, ok := cardAuthInfo.(map[string]interface{}); ok {
			if approvalCode, ok := authInfo["approvalCode"]; ok {
				switch approval := approvalCode.(type) {
				case float64:
					authCode := strconv.FormatFloat(approval, 'f', -1, 32)
					authcode = &authCode
				case int:
					authCode := strconv.Itoa(approval)
					authcode = &authCode
				case string:
					authcode = &approval
				default:
					abc.logger(c).Warningf("approvalCode type: %v (%T)", approval, approval)
				}
			}
		}
	}

	if bindingInfo, ok := (*response)["bindingInfo"]; ok {
		if binding, ok := bindingInfo.(map[string]interface{}); ok {
			if linkId, ok := binding["bindingId"]; ok {
				switch bid := linkId.(type) {
				case float64:
					link := strconv.FormatFloat(bid, 'f', -1, 32)
					bindingId = &link
				case int:
					link := strconv.Itoa(bid)
					bindingId = &link
				case string:
					bindingId = &bid
				default:
					abc.logger(c).Warningf("bindingId type: %v (%T)", bid, bid)
				}
			}
		}
	}

	return state, actionCode, actionCodeDescription, rrn, authcode, bindingId
}

func (abc *AlfaBankChannel) updateTransaction(c *gin.Context, transaction *repository.Transaction) {
	if !transaction.InFinalState() {
		data := url.Values{}
		data.Set("userName", abc.settings.Login)
		data.Set("password", abc.settings.Password)
		data.Set("orderId", *transaction.RemoteId)
		if err, jsonResp := abc.makeRequest(c, "POST", "ab/rest/getOrderStatusExtended.do", data.Encode()); err == nil {
			state, actionCode, actionCodeDescr, rrn, authCode, bindingId := abc.parseState(c, jsonResp)

			transaction.ResponseCode = actionCode
			transaction.AuthCode = authCode
			transaction.RRN = rrn
			if bindingId != nil {
				transaction.AdditionalData = &repository.AdditionalData{
					"bindingId": *bindingId,
				}
			}

			switch state {
			case
				"1",
				"2":
				transaction.Success()
			case "6":
				transaction.Declined(actionCodeDescr)
			}
		} else {
			abc.logger(c).Warningf("can not make get order status request: %v", err)
		}
	} else {
		abc.logger(c).Warningf("transaction <%d> has already got final state: %s", *transaction.Id, *transaction.Status)
	}
}

func (abc *AlfaBankChannel) processCard(
	c *gin.Context,
	transaction *repository.Transaction,
	card bankcard.Card,
	termUrl string,
	browserInfo validators.BrowserInfo,
	registerMethod string,
) error {
	data := url.Values{}
	data.Set("userName", abc.settings.Login)
	data.Set("password", abc.settings.Password)
	data.Set("orderNumber", strconv.Itoa(*transaction.Id))
	data.Set("currency", strconv.Itoa(*transaction.CurrencyConverted.NumericCode))
	data.Set("amount", strconv.Itoa(int(*transaction.AmountConverted)))
	data.Set("clientId", *transaction.Customer)
	// @note: we do not use return url at all
	data.Set("returnUrl", "1")

	err, jsonResp := abc.makeRequest(c, "POST", fmt.Sprintf("ab/rest/%s", registerMethod), data.Encode())
	if err != nil {
		return fmt.Errorf("can not make register order request: %v", err)
	}

	if orderId, ok := (*jsonResp)["orderId"]; ok {
		if remoteId, ok := orderId.(string); ok {
			transaction.RemoteId = &remoteId

			data := url.Values{}
			data.Set("userName", abc.settings.Login)
			data.Set("password", abc.settings.Password)
			data.Set("MDORDER", remoteId)
			data.Set("$PAN", card.Pan)
			data.Set("$CVC", card.Cvv)
			data.Set("YYYY", strconv.Itoa(card.ExpDate.Year()))
			data.Set("MM", fmt.Sprintf("%02d", int(card.ExpDate.Month())))
			data.Set("TEXT", card.Holder)
			data.Set("threeDSVer2FinishUrl", termUrl)

			if err, jsonResp := abc.makeRequest(c, "POST", "ab/rest/paymentorder.do", data.Encode()); err == nil {
				if is3ds20, transId, serverUrl, methodUrl, methodData := abc.is3DS20(c, jsonResp); is3ds20 {
					//3ds20
					if err := abc.putBrowserInfo(c, browserInfo, serverUrl, transId); err != nil {
						abc.logger(c).Warningf("can not put browser info: %v", err)
					}
					data.Set("threeDSServerTransId", *transId)
					if err := abc.sessionStore.Add(c, repository.NewSession(
						fmt.Sprintf("3ds20session_%d", *transaction.Id),
						repository.SessionData{
							"query": data.Encode(),
							"threeDSServerTransId": *transId,
						},
					)); err != nil {
						return fmt.Errorf("can not add 3ds20 session: %v", err)
					}
					if methodUrl != nil {
						transaction.ThreeDSMethodUrl = &repository.ThreeDSMethodUrl{
							MethodUrl: methodUrl,
							ThreeDSMethodData: methodData,
						}
						transaction.WaitMethodUrl()
					} else {
						if err, jsonResp := abc.makeRequest(c, "POST", "ab/rest/paymentorder.do", data.Encode()); err == nil {
							if iscreq, acs, creq := abc.isCREQ(c, jsonResp); iscreq {
								transaction.ThreeDSecure20 = &repository.ThreeDSecure20{
									AcsUrl: acs,
									Creq: creq,
								}
								transaction.Wait3DS()
							} else {
								abc.updateTransaction(c, transaction)
							}
						} else {
							abc.updateTransaction(c, transaction)
						}
					}
				} else {
					if is3ds10, acs, pareq := abc.is3DS10(c, jsonResp); is3ds10 {
						transaction.ThreeDSecure10 = &repository.ThreeDSecure10{
							AcsUrl: acs,
							PaReq: pareq,
						}
						transaction.Wait3DS()
					} else {
						abc.updateTransaction(c, transaction)
					}
				}
			} else {
				abc.logger(c).Warningf("can not make payment order request: %v", err)
				abc.updateTransaction(c, transaction)
			}
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

func (abc *AlfaBankChannel) Authorize(c *gin.Context, transaction *repository.Transaction, request interface{}) error {
	req, ok := request.(validators.CardAuthorizeRequest)
	if !ok {
		return fmt.Errorf("request has wrong type")
	}

	return abc.processCard(c, transaction, req.Card, req.ThreeDSVer2TermUrl, req.BrowserInfo, "register.do")
}

func (abc *AlfaBankChannel) PreAuthorize(c *gin.Context, transaction *repository.Transaction, request interface{}) error {
	req, ok := request.(validators.CardPreAuthorizeRequest)
	if !ok {
		return fmt.Errorf("request has wrong type")
	}

	return abc.processCard(c, transaction, req.Card, req.ThreeDSVer2TermUrl, req.BrowserInfo, "registerPreAuth.do")
}

func (abc *AlfaBankChannel) Confirm(c *gin.Context, transaction *repository.Transaction) error {
	err, successRefsTotal := abc.transactionStore.TypeTurnOver(c, repository.NewTransactionSpecificationByReferenceIdAndStatus(
		*transaction.Reference.Id,
		repository.SUCCESS,
	))
	abc.logger(c).Printf("successRefsTotal: %v (%v)", successRefsTotal, err)

	if err != nil {
		return fmt.Errorf("can not get success references total: %v", err)
	}

	var reversals uint
	var confirms uint

	if confirmTurnOver, ok := (*successRefsTotal)["confirmauth"]; ok {
		confirms = confirmTurnOver.Sum
	}

	if confirms > 0 {
		// @note: alfabank allow only one partial confirm
		return fmt.Errorf("transaction has already confirmed: %d", confirms)
	}

	if reversalTurnOver, ok := (*successRefsTotal)["reversal"]; ok {
		reversals = reversalTurnOver.Sum
	}

	availableAmount := *transaction.Reference.Amount - reversals - confirms
	if availableAmount < *transaction.Amount {
		return fmt.Errorf("incorrect amount: %d", *transaction.Amount)
	}

	data := url.Values{}
	data.Set("userName", abc.settings.Login)
	data.Set("password", abc.settings.Password)
	data.Set("orderId", *transaction.Reference.RemoteId)
	data.Set("amount", strconv.Itoa(int(*transaction.AmountConverted)))
	transaction.RemoteId = transaction.Reference.RemoteId

	err, jsonResp := abc.makeRequest(c, "POST", "ab/rest/deposit.do", data.Encode())
	if err != nil {
		return fmt.Errorf("can not make deposit request: %v", err)
	}

	rc, mess := abc.parseError(c, jsonResp)
	if *rc != "0" {
		transaction.ResponseCode = rc
		return errors.New(*mess)
	}

	abc.updateTransaction(c, transaction)

	return nil
}

func (abc *AlfaBankChannel) Reverse(c *gin.Context) {
	abc.logger(c).Print("reverse")
}

func (abc *AlfaBankChannel) Refund(c *gin.Context) {
	abc.logger(c).Print("refund")
}

func (abc *AlfaBankChannel) ProcessPares(c *gin.Context, transaction *repository.Transaction, pares string) error {
	abc.logger(c).Printf("Pares: %v", pares)

	data := url.Values{}
	data.Set("PaRes", pares)
	data.Set("MD", *transaction.RemoteId)

	abc.makeRequest(c, "POST", "ab/rest/finish3ds.do", data.Encode())
	abc.updateTransaction(c, transaction)

	return nil
}

func (abc *AlfaBankChannel) ProcessCres(c *gin.Context, transaction *repository.Transaction, cres string) error {
	sessionKey := fmt.Sprintf("3ds20session_%d", *transaction.Id)
	err, _, sessions := abc.sessionStore.Query(c, repository.NewSessionSpecificationByKey(sessionKey))

	if err != nil {
		return fmt.Errorf("failed to query session store: %v", err)
	}

	if len(sessions) == 0 {
		return fmt.Errorf("session with key %s not found", sessionKey)
	}

	session := sessions[0]
	sessionData := session.Data

	threeDSServerTransId, ok := (*sessionData)["threeDSServerTransId"]
	if !ok {
		return errors.New("session data has not threeDSServerTransId")
	}

	tDsTransId, ok := threeDSServerTransId.(string)
	if !ok {
		return errors.New("session data threeDSServerTransId has wrong type")
	}

	data := url.Values{}
	data.Set("userName", abc.settings.Login)
	data.Set("password", abc.settings.Password)
	data.Set("tDsTransId", tDsTransId)

	abc.makeRequest(c, "POST", "ab/rest/finish3dsVer2.do", data.Encode())
	abc.updateTransaction(c, transaction)

	return nil
}

func (abc *AlfaBankChannel) CompleteMethodUrl(c *gin.Context, transaction *repository.Transaction, completed bool) error {
	abc.logger(c).Printf("completed: %v", completed)

	sessionKey := fmt.Sprintf("3ds20session_%d", *transaction.Id)
	err, _, sessions := abc.sessionStore.Query(c, repository.NewSessionSpecificationByKey(sessionKey))

	if err != nil {
		return fmt.Errorf("failed to query session store: %v", err)
	}

	if len(sessions) == 0 {
		return fmt.Errorf("session with key %s not found", sessionKey)
	}

	session := sessions[0]
	sessionData := session.Data

	query, ok := (*sessionData)["query"]
	if !ok {
		return errors.New("session data has not query")
	}

	qwr, ok := query.(string)
	if !ok {
		return errors.New("session data query has wrong type")
	}

	if err, jsonResp := abc.makeRequest(c, "POST", "ab/rest/paymentorder.do", qwr); err == nil {
		if iscreq, acs, creq := abc.isCREQ(c, jsonResp); iscreq {
			transaction.ThreeDSecure20 = &repository.ThreeDSecure20{
				AcsUrl: acs,
				Creq: creq,
			}
			transaction.Wait3DS()
		} else {
			abc.updateTransaction(c, transaction)
		}
	} else {
		abc.updateTransaction(c, transaction)
	}

	return nil
}
