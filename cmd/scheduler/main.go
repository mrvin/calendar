package main

import (
	"context"
	"flag"
	"log"
	"log/slog"

	"github.com/mrvin/calendar/internal/config"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/queue"
	"github.com/mrvin/calendar/internal/queue/rabbitmq"
	"github.com/mrvin/calendar/internal/scheduler/app"
	"github.com/mrvin/calendar/internal/storage/postgresql"
)

//nolint:tagliatelle
type Config struct {
	Queue       queue.Conf      `yaml:"queue"`
	DB          postgresql.Conf `yaml:"db"`
	Logger      logger.Conf     `yaml:"logger"`
	SchedPeriod int             `yaml:"schedule_period"`
}

func main() {
	configFile := flag.String("config", "/etc/calendar/scheduler.yml", "path to configuration file")
	flag.Parse()

	var conf Config
	if err := config.Parse(*configFile, &conf); err != nil {
		log.Printf("Parse config: %v", err)
		return
	}

	// init logger
	logFile, err := logger.Init(&conf.Logger)
	if err != nil {
		log.Printf("Init logger: %v", err)
		return
	}
	slog.Info("Init logger", slog.String("level", conf.Logger.Level))
	defer func() {
		if err := logFile.Close(); err != nil {
			slog.Error("Close log file", slog.String("error", err.Error())) //nolint:gosec
		}
	}()

	ctx := context.Background()
	st, err := postgresql.New(ctx, &conf.DB)
	if err != nil {
		slog.Error("Failed to init storag: " + err.Error())
		return
	}
	defer st.Close()
	slog.Info("Connected to database")

	var qm rabbitmq.Queue

	url := rabbitmq.QueryBuildAMQP(&conf.Queue)

	if err := qm.ConnectAndCreate(url, conf.Queue.Name); err != nil {
		slog.Error("Failed to init queue: " + err.Error())
		return
	}
	defer qm.Close()
	slog.Info("Connected to queue")

	app := app.New(st, qm, conf.SchedPeriod)
	app.Run(ctx)
}
