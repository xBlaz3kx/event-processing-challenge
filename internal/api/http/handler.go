package http

import (
	"github.com/gin-gonic/gin"
	"github.com/xBlaz3kx/event-processing-challenge/internal/currency"
	"github.com/xBlaz3kx/event-processing-challenge/internal/event"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player"
)

type Handler struct {
	processor       *event.DescriptionProcessor
	currencyService currency.Service
	playerService   player.Service
}

func NewHandler(currencyService currency.Service, playerService player.Service) *Handler {
	return &Handler{
		processor:       event.NewDescriptionProcessor(),
		currencyService: currencyService,
		playerService:   playerService,
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
	// h.playerService.GetPlayerDetails()
}
