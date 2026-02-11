package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/calendar-front/client"
	grpcclient "github.com/mrvin/hw-otus-go/hw12-15calendar/internal/calendar-front/client/grpc"
	httpclient "github.com/mrvin/hw-otus-go/hw12-15calendar/internal/calendar-front/client/http"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/calendar-front/httpserver"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/config"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/logger"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/metric"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/tracer"
)

const serviceName = "Calendar-ws"
const ctxTimeout = 2 // in second

//nolint:tagliatelle
type Config struct {
	Client     string          `yaml:"client"`
	HTTPClient httpclient.Conf `yaml:"http-client"`
	GRPCClient grpcclient.Conf `yaml:"grpc-client"`
	HTTP       httpserver.Conf `yaml:"http"`
	Logger     logger.Conf     `yaml:"logger"`
	Tracer     tracer.Conf     `yaml:"tracer"`
	Metric     metric.Conf     `yaml:"metrics"`
}

//nolint:gocognit,cyclop
func main() {
	ctx := context.Background()

	configFile := flag.String("config", "/etc/calendar/calendar-front.yml", "path to configuration file")
	flag.Parse()

	var conf Config
	if err := config.Parse(*configFile, &conf); err != nil {
		log.Printf("Parse config: %v", err)
		return
	}

	logFile, err := logger.Init(&conf.Logger)
	if err != nil {
		log.Printf("Init logger: %v\n", err)
		return
	}
	slog.Info("Init logger")
	defer func() {
		if err := logFile.Close(); err != nil {
			slog.Error("Close log file: " + err.Error())
		}
	}()

	if conf.Tracer.Enable {
		ctxTracer, cancel := context.WithTimeout(ctx, ctxTimeout*time.Second)
		defer cancel()
		tp, err := tracer.Init(ctxTracer, &conf.Tracer, serviceName)
		if err != nil {
			slog.Warn("Failed to init tracer: " + err.Error())
		} else {
			slog.Info("Init tracer")
			defer func() {
				if err := tp.Shutdown(ctx); err != nil {
					slog.Error("Failed to shutdown tracer: " + err.Error())
				}
			}()
		}
	}

	if conf.Metric.Enable {
		ctxMetric, cancel := context.WithTimeout(ctx, ctxTimeout*time.Second)
		defer cancel()
		mp, err := metric.Init(ctxMetric, &conf.Metric, serviceName)
		if err != nil {
			slog.Warn("Failed to init metric: " + err.Error())
		} else {
			slog.Info("Init metric")
			defer func() {
				if err := mp.Shutdown(ctx); err != nil {
					slog.Error("Failed to shutdown metric: " + err.Error())
				}
			}()
		}
	}

	var client client.Calendar
	if conf.Client == "grpc" { //nolint:nestif
		client, err = grpcclient.New(ctx, &conf.GRPCClient)
		if err != nil {
			slog.Error("Failed to init gRPC client: " + err.Error())
			return
		}
		slog.Info("Connect to gRPC server")
		defer func() {
			if clientGRPC, ok := client.(*grpcclient.Client); ok {
				if err := clientGRPC.Close(); err != nil {
					slog.Error("Failed to close gRPC connect: " + err.Error())
				} else {
					slog.Info("Closing the gRPC connection")
				}
			}
		}()
	}
	serverHTTP := httpserver.New(&conf.HTTP, client)

	if err := serverHTTP.Start(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start http server: " + err.Error())
			return
		}
	}

	slog.Info("Stop service " + serviceName)
}
