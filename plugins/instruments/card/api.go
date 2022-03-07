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
	Registered = plugins.RegisterPaymentInstrument(Id, Key, func(logger repository.LoggerFunc) instruments.PaymentInstrument {
		return &BankCard{
			logger: logger,
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
	Pan     string  `json:"pan" binding:"required"`
	Cvv     string  `json:"cvv" binding:"required"`
	ExpDate ExpDate `json:"expire" binding:"required"`
	Holder  string  `json:"holder" binding:"required"`
}

type CardAuthorizeRequest struct {
	Amount uint `json:"amount" binding:"required,min=1"`
	Card   Card `json:"card" binding:"required"`
}

type BankCard struct {
	logger repository.LoggerFunc
}

func (bc *BankCard) FromContext(c *gin.Context) (error, interface{}) {
	r, exists := c.Get("cardAuthorizeRequest")
	if !exists {
		return fmt.Errorf("cardAuthorizeRequest not exists in context"), nil
	}
	cardAuthorizeRequest := r.(CardAuthorizeRequest)

	s, exists := c.Get("cardStore")
	if !exists {
		return fmt.Errorf("cardStore not exists in context"), nil
	}
	cardStore := s.(repository.CardRepository)

	pan := repository.PAN(cardAuthorizeRequest.Card.Pan)
	cvv := repository.CVV(cardAuthorizeRequest.Card.Cvv)

	err, _, cards := cardStore.Query(c, repository.NewCardSpecificationByPAN(pan))
	if err != nil {
		return fmt.Errorf("card store quering failed: %v", err), nil
	}

	if len(cards) > 0 {
		return nil, cards[0]
	}

	card := &repository.Card{
		PAN:     &pan,
		CVV:     &cvv,
		ExpDate: &cardAuthorizeRequest.Card.ExpDate.Time,
		Holder:  &cardAuthorizeRequest.Card.Holder,
	}

	if err := cardStore.Add(c, card); err != nil {
		return fmt.Errorf("card store adding failed: %v", err), nil
	}

	return nil, card
}
