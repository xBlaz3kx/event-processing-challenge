package kafka

import (
	"github.com/xBlaz3kx/event-processing-challenge/internal/casino"
	"github.com/xBlaz3kx/event-processing-challenge/internal/currency"
	"github.com/xBlaz3kx/event-processing-challenge/internal/event"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/kafka"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Api struct {
	logger          *zap.Logger
	processor       *event.DescriptionProcessor
	currencyService currency.Service
	playerService   player.Service
	cfg             kafka.Configuration

	consumers map[string]*kafka.Consumer[any]
}

func NewApi(logger *zap.Logger, cfg kafka.Configuration, currencyService currency.Service, playerService player.Service) *Api {
	return &Api{
		logger:          logger.Named("kafka-api"),
		processor:       event.NewDescriptionProcessor(),
		currencyService: currencyService,
		playerService:   playerService,
		cfg:             cfg,
	}
}

func (a *Api) StartConsumers(ctx context.Context) {
	// a.BaseConsumer(ctx)
	a.CurrencyConsumer(ctx, a.currencyService)
	a.PlayerConsumer(ctx, a.playerService)
	a.DescriptionEnrichmentConsumer(ctx, a.processor)
	a.LogConsumer(ctx)
}

// BaseConsumer creates a Kafka consumer
func (a *Api) BaseConsumer(ctx context.Context) {
	baseConsumer := kafka.NewConsumer[casino.Event](a.logger, a.cfg, "casino-event")
	a.consumers["base"] = (*kafka.Consumer[any])(baseConsumer)
	baseConsumer.Read(ctx, casino.Event{}, func(model casino.Event, err error) {
		if err != nil {
			a.logger.Error("Failed to read message", zap.Error(err))
			return
		}

		a.logger.Info("Received event", zap.Any("event", model))
	})
}

// CurrencyConsumer creates a Kafka consumer for the currency enrichment stage
func (a *Api) CurrencyConsumer(ctx context.Context, currencyService currency.Service) {
	currencyConsumer := kafka.NewConsumer[casino.Event](a.logger, a.cfg, "casino-event-currency")
	a.consumers["currency"] = (*kafka.Consumer[any])(currencyConsumer)
	currencyConsumer.Read(ctx, casino.Event{}, func(model casino.Event, err error) {
		if err != nil {
			a.logger.Error("Failed to read message", zap.Error(err))
			return
		}

		// Fetch currency details from the currency service
		conversion, err := currencyService.Convert(ctx, model.Currency, "EUR", model.Amount)
		if err != nil {
			a.logger.Error("Failed to fetch currency data", zap.Error(err))
			return
		}

		model.AmountEUR = conversion

		a.logger.Info("Received event", zap.Any("event", model))
	})
}

// PlayerConsumer creates a Kafka consumer for the player enrichment stage
func (a *Api) PlayerConsumer(ctx context.Context, playerService player.Service) {
	playerConsumer := kafka.NewConsumer[casino.Event](a.logger, a.cfg, "casino-event-player-data")
	a.consumers["player"] = (*kafka.Consumer[any])(playerConsumer)
	playerConsumer.Read(ctx, casino.Event{}, func(model casino.Event, err error) {
		if err != nil {
			a.logger.Error("Failed to read message", zap.Error(err))
			return
		}

		// Fetch player details from the player service
		playerWithId, err := playerService.GetPlayerDetails(ctx, model.PlayerID)
		if err != nil {
			a.logger.Error("Failed to fetch player data", zap.Error(err))
			return
		}

		model.Player = *playerWithId

		// todo forward to the next stage

		a.logger.Info("Received event", zap.Any("event", model))
	})
}

// DescriptionEnrichmentConsumer create a Kafka consumer for the description enrichment stage
func (a *Api) DescriptionEnrichmentConsumer(ctx context.Context, processor *event.DescriptionProcessor) {
	descriptionConsumer := kafka.NewConsumer[casino.Event](a.logger, a.cfg, "casino-event-description")
	a.consumers["description"] = (*kafka.Consumer[any])(descriptionConsumer)
	descriptionConsumer.Read(ctx, casino.Event{}, func(model casino.Event, err error) {
		if err != nil {
			a.logger.Error("Failed to read message", zap.Error(err))
			return
		}

		description, err := processor.Process(model)
		if err != nil {
			a.logger.Error("Failed to create a description", zap.Error(err))
			return
		}

		// Add a description to the event
		model.Description = description

		// todo forward to the next stage

		a.logger.Info("Added description to the event", zap.Any("event", model))
	})
}

// LogConsumer Create a Kafka consumer for the log stage
func (a *Api) LogConsumer(ctx context.Context) {
	logConsumer := kafka.NewConsumer[casino.Event](a.logger, a.cfg, "casino-event-log")
	a.consumers["log"] = (*kafka.Consumer[any])(logConsumer)
	logConsumer.Read(ctx, casino.Event{}, func(model casino.Event, err error) {
		if err != nil {
			a.logger.Error("Failed to read message", zap.Error(err))
			return
		}

		a.logger.Info("Received event", zap.Any("event", model))
	})
}

func (a *Api) Close() {
	for _, consumer := range a.consumers {
		consumer.Close()
	}
}