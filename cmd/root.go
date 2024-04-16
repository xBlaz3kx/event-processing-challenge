package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	httpHandler "github.com/xBlaz3kx/event-processing-challenge/internal/api/http"
	"github.com/xBlaz3kx/event-processing-challenge/internal/casino"
	"github.com/xBlaz3kx/event-processing-challenge/internal/currency"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/cache"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/http"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/kafka"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/observability"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player/database"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var rootCmd = &cobra.Command{
	Use:   "event-processing-challenge",
	Short: "Event Processing Challenge",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		logger := observability.NewLogger("debug")
		logger.Info("Starting the service")

		// todo fetch from cfg file/env
		kafkaCfg := kafka.Configuration{
			Connection: "kafka:9092",
		}

		dbCfg := database.Configuration{DSN: "postgres://casino:casino@postgres:5432/casino?sslmode=disable"}

		cacheCfg := cache.RedisConfiguration{
			Address:  "redis:6379",
			Password: "",
			DB:       0,
		}

		// Currency service
		currencyCache := cache.NewRedisCache(cacheCfg, logger)
		currencyService := currency.NewServiceV1(logger, currencyCache)

		// Player service
		playerRepository := database.NewPostgresPlayerRepository(logger, dbCfg)
		defer playerRepository.Close()
		playerService := player.NewServiceV1(logger, playerRepository)

		// HTTP server
		server := http.NewServer(logger, http.Configuration{Address: "0.0.0.0:8080"}, currencyCache, currencyService, playerService, playerRepository)
		router := server.Router()

		handler := httpHandler.NewHandler(currencyService, playerService)

		// Add http routes
		router.GET("/event/player/:id", handler.GetPlayerDetails)
		router.POST("/event/currency", handler.GetEurCurrency)
		router.POST("/event/description", handler.GetEventDescription)
		router.POST("/materialize", handler.Materialize)

		go server.Run(ctx)

		// Create a Kafka consumer for the log stage
		logConsumer := kafka.NewConsumer[casino.Event](logger, kafkaCfg, "casino-event-log")
		logConsumer.Read(ctx, casino.Event{}, func(model casino.Event, err error) {
			if err != nil {
				logger.Error("Failed to read message", zap.Error(err))
				return
			}

			logger.Info("Received event", zap.Any("event", model))
		})

		<-ctx.Done()
		logger.Info("Shutting down the service")
	},
	Version: "v0.1.0",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
