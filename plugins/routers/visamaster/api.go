package visamaster

import (
	"fmt"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/serg666/gateway/plugins"
	"github.com/serg666/gateway/plugins/routers"
	"github.com/serg666/gateway/plugins/instruments/card"
	"github.com/serg666/repository"
)

var (
	Id  = 1
	Key = "visamaster"
	Registered = plugins.RegisterRouter(Id, Key, func(
		route           *repository.Route,
		accountStore    repository.AccountRepository,
		instrumentStore interface{},
		logger          repository.LoggerFunc,
	) (error, routers.Router) {
		if *route.Instrument.Id != bankcard.Id {
			return fmt.Errorf("visamaster router not sutable for instrument <%d>", *route.Instrument.Id), nil
		}

		jsonbody, err := json.Marshal(route.Settings)
		if err != nil {
			return fmt.Errorf("can not marshal route settings: %v", err), nil
		}

		d := json.NewDecoder(bytes.NewReader(jsonbody))
		d.DisallowUnknownFields()

		var vms VisaMasterSettings

		if err := d.Decode(&vms); err != nil {
			return fmt.Errorf("can not decode route settings: %v", err), nil
		}

		return nil, &VisaMasterRouter{
			logger:          logger,
			accountStore:    accountStore,
			instrumentStore: instrumentStore,
			settings:        &vms,
		}
	})
)

type VisaMasterSettings struct {
	VisaAcc   int `json:"visa_acc"`
	MasterAcc int `json:"master_acc"`
}

type VisaMasterRouter struct {
	logger          repository.LoggerFunc
	accountStore    repository.AccountRepository
	instrumentStore interface{}
	settings        *VisaMasterSettings
}

func (vmr *VisaMasterRouter) Route(c *gin.Context, route *repository.Route, request interface{}) error {
	req, ok := request.(bankcard.CardAuthorizeRequest)
	if !ok {
		return fmt.Errorf("request has wrong type")
	}

	err, instrumentApi := plugins.InstrumentApi(route.Instrument, vmr.instrumentStore, vmr.logger)
	if err != nil {
		return fmt.Errorf("failed to get instrument api: %v", err)
	}

	err, instrumentInstance := instrumentApi.FromRequest(c, req)
	if err != nil {
		return fmt.Errorf("failed to get instrumentInstance: %v", err)
	}

	card, ok := instrumentInstance.(*repository.Card)
	if !ok {
		return fmt.Errorf("instrumentInstance has wrong type")
	}

	err, _, maccs := vmr.accountStore.Query(c, repository.NewAccountSpecificationByID(vmr.settings.MasterAcc))
	if err != nil {
		return fmt.Errorf("failed to query acc store: %v", err)
	}

	err, _, vaccs := vmr.accountStore.Query(c, repository.NewAccountSpecificationByID(vmr.settings.VisaAcc))
	if err != nil {
		return fmt.Errorf("failed to query acc store: %v", err)
	}

	if len(maccs) == 0 {
		return fmt.Errorf("mastercard account %d not found", vmr.settings.MasterAcc)
	}

	if len(vaccs) == 0 {
		return fmt.Errorf("visa account %d not found", vmr.settings.MasterAcc)
	}

	vmr.logger(c).Printf("visamaster routing card: %v", card)

	macc := maccs[0]
	vacc := vaccs[0]

	switch ctype := card.Type(); ctype {
	case "visa":
		route.Account = vacc
	case "mastercard":
		route.Account = macc
	}

	return nil
}
