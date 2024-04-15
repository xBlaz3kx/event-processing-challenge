package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/xBlaz3kx/event-processing-challenge/internal/currency"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/cache"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/http"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/observability"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player"
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

		// Currency service
		currencyCache := cache.NewRedisCache(cache.RedisConfiguration{}, logger)
		currencyService := currency.NewServiceV1(logger, currencyCache)

		// Player service
		playerService := player.NewServiceV1(logger)

		// HTTP server
		server := http.NewServer(logger, http.Configuration{}, currencyCache, currencyService, playerService)
		go server.Run(ctx)
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
