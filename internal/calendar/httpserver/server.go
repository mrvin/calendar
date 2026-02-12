package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrvin/calendar/internal/calendar/httpserver/handlers"
	authservice "github.com/mrvin/calendar/internal/calendar/service/auth"
	eventservice "github.com/mrvin/calendar/internal/calendar/service/event"
	"github.com/mrvin/calendar/pkg/http/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const readTimeout = 5   // in second
const writeTimeout = 10 // in second
const idleTimeout = 1   // in minute

//nolint:tagliatelle
type ConfTLS struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

//nolint:tagliatelle
type Conf struct {
	Host  string  `yaml:"host"`
	Port  string  `yaml:"port"`
	IsTLS bool    `yaml:"is_https"`
	TLS   ConfTLS `yaml:"https"`
}

type Server struct {
	http.Server

	conf *Conf
}

func New(conf *Conf, auth *authservice.AuthService, events *eventservice.EventService) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc(http.MethodPost+" /signup", handlers.NewSignup(auth))
	mux.HandleFunc(http.MethodGet+" /login", handlers.NewSignup(auth))

	mux.HandleFunc(http.MethodGet+" /user", auth.Authorized(handlers.NewGetUser(auth)))
	mux.HandleFunc(http.MethodPut+" /user", auth.Authorized(handlers.NewUpdateUser(auth)))
	mux.HandleFunc(http.MethodDelete+" /user", auth.Authorized(handlers.NewDeleteUser(auth)))

	mux.HandleFunc(http.MethodPost+" /event", auth.Authorized(handlers.NewCreateEvent(events)))
	mux.HandleFunc(http.MethodGet+" /event/{id}", auth.Authorized(handlers.NewGetEvent(events)))
	mux.HandleFunc(http.MethodPut+" /event", auth.Authorized(handlers.NewUpdateEvent(events)))
	mux.HandleFunc(http.MethodDelete+" /event/{id}", auth.Authorized(handlers.NewDeleteEvent(events)))

	loggerServer := logger.Logger{Inner: otelhttp.NewHandler(mux, "HTTP")}

	return &Server{
		//nolint:exhaustruct
		http.Server{
			Addr:         net.JoinHostPort(conf.Host, conf.Port),
			Handler:      &loggerServer,
			ReadTimeout:  readTimeout * time.Second,
			WriteTimeout: writeTimeout * time.Second,
			IdleTimeout:  idleTimeout * time.Minute,
		},
		conf,
	}
}

func (s *Server) Run(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(
		ctx,
		os.Interrupt,    // SIGINT, (Control-C)
		syscall.SIGTERM, // systemd
		syscall.SIGQUIT,
	)

	go func() {
		defer cancel()
		if s.conf.IsTLS {
			if err := s.ListenAndServeTLS(s.conf.TLS.CertFile, s.conf.TLS.KeyFile); !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Failed to start https server: " + err.Error())
				return
			}
		} else {
			if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Failed to start http server: " + err.Error())
				return
			}
		}
	}()
	if s.conf.IsTLS {
		slog.Info("Start https server: https://" + s.Addr)
	} else {
		slog.Info("Start http server: http://" + s.Addr)
	}

	<-ctx.Done()

	if err := s.Shutdown(ctx); err != nil {
		slog.Error("Failed to stop http server: " + err.Error())
		return
	}
	if s.conf.IsTLS {
		slog.Info("Stop https server")
	} else {
		slog.Info("Stop http server")
	}

}
