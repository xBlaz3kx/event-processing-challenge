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
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type serviceConfig struct {
	Kafka  kafka.Configuration     `yaml:"kafka"`
	Http   http.Configuration      `yaml:"http"`
	KsqlDb kafka.KsqlConfiguration `yaml:"ksqlDb"`
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

		// Casino service
		ksqlClient := kafka.NewClientV1(logger, config.KsqlDb)
		defer ksqlClient.Close()
		casinoService := casino.NewServiceV1(logger, ksqlClient)

		// HTTP server
		server := http.NewServer(logger, http.Configuration{Address: "0.0.0.0:8080"})
		router := server.Router()

		handler := httpHandler.NewHandler(nil, nil, casinoService)
		router.POST("/materialize", handler.Materialize)
		go server.Run(ctx)

		// Create Kafka API
		logConsumer := kafka.NewConsumer[casino.Event](logger, config.Kafka, kafkaApi.LogTopic)
		logConsumer.Read(ctx, casino.Event{}, func(model casino.Event, err error) {
			if err != nil {
				logger.Error("Failed to read message", zap.Error(err))
				return
			}

			logger.Info("Logging event", zap.Any("event", model))
		})
		defer logConsumer.Close()

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
