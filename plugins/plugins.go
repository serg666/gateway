package plugins

import (
	"fmt"
	"github.com/serg666/gateway/config"
	"github.com/serg666/gateway/plugins/routers"
	"github.com/serg666/gateway/plugins/channels"
	"github.com/serg666/gateway/plugins/instruments"

	"github.com/serg666/repository"
)

var (
	Routers            = make(map[int]*Router)
	BankChannels       = make(map[int]*BankChannel)
	PaymentInstruments = make(map[int]*PaymentInstrument)
)

type RouterFunc func (*repository.Route, repository.AccountRepository, repository.LoggerFunc) (error, routers.Router)

type Router struct {
	Key    string
	Plugin RouterFunc
}

func (r Router) String() string {
	return fmt.Sprintf("router <%s>", r.Key)
}

func RouterApi(route *repository.Route, accountStore repository.AccountRepository, logger repository.LoggerFunc) (error, routers.Router) {
	if route.Router == nil {
		return fmt.Errorf("Route ID=<%d> has no router", *route.Id), nil
	}

	if val, ok := Routers[*route.Router.Id]; ok {
		err, api := val.Plugin(route, accountStore, logger)
		if err != nil {
			return fmt.Errorf("failed to initiate router api: %v", err), nil
		}

		return nil, api
	}

	return fmt.Errorf("Router with ID=%v not found", *route.Router.Id), nil
}

func RegisterRouter(id int, key string, routerFunc RouterFunc) error {
	if val, ok := Routers[id]; ok {
		return fmt.Errorf("ID <%d> has already used for: %s", id, val)
	}

	Routers[id] = &Router{
		Key:    key,
		Plugin: routerFunc,
	}

	return nil
}

func RegisterRouters(routerStore repository.RouterRepository) error {
	for Id, Router := range Routers {
		err, _, routers := routerStore.Query(nil, repository.NewRouterSpecificationByID(Id))
		if err != nil {
			return fmt.Errorf("Failed to query routers: %v", err)
		}

		if len(routers) > 0 {
			router := routers[0]
			if Router.Key != *router.Key {
				return fmt.Errorf("Router %s already uses id=%d", *router.Key, Id)
			}
		} else {
			router := &repository.Router{
				Id:  &Id,
				Key: &Router.Key,
			}
			if err := routerStore.Add(nil, router); err != nil {
				return fmt.Errorf("Failed to add router: %v", err)
			}
		}
	}
	return nil
}

func CheckRouters(routerStore repository.RouterRepository) error {
	err, _, routers := routerStore.Query(nil, repository.NewRouterWithoutSpecification())
	if err != nil {
		return fmt.Errorf("Failed to query routers: %v", err)
	}

	regged := len(routers)
	loaded := len(Routers)
	if loaded != regged {
		return fmt.Errorf("Loaded %d, registered %d routers", loaded, regged)
	}

	for _, router := range routers {
		if val, ok := Routers[*router.Id]; ok {
			if val.Key != *router.Key {
				return fmt.Errorf("%s (id=%d) registered with key=%s", val, *router.Id, *router.Key)
			}
		}
	}

	return nil
}

type PaymentInstrumentFunc func (repository.LoggerFunc) instruments.PaymentInstrument

type PaymentInstrument struct {
	Key    string
	Plugin PaymentInstrumentFunc
}

func (pi PaymentInstrument) String() string {
	return fmt.Sprintf("payment instrument <%s>", pi.Key)
}

func InstrumentApi(instrument *repository.Instrument, logger repository.LoggerFunc) (error, instruments.PaymentInstrument) {
	if val, ok := PaymentInstruments[*instrument.Id]; ok {
		return nil, val.Plugin(logger)
	}

	return fmt.Errorf("Instrument with ID=%v not found", *instrument.Id), nil
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

type BankChannelFunc func (*config.Config, *repository.Route, repository.LoggerFunc) (error, channels.BankChannel)

type BankChannel struct {
	Key    string
	Type   int
	Plugin BankChannelFunc
}

func (bc BankChannel) String() string {
	return fmt.Sprintf("bank channel <%s>", bc.Key)
}

func BankApi(cfg *config.Config, route *repository.Route, logger repository.LoggerFunc) (error, channels.BankChannel) {
	cid := *route.Account.Channel.Id

	if val, ok := BankChannels[cid]; ok {
		err, api := val.Plugin(cfg, route, logger)
		if err != nil {
			return fmt.Errorf("failed to initiate bank api: %v", err), nil
		}

		return nil, api
	}

	return fmt.Errorf("Bank channel with ID=%v not found", cid), nil
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
