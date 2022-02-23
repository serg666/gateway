package kvellbank

import (
	"log"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/channels"
)

var (
	Id  = 1
	Key = "kvellbank"
	Registered = plugins.RegisterBankChannel(Id, Key, func(a int) channels.BankChannel {
		return &KvellBankChannel{
			a: a,
		}
	})
)

type KvellBankChannel struct {
	a int
}

func (kbc *KvellBankChannel) Authorize() {
	log.Print("authorize int")
}

func (kbc *KvellBankChannel) PreAuthorize() {
	log.Print("preauthorize int")
}

func (kbc *KvellBankChannel) Confirm() {
	log.Print("confirm int")
}

func (kbc *KvellBankChannel) Reverse() {
	log.Print("reverse int")
}

func (kbc *KvellBankChannel) Refund() {
	log.Print("refund int")
}

func (kbc *KvellBankChannel) Complete3DS() {
	log.Print("complete3ds int")
}
