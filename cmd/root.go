package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpHandler "github.com/xBlaz3kx/event-processing-challenge/internal/api/http"
	kafkaApi "github.com/xBlaz3kx/event-processing-challenge/internal/api/kafka"
	"github.com/xBlaz3kx/event-processing-challenge/internal/currency"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/cache"
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
	Kafka           kafka.Configuration      `yaml:"kafka"`
	Database        database.Configuration   `yaml:"database"`
	Cache           cache.RedisConfiguration `yaml:"cache"`
	Http            http.Configuration       `yaml:"http"`
	ExchangeRateUrl string                   `yaml:"exchangeRateUrl"`
}

var configPath string

var rootCmd = &cobra.Command{
	Use:   "event-processing-challenge",
	Short: "Event Processing Challenge",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		logger := observability.NewLogger("debug")
		logger.Info("Starting the service")

		// Fetch configuration
		config := &serviceConfig{}
		err := configuration.GetConfiguration(viper.GetViper(), config)
		if err != nil {
			logger.Fatal("Failed to get configuration", zap.Error(err))
		}

		logger.Info("Configuration loaded", zap.Any("config", config))

		// Currency service
		currencyCache := cache.NewRedisCache(config.Cache, logger)
		currencyService := currency.NewServiceV1(logger, currencyCache, config.ExchangeRateUrl)

		// Player service
		playerRepository := database.NewPostgresPlayerRepository(logger, config.Database)
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

		// Create Kafka API
		kafkaApi := kafkaApi.NewApi(logger, config.Kafka, currencyService, playerService)
		kafkaApi.StartConsumers(ctx)
		defer kafkaApi.Close()

		<-ctx.Done()
		logger.Info("Shutting down the service")
	},
	Version: "v0.1.0",
}

func Execute() {
	cobra.OnInitialize(onInit)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file (default to $HOME/service/ or /usr/service/config)")
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func onInit() {
	configuration.SetupEnv("service")
	configuration.InitConfig(configPath, "$HOME/service/", "/usr/service/config")
}
