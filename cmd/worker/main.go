package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/projectdiscovery/nuclei-platform/internal/config"
	"github.com/projectdiscovery/nuclei-platform/internal/worker"
	"github.com/rs/zerolog"
)

func main() {
	configPath := flag.String("config", "configs/worker.yaml", "path to config file")
	flag.Parse()

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.LoadWorkerConfig(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	switch cfg.Log.Level {
	case "debug":
		log = log.Level(zerolog.DebugLevel)
	case "warn":
		log = log.Level(zerolog.WarnLevel)
	case "error":
		log = log.Level(zerolog.ErrorLevel)
	default:
		log = log.Level(zerolog.InfoLevel)
	}

	w, err := worker.New(cfg, &log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create worker")
	}
	defer w.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := w.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("worker exited with error")
	}
	log.Info().Msg("worker stopped")
}
