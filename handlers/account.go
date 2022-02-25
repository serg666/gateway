package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type CreateAccountRequest struct {
	IsEnabled                 *bool                       `json:"is_enabled"`
	IsTest                    *bool                       `json:"is_test"`
	RebillEnabled             *bool                       `json:"rebill_enabled"`
	RefundEnabled             *bool                       `json:"refund_enabled"`
	ReversalEnabled           *bool                       `json:"reversal_enabled"`
	PartialConfirmEnabled     *bool                       `json:"partial_confirm_enabled"`
	PartialReversalEnabled    *bool                       `json:"partial_reversal_enabled"`
	PartialRefundEnabled      *bool                       `json:"partial_refund_enabled"`
	CurrencyConversionEnabled *bool                       `json:"currency_conversion_enabled"`
	CurrencyCode              *int                        `json:"currency_code" binding:"required"`
	ChannelKey                *string                     `json:"channel_key" binding:"required"`
	Settings                  *repository.AccountSettings `json:"settings"`
}

type UpdateAccountRequest struct {
	IsEnabled                 *bool                       `json:"is_enabled" binding:"required_without_all=IsTest RebillEnabled RefundEnabled ReversalEnabled PartialConfirmEnabled PartialReversalEnabled PartialRefundEnabled CurrencyConversionEnabled CurrencyCode ChannelKey Settings"`
	IsTest                    *bool                       `json:"is_test" binding:"required_without_all=IsEnabled RebillEnabled RefundEnabled ReversalEnabled PartialConfirmEnabled PartialReversalEnabled PartialRefundEnabled CurrencyConversionEnabled CurrencyCode ChannelKey Settings"`
	RebillEnabled             *bool                       `json:"rebill_enabled" binding:"required_without_all=IsEnabled IsTest RefundEnabled ReversalEnabled PartialConfirmEnabled PartialReversalEnabled PartialRefundEnabled CurrencyConversionEnabled CurrencyCode ChannelKey Settings"`
	RefundEnabled             *bool                       `json:"refund_enabled" binding:"required_without_all=IsEnabled IsTest RebillEnabled ReversalEnabled PartialConfirmEnabled PartialReversalEnabled PartialRefundEnabled CurrencyConversionEnabled CurrencyCode ChannelKey Settings"`
	ReversalEnabled           *bool                       `json:"reversal_enabled" binding:"required_without_all=IsEnabled IsTest RebillEnabled RefundEnabled PartialConfirmEnabled PartialReversalEnabled PartialRefundEnabled CurrencyConversionEnabled CurrencyCode ChannelKey Settings"`
	PartialConfirmEnabled     *bool                       `json:"partial_confirm_enabled" binding:"required_without_all=IsEnabled IsTest RebillEnabled RefundEnabled ReversalEnabled PartialReversalEnabled PartialRefundEnabled CurrencyConversionEnabled CurrencyCode ChannelKey Settings"`
	PartialReversalEnabled    *bool                       `json:"partial_reversal_enabled" binding:"required_without_all=IsEnabled IsTest RebillEnabled RefundEnabled ReversalEnabled PartialConfirmEnabled PartialRefundEnabled CurrencyConversionEnabled CurrencyCode ChannelKey Settings"`
	PartialRefundEnabled      *bool                       `json:"partial_refund_enabled" binding:"required_without_all=IsEnabled IsTest RebillEnabled RefundEnabled ReversalEnabled PartialConfirmEnabled PartialReversalEnabled CurrencyConversionEnabled CurrencyCode ChannelKey Settings"`
	CurrencyConversionEnabled *bool                       `json:"currency_conversion_enabled" binding:"required_without_all=IsEnabled IsTest RebillEnabled RefundEnabled ReversalEnabled PartialConfirmEnabled PartialReversalEnabled PartialRefundEnabled CurrencyCode ChannelKey Settings"`
	CurrencyCode              *int                        `json:"currency_code" binding:"required_without_all=IsEnabled IsTest RebillEnabled RefundEnabled ReversalEnabled PartialConfirmEnabled PartialReversalEnabled PartialRefundEnabled CurrencyConversionEnabled ChannelKey Settings"`
	ChannelKey                *string                     `json:"channel_key" binding:"omitempty,required_without_all=IsEnabled IsTest RebillEnabled RefundEnabled ReversalEnabled PartialConfirmEnabled PartialReversalEnabled PartialRefundEnabled CurrencyConversionEnabled CurrencyCode Settings,notempty"`
	Settings                  *repository.AccountSettings `json:"settings" binding:"required_without_all=IsEnabled IsTest RebillEnabled RefundEnabled ReversalEnabled PartialConfirmEnabled PartialReversalEnabled PartialRefundEnabled CurrencyConversionEnabled CurrencyCode ChannelKey"`
}

type accountHandler struct {
	loggerFunc    repository.LoggerFunc
	store         repository.AccountRepository
	currencyStore repository.CurrencyRepository
	channelStore  repository.ChannelRepository
}

