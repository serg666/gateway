package plugins

import (
	"fmt"
	"github.com/serg666/gateway/plugins/channels"
	"github.com/serg666/gateway/plugins/instruments"

	"github.com/serg666/repository"
)

var (
	BankChannels       = make(map[int]*BankChannel)
	PaymentInstruments = make(map[int]*PaymentInstrument)
)

type PaymentInstrumentFunc func (repository.LoggerFunc) instruments.PaymentInstrument

type PaymentInstrument struct {
	Key    string
	Plugin PaymentInstrumentFunc
}

func (pi PaymentInstrument) String() string {
	return fmt.Sprintf("payment instrument <%s>", pi.Key)
}

func InstrumentApi(instrument *repository.Instrument, logger repository.LoggerFunc) (error, instruments.PaymentInstrument) {
	return nil, PaymentInstruments[*instrument.Id].Plugin(logger)
}

func RegisterPaymentInstrument(id int, key string, instrumentFunc PaymentInstrumentFunc) error {
	if val, ok := PaymentInstruments[id]; ok {
		return fmt.Errorf("ID <%d> has already used for: %s", id, val)
	}

	PaymentInstruments[id] = &PaymentInstrument{
		Key:    key,
		Plugin: instrumentFunc,
	}

	return nil
}

func RegisterPaymentInstruments(instrumentStore repository.InstrumentRepository) error {
	for Id, PaymentInstrument := range PaymentInstruments {
		err, _, paymentInstruments := instrumentStore.Query(nil, repository.NewInstrumentSpecificationByID(Id))
		if err != nil {
			return fmt.Errorf("Failed to query payment instruments: %v", err)
		}

		if len(paymentInstruments) > 0 {
			paymentInstrument := paymentInstruments[0]
			if PaymentInstrument.Key != *paymentInstrument.Key {
				return fmt.Errorf("Payment instrument %s already uses id=%d", *paymentInstrument.Key, Id)
			}
		} else {
			instrument := &repository.Instrument{
				Id:  &Id,
				Key: &PaymentInstrument.Key,
			}
			if err := instrumentStore.Add(nil, instrument); err != nil {
				return fmt.Errorf("Failed to add payment instrument: %v", err)
			}
		}
	}
	return nil
}

func CheckPaymentInstruments(instrumentStore repository.InstrumentRepository) error {
	err, _, paymentInstruments := instrumentStore.Query(nil, repository.NewInstrumentWithoutSpecification())
	if err != nil {
		return fmt.Errorf("Failed to query payment instruments: %v", err)
	}

	regged := len(paymentInstruments)
	loaded := len(PaymentInstruments)
	if loaded != regged {
		return fmt.Errorf("Loaded %d, registered %d payment instruments", loaded, regged)
	}

	for _, paymentInstrument := range paymentInstruments {
		if val, ok := PaymentInstruments[*paymentInstrument.Id]; ok {
			if val.Key != *paymentInstrument.Key {
				return fmt.Errorf("%s (id=%d) registered with key=%s", val, *paymentInstrument.Id, *paymentInstrument.Key)
			}
		}
	}

	return nil
}

type BankChannelFunc func (*repository.Account, repository.LoggerFunc) channels.BankChannel

type BankChannel struct {
	Key    string
	Type   int
	Plugin BankChannelFunc
}

func (bc BankChannel) String() string {
	return fmt.Sprintf("bank channel <%s>", bc.Key)
}

func BankApi(channel *repository.Channel, account *repository.Account, logger repository.LoggerFunc) (error, channels.BankChannel) {
	aid := *account.Channel.Id
	cid := *channel.Id
	if aid != cid {
		return fmt.Errorf("account channel id %d != channel id %d", aid, cid), nil
	}

	return nil, BankChannels[cid].Plugin(account, logger)
}

func RegisterBankChannel(id int, key string, channelFunc BankChannelFunc) error {
	if val, ok := BankChannels[id]; ok {
		return fmt.Errorf("ID <%d> has already used for: %s", id, val)
	}

	BankChannels[id] = &BankChannel{
		Key:    key,
		Type:   channels.BankChannelType,
		Plugin: channelFunc,
	}

	return nil
}

func RegisterBankChannels(channelStore repository.ChannelRepository) error {
	for Id, BankChannel := range BankChannels {
		err, _, bankChannels := channelStore.Query(nil, repository.NewChannelSpecificationByID(Id))
		if err != nil {
			return fmt.Errorf("Failed to query bank channels: %v", err)
		}

		if len(bankChannels) > 0 {
			bankChannel := bankChannels[0]
			if BankChannel.Key != *bankChannel.Key {
				return fmt.Errorf("Bank channel %s already uses id=%d", *bankChannel.Key, Id)
			}
		} else {
			channel := &repository.Channel{
				Id:     &Id,
				TypeId: &BankChannel.Type,
				Key:    &BankChannel.Key,
			}
			if err := channelStore.Add(nil, channel); err != nil {
				return fmt.Errorf("Failed to add bank channel: %v", err)
			}
		}
	}
	return nil
}

func CheckBankChannels(channelStore repository.ChannelRepository) error {
	err, _, bankChannels := channelStore.Query(nil, repository.NewChannelSpecificationByTypeID(channels.BankChannelType))
	if err != nil {
		return fmt.Errorf("Failed to query bank channels: %v", err)
	}

	regged := len(bankChannels)
	loaded := len(BankChannels)
	if loaded != regged {
		return fmt.Errorf("Loaded %d, registered %d bank channels", loaded, regged)
	}

	for _, bankChannel := range bankChannels {
		if val, ok := BankChannels[*bankChannel.Id]; ok {
			if val.Key != *bankChannel.Key {
				return fmt.Errorf("%s (id=%d) registered with key=%s", val, *bankChannel.Id, *bankChannel.Key)
			}
		}
	}

	return nil
}
