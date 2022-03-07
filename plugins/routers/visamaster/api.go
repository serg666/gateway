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
		settings *repository.RouterSettings,
		logger repository.LoggerFunc,
	) routers.Router {
		return &VisaMasterRouter{
			logger:   logger,
			settings: settings,
		}
	})
)

type VisaMasterSettings struct {
	VisaAcc   int `json:"visa_acc"`
	MasterAcc int `json:"master_acc"`
}

type VisaMasterRouter struct {
	logger   repository.LoggerFunc
	settings *repository.RouterSettings
}

func (vmr *VisaMasterRouter) SutableForInstrument(instrument *repository.Instrument) bool {
	return *instrument.Id == bankcard.Id
}

func (vmr *VisaMasterRouter) decodeSettings() (error, *VisaMasterSettings) {
	jsonbody, err := json.Marshal(vmr.settings)
	if err != nil {
		return fmt.Errorf("can not marshal route settings: %v", err), nil
	}

	d := json.NewDecoder(bytes.NewReader(jsonbody))
	d.DisallowUnknownFields()

	var vms VisaMasterSettings
	if err := d.Decode(&vms); err != nil {
		return fmt.Errorf("can not decode route settings: %v", err), nil
	}

	return nil, &vms
}

func (vmr *VisaMasterRouter) Route(c *gin.Context, route *repository.Route) error {
	err, vms := vmr.decodeSettings()
	if err != nil {
		return fmt.Errorf("failed to validate router settings: %v", err)
	}

	err, instrumentApi := plugins.InstrumentApi(route.Instrument, vmr.logger)
	if err != nil {
		return fmt.Errorf("failed to get instrumentApi: %v", err)
	}

	err, instrument := instrumentApi.FromContext(c)
	if err != nil {
		return fmt.Errorf("failed to get instrument from context: %v", err)
	}

	card := instrument.(*repository.Card)
	vmr.logger(c).Printf("card: %v", card)

	a, exists := c.Get("accountStore")
	if !exists {
		return fmt.Errorf("accountStore not exists in context")
	}
	accountStore := a.(repository.AccountRepository)

	err, _, maccs := accountStore.Query(c, repository.NewAccountSpecificationByID(vms.MasterAcc))
	if err != nil {
		return fmt.Errorf("failed to query acc store: %v", err)
	}

	err, _, vaccs := accountStore.Query(c, repository.NewAccountSpecificationByID(vms.VisaAcc))
	if err != nil {
		return fmt.Errorf("failed to query acc store: %v", err)
	}

	if len(maccs) == 0 {
		return fmt.Errorf("mastercard account %d not found", vms.MasterAcc)
	}

	if len(vaccs) == 0 {
		return fmt.Errorf("visa account %d not found", vms.MasterAcc)
	}

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
