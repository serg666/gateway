package bankcard

import (
	"fmt"
	"time"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/instruments"
	"github.com/serg666/repository"
)

var (
	Id  = 1
	Key = "card"
	Registered = plugins.RegisterPaymentInstrument(Id, Key, func(
		instrumentStore interface{},
		logger repository.LoggerFunc,
	) instruments.PaymentInstrument {
		return &BankCard{
			instrumentStore: instrumentStore,
			logger:          logger,
		}
	})
)

type ExpDate struct {
	time.Time
}

func (t *ExpDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	t.Time, err = time.Parse("2006-01-02", s)
	return
}

type Card struct {
	Pan     string  `json:"pan" binding:"required,luhncheck"`
	Cvv     string  `json:"cvv" binding:"required,iscvv"`
	ExpDate ExpDate `json:"expire" binding:"required"`
	Holder  string  `json:"holder" binding:"required"`
}

type CardAuthorizeRequest struct {
	OrderId string `json:"order_id" binding:"required"`
	Amount  uint   `json:"amount" binding:"required,min=1"`
	Card    Card   `json:"card" binding:"required"`
}

type BankCard struct {
	instrumentStore interface{}
	logger          repository.LoggerFunc
}

func (bc *BankCard) FromRequest(c *gin.Context, request interface{}) (error, interface{}) {
	cardAuthorizeRequest, ok := request.(CardAuthorizeRequest)
	if !ok {
		return fmt.Errorf("request has wrong type"), nil
	}

	cardStore, ok := bc.instrumentStore.(repository.CardRepository)
	if !ok {
		return fmt.Errorf("instrumentStore has wrong type"), nil
	}

	pan := repository.PAN(cardAuthorizeRequest.Card.Pan)

	err, _, cards := cardStore.Query(c, repository.NewCardSpecificationByPAN(pan))
	if err != nil {
		return fmt.Errorf("card store quering failed: %v", err), nil
	}

	if len(cards) > 0 {
		return nil, cards[0]
	}

	card := &repository.Card{
		PAN:     &pan,
		ExpDate: &cardAuthorizeRequest.Card.ExpDate.Time,
		Holder:  &cardAuthorizeRequest.Card.Holder,
	}

	if err := cardStore.Add(c, card); err != nil {
		return fmt.Errorf("card store adding failed: %v", err), nil
	}

	return nil, card
}
