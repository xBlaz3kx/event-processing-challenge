package main

import (
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpHandler "github.com/xBlaz3kx/event-processing-challenge/internal/api/http"
	kafkaApi "github.com/xBlaz3kx/event-processing-challenge/internal/api/kafka"
	"github.com/xBlaz3kx/event-processing-challenge/internal/casino"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/configuration"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/http"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/kafka"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/observability"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player/database"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type serviceConfig struct {
	Kafka    kafka.Configuration    `yaml:"kafka"`
	Database database.Configuration `yaml:"database"`
	Http     http.Configuration     `yaml:"http"`
}

var configPath string

var playerCmd = &cobra.Command{
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

		// Player service
		playerRepository := database.NewPostgresPlayerRepository(logger, config.Database)
		defer playerRepository.Close()
		playerService := player.NewServiceV1(logger, playerRepository)

		// HTTP server
		server := http.NewServer(logger, http.Configuration{Address: "0.0.0.0:8080"}, playerService, playerRepository)
		router := server.Router()

		handler := httpHandler.NewHandler(nil, playerService, nil)

		// Add http routes
		router.GET("/player/:id", handler.GetPlayerDetails)
		go server.Run(ctx)

		// Create a producer for the next stage
		producer := kafka.NewProducer(logger, config.Kafka, kafkaApi.DescriptionTopic)

		// Create Kafka API
		playerConsumer := kafka.NewConsumer[casino.Event](logger, config.Kafka, kafkaApi.PlayerDataTopic)
		playerConsumer.Read(ctx, casino.Event{}, func(model casino.Event, err error) {
			if err != nil {
				logger.Error("Failed to read message", zap.Error(err))
				return
			}
			logger.Info("Received Player event", zap.Any("event", model))

			// Fetch player details from the player service
			playerWithId, err := playerService.GetPlayerDetails(ctx, model.PlayerID)
			if err != nil {
				logger.Error("Failed to fetch player data", zap.Error(err))
			}

			if playerWithId != nil {
				model.Player = *playerWithId
			}

			err = producer.Publish(ctx, model)
			if err != nil {
				logger.Error("Failed to publish message", zap.Error(err))
			}
		})
		defer playerConsumer.Close()

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
	playerCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file (default to $HOME/service/ or /usr/service/config)")
	_ = viper.BindPFlag("config", playerCmd.PersistentFlags().Lookup("config"))

	err := playerCmd.Execute()
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
