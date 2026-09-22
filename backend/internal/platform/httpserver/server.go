package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/platform/config"
)

// Server owns the HTTP listener lifecycle and graceful shutdown policy.
type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

// New creates a configured HTTP server without starting its listener.
func New(configuration config.HTTPServer, handler http.Handler) (*Server, error) {
	if handler == nil {
		return nil, errors.New("HTTP handler is required")
	}

	if err := configuration.Validate(); err != nil {
		return nil, fmt.Errorf("validate HTTP server configuration: %w", err)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              configuration.Address(),
			Handler:           handler,
			ReadHeaderTimeout: configuration.ReadTimeout,
			ReadTimeout:       configuration.ReadTimeout,
			WriteTimeout:      configuration.WriteTimeout,
			IdleTimeout:       configuration.IdleTimeout,
		},
		shutdownTimeout: configuration.ShutdownTimeout,
	}, nil
}

// Run serves requests until the context is cancelled or the listener fails.
func (s *Server) Run(ctx context.Context) error {
	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-serveErrors:
		return normalizeServeError(err)

	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(
			context.Background(),
			s.shutdownTimeout,
		)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownContext); err != nil {
			_ = s.httpServer.Close()
			return fmt.Errorf("shut down HTTP server: %w", err)
		}

		return normalizeServeError(<-serveErrors)
	}
}

func normalizeServeError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return fmt.Errorf("serve HTTP: %w", err)
}