func (ah *accountHandler) CreateAccountHandler(c *gin.Context) {
	t := true
	f := false
	// @note: set default values here
	req := CreateAccountRequest{
		IsEnabled: &t,
		IsTest: &f,
		RebillEnabled: &f,
		RefundEnabled: &t,
		ReversalEnabled: &t,
		PartialConfirmEnabled: &f,
		PartialReversalEnabled: &f,
		PartialRefundEnabled: &f,
		CurrencyConversionEnabled: &f,
		Settings: &repository.AccountSettings{},
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, _, currencies := ah.currencyStore.Query(c, repository.NewCurrencySpecificationByNumericCode(*req.CurrencyCode))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(currencies) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Currency with CurrencyCode=%v not found", *req.CurrencyCode),
		})
		return
	}

	err, _, channels := ah.channelStore.Query(c, repository.NewChannelSpecificationByKey(*req.ChannelKey))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(channels) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Channel with key %v not found", *req.ChannelKey),
		})
		return
	}

	account := &repository.Account{
		IsEnabled:                 req.IsEnabled,
		IsTest:                    req.IsTest,
		RebillEnabled:             req.RebillEnabled,
		RefundEnabled:             req.RefundEnabled,
		ReversalEnabled:           req.ReversalEnabled,
		PartialConfirmEnabled:     req.PartialConfirmEnabled,
		PartialReversalEnabled:    req.PartialReversalEnabled,
		PartialRefundEnabled:      req.PartialRefundEnabled,
		CurrencyConversionEnabled: req.CurrencyConversionEnabled,
		Currency:                  currencies[0],
		Channel:                   channels[0],
		Settings:                  req.Settings,
	}
	if err := ah.store.Add(c, account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (ah *accountHandler) refreshAccountCurrency(c *gin.Context, account *repository.Account) error {
	err, _, currencies := ah.currencyStore.Query(c, repository.NewCurrencySpecificationByID(
		*account.Currency.Id,
	))

	if err != nil {
		return fmt.Errorf("Can not update account currency: %v", err)
	}

	for _, currency := range currencies {
		account.Currency = currency
	}

	return nil
}

func (ah *accountHandler) refreshAccountChannel(c *gin.Context, account *repository.Account) error {
	err, _, channels := ah.channelStore.Query(c, repository.NewChannelSpecificationByID(
		*account.Channel.Id,
	))

	if err != nil {
		return fmt.Errorf("Can not update account channel: %v", err)
	}

	for _, channel := range channels {
		account.Channel = channel
	}

	return nil
}

func (ah *accountHandler) GetAccountsHandler(c *gin.Context) {
	var req LimitAndOffsetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, overall, accounts := ah.store.Query(c, repository.NewAccountSpecificationWithLimitAndOffset(
		req.Limit,
		req.Offset,
	))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	for _, account := range accounts {
		if err := ah.refreshAccountCurrency(c, account); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		if err := ah.refreshAccountChannel(c, account); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"overall": overall,
		"accounts": accounts,
	})
}

func (ah *accountHandler) GetAccountHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err, _, accounts := ah.store.Query(c, repository.NewAccountSpecificationByID(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(accounts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Account with id=%v not found", id),
		})
		return
	}

	for _, account := range accounts {
		if err := ah.refreshAccountCurrency(c, account); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		if err := ah.refreshAccountChannel(c, account); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, accounts[0])
}

func (ah *accountHandler) DeleteAccountHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	account := &repository.Account{
		Id:       &id,
		Currency: &repository.Currency{},
		Channel:  &repository.Channel{},
	}

	err, notfound := ah.store.Delete(c, account)

	if notfound {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err !=  nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if account.Currency != nil {
		if err := ah.refreshAccountCurrency(c, account); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	if account.Channel != nil {
		if err := ah.refreshAccountChannel(c, account); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, account)
}

func (ah *accountHandler) PatchAccountHandler(c *gin.Context) {
	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	currency := &repository.Currency{}
	channel := &repository.Channel{}

	if req.CurrencyCode != nil {
		err, _, currencies := ah.currencyStore.Query(c, repository.NewCurrencySpecificationByNumericCode(*req.CurrencyCode))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(currencies) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Currency with CurrencyCode=%v not found", *req.CurrencyCode),
			})
			return
		}
		currency = currencies[0]
	}

	if req.ChannelKey != nil {
		err, _, channels := ah.channelStore.Query(c, repository.NewChannelSpecificationByKey(*req.ChannelKey))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		if len(channels) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Channel with ChannelKey=%v not found", *req.ChannelKey),
			})
			return
		}
		channel = channels[0]
	}

	account := &repository.Account{
		Id:                        &id,
		IsEnabled:                 req.IsEnabled,
		IsTest:                    req.IsTest,
		RebillEnabled:             req.RebillEnabled,
		RefundEnabled:             req.RefundEnabled,
		ReversalEnabled:           req.ReversalEnabled,
		PartialConfirmEnabled:     req.PartialConfirmEnabled,
		PartialReversalEnabled:    req.PartialReversalEnabled,
		PartialRefundEnabled:      req.PartialRefundEnabled,
		CurrencyConversionEnabled: req.CurrencyConversionEnabled,
		Settings:                  req.Settings,
		Currency:                  currency,
		Channel:                   channel,
	}

	err, notfound := ah.store.Update(c, account)

	if notfound {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err !=  nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	// @note: refresh account currency
	if account.Currency.Id != nil && req.CurrencyCode == nil {
		if err := ah.refreshAccountCurrency(c, account); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	// @note: refresh account channel
	if account.Channel.Id != nil && req.ChannelKey == nil {
		if err := ah.refreshAccountChannel(c, account); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, account)
}

func NewAccountHandler(
	store repository.AccountRepository,
	currencyStore repository.CurrencyRepository,
	channelStore repository.ChannelRepository,
	loggerFunc repository.LoggerFunc,
) *accountHandler {
	return &accountHandler{
		loggerFunc:    loggerFunc,
		store:         store,
		currencyStore: currencyStore,
		channelStore:  channelStore,
	}
}
