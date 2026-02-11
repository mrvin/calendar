package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/calendar-front/client"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/calendar-front/httpserver/handlers"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/pkg/http/logger"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Conf struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Server struct {
	http.Server
}

func New(conf *Conf, client client.Calendar) *Server {
	mux := http.NewServeMux()

	h := handlers.New(client)

	mux.HandleFunc(http.MethodGet+" /form-user", h.DisplayFormRegistration)
	mux.HandleFunc(http.MethodPost+" /create-user", h.Registration)

	mux.HandleFunc(http.MethodGet+" /form-login", h.DisplayFormLogin)
	mux.HandleFunc(http.MethodPost+" /login-user", h.Login)

	mux.HandleFunc(http.MethodGet+" /list-events/{accessToken}", h.DisplayListEventsForUser)

	mux.HandleFunc(http.MethodGet+" /form-event/{accessToken}", h.DisplayFormEvent)
	mux.HandleFunc(http.MethodPost+" /create-event/{accessToken}", h.CreateEvent)

	mux.HandleFunc(http.MethodGet+" /list-users/{accessToken}", h.DisplayListUsers)

	mux.HandleFunc(http.MethodGet+" /user/{userName}", h.DisplayUser)
	mux.HandleFunc(http.MethodGet+" /event/{id}", h.DisplayEvent)

	mux.HandleFunc(http.MethodGet+" /delete-user/{id}", h.DeleteUser)
	mux.HandleFunc(http.MethodGet+" /delete-event/{id}", h.DeleteEvent)

	loggerServer := logger.Logger{Inner: otelhttp.NewHandler(mux, "HTTP")}

	return &Server{
		//nolint:exhaustivestruct,exhaustruct
		http.Server{
			Addr:         fmt.Sprintf("%s:%d", conf.Host, conf.Port),
			Handler:      &loggerServer,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  1 * time.Minute,
		},
	}
}

func (s *Server) Start() error {
	slog.Info("Start http server: http://" + s.Addr)
	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("start http server: %w", err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	slog.Info("Stop http server")
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("stop http server: %w", err)
	}

	return nil
}
