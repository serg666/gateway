package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/serg666/repository"
)

type LimitAndOffsetRequest struct {
	Limit  int `form:"limit,default=100"`
	Offset int `form:"offset,default=0"`
}

type CreateProfileRequest struct {
	Key          *string `json:"key" binding:"required,notempty"`
	CurrencyCode *int    `json:"currencycode" binding:"required"`
	Description  *string `json:"description" binding:"omitempty,notempty"`
}

type UpdateProfileRequest struct {
	Key          *string `json:"key" binding:"omitempty,required_without_all=CurrencyCode Description,notempty"`
	CurrencyCode *int    `json:"currencycode" binding:"required_without_all=Key Description"`
	Description  *string `json:"description" binding:"omitempty,required_without_all=Key CurrencyCode,notempty"`
}

type profileHandler struct {
	store         repository.ProfileRepository
	currencyStore repository.CurrencyRepository
}

func (ph *profileHandler) DeleteProfileHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	profile := &repository.Profile{Id: id}
	if err := ph.store.Delete(profile); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (ph *profileHandler) GetProfileHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err !=  nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	_, profiles := ph.store.Query(repository.NewProfileSpecificationByID(id))

	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Profile with id=%v not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, profiles[0])
}

func (ph *profileHandler) PatchProfileHandler(c *gin.Context) {
	var req UpdateProfileRequest
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

	var currency *repository.Currency

	if req.CurrencyCode != nil {
		_, currencies := ph.currencyStore.Query(repository.NewCurrencySpecificationByNumericCode(*req.CurrencyCode))
		if len(currencies) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Currency with CurrencyCode=%v not found", *req.CurrencyCode),
			})
			return
		}
		currency = &currencies[0]
	}

	profile := &repository.Profile{
		Id:          id,
		Key:         req.Key,
		Description: req.Description,
		Currency:    currency,
	}
	if err := ph.store.Update(profile); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (ph *profileHandler) CreateProfileHandler(c *gin.Context) {
	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	_, currencies := ph.currencyStore.Query(repository.NewCurrencySpecificationByNumericCode(*req.CurrencyCode))
	if len(currencies) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Currency with CurrencyCode=%v not found", *req.CurrencyCode),
		})
		return
	}

	profile := &repository.Profile{
		Key:         req.Key,
		Description: req.Description,
		Currency:    &currencies[0],
	}
	if err := ph.store.Add(profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (ph *profileHandler) GetProfilesHandler(c *gin.Context) {
	var req LimitAndOffsetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	overall, profiles := ph.store.Query(repository.NewProfileSpecificationWithLimitAndOffset(
		req.Limit,
		req.Offset,
	))
	c.JSON(http.StatusOK, gin.H{
		"overall": overall,
		"profiles": profiles,
	})
}

func NewProfileHandler(store repository.ProfileRepository, currencyStore repository.CurrencyRepository) *profileHandler {
	return &profileHandler{
		store:         store,
		currencyStore: currencyStore,
	}
}
