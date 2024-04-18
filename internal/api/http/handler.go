package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xBlaz3kx/event-processing-challenge/internal/casino"
	"github.com/xBlaz3kx/event-processing-challenge/internal/currency"
	"github.com/xBlaz3kx/event-processing-challenge/internal/event"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player"
)

type Handler struct {
	processor       *event.DescriptionProcessor
	currencyService currency.Service
	playerService   player.Service
	casinoService   casino.Service
}

func NewHandler(currencyService currency.Service, playerService player.Service, casinoService casino.Service) *Handler {
	return &Handler{
		processor:       event.NewDescriptionProcessor(),
		currencyService: currencyService,
		playerService:   playerService,
		casinoService:   casinoService,
	}
}

func (h *Handler) GetPlayerDetails(c *gin.Context) {
	playerId := c.Param("id")

	playerIdInt, err := strconv.Atoi(playerId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	playerDetails, err := h.playerService.GetPlayerDetails(c.Request.Context(), playerIdInt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, playerDetails)
}

func (h *Handler) GetEventDescription(c *gin.Context) {
	var ev casino.Event
	err := c.BindJSON(&ev)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	description, err := h.processor.Process(ev)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, description)
}

func (h *Handler) GetEurCurrency(c *gin.Context) {
	var conv Conversion
	err := c.BindJSON(&conv)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	conversion, err := h.currencyService.Convert(c.Request.Context(), conv.Currency, conv.DesiredCurrency, conv.Amount)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ConversionResponse{conversion})
}

func (h *Handler) Materialize(c *gin.Context) {
	statistics, err := h.casinoService.GetStatistics(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, statistics)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Conversion struct {
	Amount          int    `json:"amount"`
	Currency        string `json:"currency"`
	DesiredCurrency string `json:"desired_currency"`
}

type ConversionResponse struct {
	Amount int `json:"amount"`
}
