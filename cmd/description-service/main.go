package main

import (
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpHandler "github.com/xBlaz3kx/event-processing-challenge/internal/api/http"
	kafkaApi "github.com/xBlaz3kx/event-processing-challenge/internal/api/kafka"
	"github.com/xBlaz3kx/event-processing-challenge/internal/casino"
	"github.com/xBlaz3kx/event-processing-challenge/internal/currency"
	"github.com/xBlaz3kx/event-processing-challenge/internal/event"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/cache"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/configuration"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/http"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/kafka"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/observability"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player/database"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type serviceConfig struct {
	Kafka       kafka.Configuration        `yaml:"kafka"`
	Database    database.Configuration     `yaml:"database"`
	Cache       cache.RedisConfiguration   `yaml:"cache"`
	Http        http.Configuration         `yaml:"http"`
	ExchangeApi currency.ExchangeApiConfig `yaml:"exchangeApi"`
	KsqlDb      kafka.KsqlConfiguration    `yaml:"ksqlDb"`
}

var configPath string

var logCmd = &cobra.Command{
	Use:   "event-processing-challenge",
	Short: "Event Processing Challenge",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		logger := zap.L()
		logger.Info("Starting the service")

		// Fetch configuration
		config := &serviceConfig{}
		err := configuration.GetConfiguration(viper.GetViper(), config)
		if err != nil {
			logger.Fatal("Failed to get configuration", zap.Error(err))
		}

		logger.Info("Configuration loaded", zap.Any("config", config))

		// HTTP server
		server := http.NewServer(logger, http.Configuration{Address: "0.0.0.0:8080"})
		router := server.Router()
		handler := httpHandler.NewHandler(nil, nil, nil)

		// Add http routes
		router.POST("/event/description", handler.GetEventDescription)
		go server.Run(ctx)

		eventProcessor := event.NewDescriptionProcessor()

		// Create a producer for the next stage
		producer := kafka.NewProducer(logger, config.Kafka, kafkaApi.LogTopic)
		descriptionConsumer := kafka.NewConsumer[casino.Event](logger, config.Kafka, kafkaApi.DescriptionTopic)
		descriptionConsumer.Read(ctx, casino.Event{}, func(model casino.Event, err error) {
			if err != nil {
				logger.Error("Failed to read message", zap.Error(err))
				return
			}

			logger.Info("Added description to the event", zap.Any("event", model))

			description, err := eventProcessor.Process(model)
			if err != nil {
				logger.Error("Failed to create a description", zap.Error(err))
				return
			}

			// Add a description to the event
			model.Description = description

			err = producer.Publish(ctx, model)
			if err != nil {
				logger.Error("Failed to publish message", zap.Error(err))
			}
		})
		defer descriptionConsumer.Close()

		<-ctx.Done()
		logger.Info("Shutting down the service")
	},
	Version: "v0.1.0",
}

func init() {
	cobra.OnInitialize(setupLogger, setupConfig)
}

func main() {
	// Setup flag - config binding
	logCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file (default to $HOME/service/ or /usr/service/config)")
	_ = viper.BindPFlag("config", logCmd.PersistentFlags().Lookup("config"))

	err := logCmd.Execute()
	if err != nil {
		zap.L().Fatal("Failed to execute command", zap.Error(err))
	}
}

func setupLogger() {
	logger := observability.NewLogger("debug")
	zap.ReplaceGlobals(logger)
}

func setupConfig() {
	configuration.InitConfig(configPath, "service", "$HOME/service/", "/usr/service/config")
	zap.L().Info("Configuration set up", zap.Any("cfg", viper.AllSettings()), zap.String("configPath", configPath))
}
