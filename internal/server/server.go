package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/projectdiscovery/nuclei-platform/internal/config"
	"github.com/projectdiscovery/nuclei-platform/internal/server/handler"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Server struct {
	cfg      config.AppConfig
	db       *gorm.DB
	nats     *NATSManager
	log      *zerolog.Logger
	server   *http.Server
	frontend http.FileSystem
}

func New(cfg config.AppConfig, db *gorm.DB, natsMgr *NATSManager, log *zerolog.Logger, frontend http.FileSystem) *Server {
	return &Server{cfg: cfg, db: db, nats: natsMgr, log: log, frontend: frontend}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/healthz"))

	scanH := handler.NewScanHandler(s.db, s.nats, s.log)
	resultH := handler.NewResultHandler(s.db)
	workerH := handler.NewWorkerHandler(s.db)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/scans", func(r chi.Router) {
			r.Post("/", scanH.Create)
			r.Get("/", scanH.List)
			r.Get("/{id}", scanH.Get)
			r.Delete("/{id}", scanH.Cancel)
			r.Post("/{id}/retry", scanH.Retry)
			r.Get("/{id}/results", resultH.ListByTask)
		})
		r.Route("/results", func(r chi.Router) {
			r.Get("/", resultH.ListGlobal)
			r.Get("/stats", resultH.Stats)
		})
		r.Route("/workers", func(r chi.Router) {
			r.Get("/", workerH.List)
			r.Get("/{id}", workerH.Get)
			r.Post("/{id}/disable", workerH.Disable)
			r.Post("/{id}/enable", workerH.Enable)
		})
	})

	// Serve frontend
	if s.frontend != nil {
		fileServer := http.FileServer(s.frontend)
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Check if file exists
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			if f, err := s.frontend.Open(path); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
			// SPA fallback
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	}

	return r
}

func (s *Server) Start(ctx context.Context) error {
	// Start NATS subscriptions
	if err := s.nats.SubscribeResults(ctx); err != nil {
		return fmt.Errorf("subscribe results: %w", err)
	}
	if err := s.nats.SubscribeHeartbeats(ctx); err != nil {
		return fmt.Errorf("subscribe heartbeats: %w", err)
	}

	// Start background detectors
	s.nats.StartStaleWorkerDetector(ctx, s.cfg.Worker.HeartbeatInterval, s.cfg.Worker.OfflineThreshold)
	s.nats.StartTaskTimeoutDetector(ctx, s.cfg.Worker.HeartbeatInterval, s.cfg.Worker.MaxTaskDuration)

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.Router(),
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
	}

	s.log.Info().Str("addr", addr).Msg("starting HTTP server")
	go func() {
		<-ctx.Done()
		s.log.Info().Msg("shutting down HTTP server")
		s.server.Shutdown(context.Background())
	}()

	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
