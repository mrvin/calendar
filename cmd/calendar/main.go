//go:generate protoc -I=../../api/ --go_out=../../internal/grpcapi --go-grpc_out=require_unimplemented_servers=false:../../internal/grpcapi ../../api/event_service.proto
//go:generate protoc -I=../../api/ --go_out=../../internal/grpcapi --go-grpc_out=require_unimplemented_servers=false:../../internal/grpcapi ../../api/user_service.proto
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"sync"
	"time"

	authservice "github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/calendar/httpserver"
	"github.com/mrvin/calendar/internal/config"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/metric"
	"github.com/mrvin/calendar/internal/storage"
	"github.com/mrvin/calendar/internal/storage/memory"
	"github.com/mrvin/calendar/internal/storage/postgresql"
)

const serviceName = "Calendar"
const ctxTimeout = 2 // in second

type Config struct {
	InMem bool            `yaml:"inmemory"`
	DB    postgresql.Conf `yaml:"db"`
	HTTP  httpserver.Conf `yaml:"http"`
	//	GRPC   grpcserver.Conf  `yaml:"grpc"`
	Logger logger.Conf      `yaml:"logger"`
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
		storage = memory.New()
	} else {
		var err error
		slog.Info("Storage in sql database")
		storage, err = postgresql.New(ctx, &conf.DB)
		if err != nil {
			slog.Error("Failed to init storage: " + err.Error())
			return
		}
		defer func() {
			if storageSQL, ok := storage.(*postgresql.Storage); ok {
				storageSQL.Close()
				slog.Info("Closing the database connection")
			}
		}()
		slog.Info("Connected to database")
	}

	authService := authservice.New(storage, &conf.Auth)
	serverHTTP := httpserver.New(&conf.HTTP, storage, authService)
	//	serverGRPC, err := grpcserver.New(&conf.GRPC, authService, eventService)
	//	if err != nil {
	//		slog.Error("Failed to init gRPC server: " + err.Error())
	//		return
	//	}

	var wg sync.WaitGroup
	wg.Go(func() {
		serverHTTP.Run(ctx)
	})
	wg.Go(func() {
		/*
			if err := serverGRPC.Start(); err != nil {
				slog.Error("Failed to start gRPC server: " + err.Error())
				return
			}
		*/
	})
	wg.Wait()

	slog.Info("Stop service " + serviceName)
}
