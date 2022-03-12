package alfabank

import (
	"fmt"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
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
		cfg     *config.Config,
		logger  repository.LoggerFunc,
	) channels.BankChannel {
		return &AlfaBankChannel{
			cfg:     cfg,
			logger:  logger,
		}
	})
)

type AlfaBankSettings struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AlfaBankChannel struct {
	cfg      *config.Config
	logger   repository.LoggerFunc
	settings AlfaBankSettings
}

func (abc *AlfaBankChannel) SutableForInstrument(instrument *repository.Instrument) bool {
	return *instrument.Id == bankcard.Id
}

func (abc *AlfaBankChannel) DecodeSettings(settings *repository.AccountSettings) error {
	jsonbody, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("can not marshal alfabank account settings: %v", err)
	}

	d := json.NewDecoder(bytes.NewReader(jsonbody))
	d.DisallowUnknownFields()

	if err := d.Decode(&abc.settings); err != nil {
		return fmt.Errorf("can not decode alfabank account settings: %v", err)
	}

	return nil
}

func (abc *AlfaBankChannel) Authorize(c *gin.Context, transaction *repository.Transaction, instrumentInstance interface{}) {
	card := instrumentInstance.(*repository.Card)

	abc.logger(c).Printf("authorize card: %v", card)
	abc.logger(c).Printf("url: %v", abc.cfg.Alfabank.Ecom.Url)
	abc.logger(c).Printf("settings: %v", abc.settings)
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
