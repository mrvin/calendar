//go:generate protoc -I=../../api/ --go_out=../../internal/grpcapi --go-grpc_out=require_unimplemented_servers=false:../../internal/grpcapi ../../api/event_service.proto
//go:generate protoc -I=../../api/ --go_out=../../internal/grpcapi --go-grpc_out=require_unimplemented_servers=false:../../internal/grpcapi ../../api/user_service.proto
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	authservice "github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/calendar/grpcserver"
	"github.com/mrvin/calendar/internal/calendar/httpserver"
	eventservice "github.com/mrvin/calendar/internal/calendar/service/event"
	"github.com/mrvin/calendar/internal/config"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/metric"
	"github.com/mrvin/calendar/internal/storage"
	memorystorage "github.com/mrvin/calendar/internal/storage/memory"
	sqlstorage "github.com/mrvin/calendar/internal/storage/sql"
	"github.com/mrvin/calendar/internal/tracer"
)

const serviceName = "Calendar"
const ctxTimeout = 2 // in second
const numServer = 2  // HTTP and gRPC

type Config struct {
	InMem  bool             `yaml:"inmemory"`
	DB     sqlstorage.Conf  `yaml:"db"`
	HTTP   httpserver.Conf  `yaml:"http"`
	GRPC   grpcserver.Conf  `yaml:"grpc"`
	Logger logger.Conf      `yaml:"logger"`
	Tracer tracer.Conf      `yaml:"tracer"`
	Metric metric.Conf      `yaml:"metrics"`
	Auth   authservice.Conf `yaml:"auth"`
}

//nolint:gocognit,cyclop
func main() {
	ctx := context.Background()

	configFile := flag.String("config", "/etc/calendar/calendar.yml", "path to configuration file")
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
	slog.Info("Init logger", slog.String("Logging level", conf.Logger.Level))
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

	var storage storage.Storage
	//nolint:nestif
	if conf.InMem {
		slog.Info("Storage in memory")
		storage = memorystorage.New()
	} else {
		var err error
		slog.Info("Storage in sql database")
		storage, err = sqlstorage.New(ctx, &conf.DB)
		if err != nil {
			slog.Error("Failed to init storage: " + err.Error())
			return
		}
		defer func() {
			if storageSQL, ok := storage.(*sqlstorage.Storage); ok {
				if err := storageSQL.Close(); err != nil {
					slog.Error("Failed to close storage: " + err.Error())
				} else {
					slog.Info("Closing the database connection")
				}
			}
		}()
		slog.Info("Connected to database")
	}

	authService := authservice.New(storage, &conf.Auth)
	eventService := eventservice.New(storage)
	serverHTTP := httpserver.New(&conf.HTTP, authService, eventService)
	serverGRPC, err := grpcserver.New(&conf.GRPC, authService, eventService)
	if err != nil {
		slog.Error("Failed to init gRPC server: " + err.Error())
		return
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT /*(Control-C)*/, syscall.SIGTERM)
	go listenForShutdown(signals, serverGRPC)

	var wg sync.WaitGroup
	wg.Go(func() {
		serverHTTP.Run(ctx)
	})
	wg.Go(func() {
		if err := serverGRPC.Start(); err != nil {
			slog.Error("Failed to start gRPC server: " + err.Error())
			return
		}
	})
	wg.Wait()

	slog.Info("Stop service " + serviceName)
}

func listenForShutdown(
	signals chan os.Signal,
	serverGRPC *grpcserver.Server,
) {
	<-signals
	signal.Stop(signals)

	serverGRPC.Stop()
}
