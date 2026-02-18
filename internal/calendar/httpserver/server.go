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

	authservice "github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/calendar/httpserver/handlers"
	"github.com/mrvin/calendar/internal/storage"
	"github.com/mrvin/calendar/pkg/http/logger"
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

func New(conf *Conf, st storage.Storage, auth *authservice.Auth) *Server {
	mux := http.NewServeMux()

	// info
	mux.HandleFunc(http.MethodGet+" /api/health", handlers.Health)
	mux.HandleFunc(http.MethodGet+" /api/info", handlers.ErrorHandler("Info", handlers.Info))

	// Users
	mux.HandleFunc(http.MethodPost+" /api/auth/register", handlers.ErrorHandler("Register", handlers.NewRegister(st)))
	mux.HandleFunc(http.MethodPost+" /api/auth/login", handlers.ErrorHandler("Login", handlers.NewLogin(auth)))
	mux.HandleFunc(http.MethodGet+" /api/auth/me", auth.Authorized(handlers.ErrorHandler("Get user", handlers.NewGetUser(st))))
	mux.HandleFunc(http.MethodDelete+" /api/auth/me", auth.Authorized(handlers.ErrorHandler("Delete user", handlers.NewDeleteUser(st))))
	//	mux.HandleFunc(http.MethodPost+" /api/auth/refresh")
	//	mux.HandleFunc(http.MethodPost+" /api/auth/logout")

	// Events
	mux.HandleFunc(http.MethodPost+" /api/events", auth.Authorized(handlers.ErrorHandler("Create event", handlers.NewCreateEvent(st))))
	mux.HandleFunc(http.MethodGet+" /api/events/{id}", auth.Authorized(handlers.ErrorHandler("Get event", handlers.NewGetEvent(st))))
	mux.HandleFunc(http.MethodGet+" /api/events", auth.Authorized(handlers.ErrorHandler("List events", handlers.NewListEvents(st))))
	mux.HandleFunc(http.MethodPut+" /api/events/{id}", auth.Authorized(handlers.ErrorHandler("Update event", handlers.NewUpdateEvent(st))))
	mux.HandleFunc(http.MethodDelete+" /api/events/{id}", auth.Authorized(handlers.ErrorHandler("Delete event", handlers.NewDeleteEvent(st))))

	loggerServer := logger.Logger{Inner: mux}

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
