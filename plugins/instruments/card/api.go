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
		requesterFunc plugins.InstrumentRequesterFunc,
	) instruments.PaymentInstrument {
		return &BankCard{
			requesterFunc:   requesterFunc,
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

type BankCard struct {
	requesterFunc   plugins.InstrumentRequesterFunc
	instrumentStore interface{}
	logger          repository.LoggerFunc
}

func (bc *BankCard) FromRequest(c *gin.Context, request interface{}) (error, interface{}) {
	err, bankCardRequest := bc.requesterFunc(request)
	if err != nil {
		return fmt.Errorf("can not get bank card request: %v"), nil
	}

	bankCardReq, ok := bankCardRequest.(Card)
	if !ok {
		return fmt.Errorf("bank card request has wrong type"), nil
	}

	cardStore, ok := bc.instrumentStore.(repository.CardRepository)
	if !ok {
		return fmt.Errorf("instrumentStore has wrong type"), nil
	}

	pan := repository.PAN(bankCardReq.Pan)

	err, _, cards := cardStore.Query(c, repository.NewCardSpecificationByPAN(pan))
	if err != nil {
		return fmt.Errorf("card store quering failed: %v", err), nil
	}

	if len(cards) > 0 {
		return nil, cards[0]
	}

	card := &repository.Card{
		PAN:     &pan,
		ExpDate: &repository.ExpDate{bankCardReq.ExpDate.Time},
		Holder:  &bankCardReq.Holder,
	}

	if err := cardStore.Add(c, card); err != nil {
		return fmt.Errorf("card store adding failed: %v", err), nil
	}

	return nil, card
}
