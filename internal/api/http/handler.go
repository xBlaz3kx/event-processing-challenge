package http

import (
	"net/http"

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
	// h.playerService.GetPlayerDetails()
}

func (h *Handler) GetEventDescription(c *gin.Context) {
	// h.playerService.GetPlayerDetails()
}

func (h *Handler) GetEurCurrency(c *gin.Context) {
	// h.playerService.GetPlayerDetails()
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
