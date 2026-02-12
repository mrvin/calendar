package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrvin/calendar/internal/config"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/queue"
	"github.com/mrvin/calendar/internal/queue/rabbitmq"
	sqlstorage "github.com/mrvin/calendar/internal/storage/sql"
)

//nolint:tagliatelle
type Config struct {
	Queue       queue.Conf      `yaml:"queue"`
	DB          sqlstorage.Conf `yaml:"db"`
	Logger      logger.Conf     `yaml:"logger"`
	SchedPeriod int             `yaml:"schedule_period"`
}

//nolint:gocognit,cyclop
func main() {
	ctx := context.Background()

	configFile := flag.String("config", "/etc/calendar/scheduler.yml", "path to configuration file")
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

	ctxInitStorag, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	st, err := sqlstorage.New(ctxInitStorag, &conf.DB)
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

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT /*(Control-C)*/, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	schedPeriod := time.Duration(conf.SchedPeriod) * time.Minute
	ticker := time.Tick(schedPeriod)
	for {
		ctxGetAllEvents, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		events, err := st.ListEvents(ctxGetAllEvents)
		if err != nil {
			slog.Error("List event: " + err.Error())
		}
		slog.Info("Start send event")
		for _, event := range events {
			if cancelled(ctx) {
				break
			}
			nowTime := time.Now()
			if event.StartTime.After(nowTime) && event.StartTime.Before(nowTime.Add(schedPeriod)) {
				user, err := st.GetUser(ctx, event.UserName)
				if err != nil {
					slog.Error(err.Error())
					continue
				}
				alertEvent := queue.AlertEvent{EventID: event.ID, Title: event.Title, Description: event.Description,
					StartTime: event.StartTime, UserName: user.Name, UserEmail: user.Email}

				byteAlertEvent, err := queue.EncodeAlertEvent(&alertEvent)
				if err != nil {
					slog.Error(err.Error())
					continue
				}

				ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
				defer cancel()
				if err := qm.SendMsg(ctx, byteAlertEvent); err != nil {
					slog.Error(err.Error())
					continue
				}
				slog.Info("Put alert message in queue", slog.Int64("Event id", event.ID))
			}
		}
		select {
		case <-ticker:
		// do nothing.
		case <-ctx.Done():
			slog.Info("Stop scheduler")
			return
		}
	}
}

func cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
