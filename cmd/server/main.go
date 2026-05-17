package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/projectdiscovery/nuclei-platform/internal/config"
	"github.com/projectdiscovery/nuclei-platform/internal/model"
	"github.com/projectdiscovery/nuclei-platform/internal/server"
	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed all:web/dist
var frontendFS embed.FS

func main() {
	configPath := flag.String("config", "configs/server.yaml", "path to config file")
	flag.Parse()

	// Logger
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Config
	cfg, err := config.LoadServerConfig(*configPath)
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

	// Database
	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}
	if err := db.AutoMigrate(&model.Task{}, &model.Result{}, &model.Worker{}); err != nil {
		log.Fatal().Err(err).Msg("failed to migrate database")
	}
	log.Info().Str("dsn", cfg.Database.DSN).Msg("database initialized")

	// NATS
	natsMgr, err := server.NewNATSManager(cfg.NATS, db, &log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect NATS")
	}
	defer natsMgr.Close()

	// Frontend FS (skip in dev mode, use Vite dev server instead)
	var frontend http.FileSystem
	if os.Getenv("NUCLEI_PLATFORM_DEV") != "true" {
		sub, err := fs.Sub(frontendFS, "web/dist")
		if err != nil {
			log.Warn().Err(err).Msg("failed to load embedded frontend")
		} else {
			frontend = http.FS(sub)
		}
	}

	// Context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start server
	srv := server.New(*cfg, db, natsMgr, &log, frontend)
	if err := srv.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("server exited with error")
	}
	log.Info().Msg("server stopped")
}
