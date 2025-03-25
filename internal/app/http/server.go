package httpapp

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	authhttp "sso/internal/http/auth"
)

type Srv struct {
	log        *slog.Logger
	httpServer *http.Server
	addr       int
}

func New(log *slog.Logger, handlers *authhttp.Handler, port int) *Srv {
	log.Info("starting http server")

	mux := http.NewServeMux()

	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/register", handlers.RegisterHandler)
	mux.HandleFunc("/isadmin", handlers.IsAdminHandler)

	return &Srv{log: log, httpServer: &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}, addr: port}
}

func (s *Srv) MustRun() {
	if err := s.Run(); err != nil {
		panic(err)
	}
}

func (s *Srv) Run() error {
	const op = "httpapp.Run"

	log := s.log.With(
		slog.String("op", op),
		slog.Int("addr", s.addr),
	)

	log.Info("http server starting")

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Srv) Stop() {
	const op = "httpapp.Shutdown"
	log := s.log.With(slog.String("op", op))
	log.Info("http server shutting down")

	err := s.httpServer.Close()
	if err != nil {
		log.Error("failed to stop HTTP server", slog.String("error", err.Error()))
	}

}
