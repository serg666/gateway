package alfabank

import (
	"log"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/channels"
)

var (
	Id  = 2
	Key = "alfabank"
	Registered = plugins.RegisterBankChannel(Id, Key, func(a int) channels.BankChannel {
		return &AlfaBankChannel{
			a: a,
		}
	})
)

type AlfaBankChannel struct {
	a int
}

func (abc *AlfaBankChannel) Authorize() {
	log.Print("authorize")
}

func (abc *AlfaBankChannel) PreAuthorize() {
	log.Print("preauthorize")
}

func (abc *AlfaBankChannel) Confirm() {
	log.Print("confirm")
}

func (abc *AlfaBankChannel) Reverse() {
	log.Print("reverse")
}

func (abc *AlfaBankChannel) Refund() {
	log.Print("refund")
}

func (abc *AlfaBankChannel) Complete3DS() {
	log.Print("complete3ds")
}
