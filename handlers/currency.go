package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type CreateCurrencyRequest struct {
	NumericCode *int    `json:"numericcode" binding:"required"`
	Name        *string `json:"name" binding:"required,notempty"`
	CharCode    *string `json:"charcode" binding:"required,notempty"`
	Exponent    *int    `json:"exponent"`
}

type UpdateCurrencyRequest struct {
	NumericCode *int    `json:"numericcode" binding:"required_without_all=Name CharCode Exponent"`
	Name        *string `json:"name" binding:"omitempty,required_without_all=NumericCode CharCode Exponent,notempty"`
	CharCode    *string `json:"charcode" binding:"omitempty,required_without_all=NumericCode Name Exponent,notempty"`
	Exponent    *int    `json:"exponent" binding:"required_without_all=NumericCode Name CharCode"`
}

type currencyHandler struct {
	store repository.CurrencyRepository
}

func (ch *currencyHandler) CreateCurrencyHandler(c *gin.Context) {
	var req CreateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	currency := &repository.Currency{
		NumericCode: req.NumericCode,
		Name:        req.Name,
		CharCode:    req.CharCode,
		Exponent:    req.Exponent,
	}
	if err := ch.store.Add(currency); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, currency)
}

func (ch *currencyHandler) GetCurrenciesHandler(c *gin.Context) {
	var req LimitAndOffsetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	overall, currencies := ch.store.Query(repository.NewCurrencySpecificationWithLimitAndOffset(
		req.Limit,
		req.Offset,
	))
	c.JSON(http.StatusOK, gin.H{
		"overall": overall,
		"currencies": currencies,
	})
}

func (ch *currencyHandler) GetCurrencyHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	_, currencies := ch.store.Query(repository.NewCurrencySpecificationByID(id))

	if len(currencies) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Currency with id=%v not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, currencies[0])
}

func (ch *currencyHandler) PatchCurrencyHandler(c *gin.Context) {
	var req UpdateCurrencyRequest
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

	currency := &repository.Currency{
		Id:          id,
		NumericCode: req.NumericCode,
		Name:        req.Name,
		CharCode:    req.CharCode,
		Exponent:    req.Exponent,
	}
	if err := ch.store.Update(currency); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, currency)
}

func (ch *currencyHandler) DeleteCurrencyHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	currency := &repository.Currency{Id: id}
	if err := ch.store.Delete(currency); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, currency)
}

func NewCurrencyHandler(store repository.CurrencyRepository) *currencyHandler {
	return &currencyHandler{
		store: store,
	}
}
