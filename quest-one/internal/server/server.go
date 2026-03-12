// Package server implements the chi HTTP server.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/quest-one/quest-one/internal/application"
)

// Config holds server startup configuration.
type Config struct {
	Port int
	Log  *slog.Logger
}

// Server wraps an http.Server with graceful shutdown.
type Server struct {
	http *http.Server
	log  *slog.Logger
}

// New constructs a Server with all routes registered.
func New(app *application.App, cfg Config) *Server {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RealIP)
	r.Use(requestLogger(cfg.Log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Register all routes
	registerRoutes(r, app, cfg.Log)

	return &Server{
		http: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		log: cfg.Log,
	}
}

// Start begins listening and blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("server starting", slog.String("addr", s.http.Addr))
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.log.Info("server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.http.Shutdown(shutdownCtx)
	}
}

// requestLogger returns a slog-based request logging middleware.
func requestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			log.Info("http",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Duration("latency", time.Since(start)),
			)
		})
	}
}
