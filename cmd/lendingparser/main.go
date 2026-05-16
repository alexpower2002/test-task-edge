package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"test-task-edge/internal/app/lendingparser"
	"test-task-edge/internal/logger"
)

type rrr interface {
	Register() error
	Resolve(ctx context.Context) error
	Release() error
}

func handleSignals(app rrr, cancel context.CancelFunc) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	for range done {
		log.Info().Msg("shutting server down...")
		cancel()
		if err := app.Release(); err != nil {
			log.Error().Err(err).Msg("an error occurred during server shutdown")
			os.Exit(1)
		}
		log.Info().Msg("server exited successfully")
		os.Exit(0)
	}
}

func main() {
	log.Logger = *logger.NewLogger("lendingparser", "info")

	ctx, cancel := context.WithCancel(context.Background())
	var app rrr
	app = lendingparser.NewApp()

	go handleSignals(app, cancel)

	if err := app.Register(); err != nil {
		log.Error().Err(err).Msg("register failed")
		os.Exit(1)
	}

	log.Info().Msg("initialized successfully")
	if err := app.Resolve(ctx); err != nil {
		log.Error().Err(err).Msg("resolve failed")
		os.Exit(1)
	}
}
