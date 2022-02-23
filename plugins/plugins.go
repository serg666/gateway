package plugins

import (
	"fmt"
	"github.com/serg666/gateway/plugins/channels"

	"github.com/serg666/repository"
)

type BankChannelFunc func (int) channels.BankChannel

type BankChannel struct {
	Key    string
	Type   int
	Plugin BankChannelFunc
}

func (bc BankChannel) String() string {
	return fmt.Sprintf("bank channel <%s>", bc.Key)
}

var BankChannels = make(map[int]*BankChannel)

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
				Id:     Id,
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
		if val, ok := BankChannels[bankChannel.Id]; ok {
			if val.Key != *bankChannel.Key {
				return fmt.Errorf("%s (id=%d) registered with key=%s", val, bankChannel.Id, *bankChannel.Key)
			}
		}
	}

	return nil
}
