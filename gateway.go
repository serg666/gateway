package main

import (
	"log"
	"github.com/sirupsen/logrus"
	"github.com/wk8/go-ordered-map"
	"github.com/serg666/repository"
	"github.com/serg666/gateway/client"
	"github.com/serg666/gateway/config"

	"github.com/serg666/gateway/plugins"

	"github.com/serg666/gateway/plugins/routers/visamaster"

	"github.com/serg666/gateway/plugins/instruments/card"

	"github.com/serg666/gateway/plugins/channels/kvellbank"
	"github.com/serg666/gateway/plugins/channels/alfabank"
)

func main() {
	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatalf("can not parse flags due to: %v", err)
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatalf("can not get new config due to: %v", err)
	}

	pgPool, err := repository.MakePgPoolFromDSN(cfg.Databases.Default.Dsn)
	if err != nil {
		log.Fatalf("Can not make pg pool: %v", err)
	}

	loggerFunc := func (c interface{}) logrus.FieldLogger {
		return cfg.LogRusLogger(c)
	}

	//currencyStore := repository.NewOrderedMapCurrencyStore(orderedmap.New(), loggerFunc)
	currencyStore := repository.NewPGPoolCurrencyStore(pgPool, loggerFunc)
	//profileStore := repository.NewOrderedMapProfileStore(orderedmap.New(), currencyStore, loggerFunc)
	profileStore := repository.NewPGPoolProfileStore(pgPool, currencyStore, loggerFunc)
	sessionStore := repository.NewOrderedMapSessionStore(orderedmap.New(), loggerFunc)
	cardStore := repository.NewOrderedMapCardStore(orderedmap.New(), loggerFunc)
	channelStore := repository.NewPGPoolChannelStore(pgPool, loggerFunc)
	accountStore := repository.NewPGPoolAccountStore(pgPool, currencyStore, channelStore, loggerFunc)
	instrumentStore := repository.NewPGPoolInstrumentStore(pgPool, loggerFunc)
	routerStore := repository.NewPGPoolRouterStore(pgPool, loggerFunc)
	routeStore := repository.NewPGPoolRouteStore(
		pgPool,
		profileStore,
		instrumentStore,
		accountStore,
		routerStore,
		loggerFunc,
	)
	transactionStore := repository.NewPGPoolTransactionStore(
		pgPool,
		profileStore,
		instrumentStore,
		accountStore,
		currencyStore,
		loggerFunc,
	)

	if visamaster.Registered != nil {
		log.Fatalf("Can not register visamaster router: %v", visamaster.Registered)
	}

	if bankcard.Registered != nil {
		log.Fatalf("Can not register bank card instrument type: %v", bankcard.Registered)
	}

	if kvellbank.Registered != nil {
		log.Fatalf("Can not register kvellbank channel: %v", kvellbank.Registered)
	}

	if alfabank.Registered != nil {
		log.Fatalf("Can not register alfabank channel: %v", alfabank.Registered)
	}

	if err := plugins.RegisterBankChannels(channelStore); err != nil {
		log.Fatalf("Failed to register bank channels: %v", err)
	}

	if err := plugins.CheckBankChannels(channelStore); err != nil {
		log.Fatalf("Failed to check bank channels: %v", err)
	}

	if err := plugins.RegisterPaymentInstruments(instrumentStore); err != nil {
		log.Fatalf("Failed to register payment instruments: %v", err)
	}

	if err := plugins.CheckPaymentInstruments(instrumentStore); err != nil {
		log.Fatalf("Failed to check payment instruments: %v", err)
	}

	if err := plugins.RegisterRouters(routerStore); err != nil {
		log.Fatalf("Failed to register routers: %v", err)
	}

	if err := plugins.CheckRouters(routerStore); err != nil {
		log.Fatalf("Failed to check routers: %v", err)
	}

    handler := MakeHandler(
		routeStore,
		routerStore,
		instrumentStore,
		accountStore,
		channelStore,
		profileStore,
		currencyStore,
		cardStore,
		transactionStore,
		sessionStore,
		cfg,
		loggerFunc,
    )

	client.Client = cfg.HttpClient()

	// Run the server
	cfg.RunServer(handler, loggerFunc)
}
