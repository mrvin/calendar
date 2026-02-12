package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/mrvin/calendar/internal/config"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/queue"
	"github.com/mrvin/calendar/internal/queue/rabbitmq"
	"github.com/mrvin/calendar/internal/sender/email"
)

type Config struct {
	Queue  queue.Conf  `yaml:"queue"`
	Email  email.Conf  `yaml:"email"`
	Logger logger.Conf `yaml:"logger"`
}

//nolint:cyclop
func main() {
	ctx := context.Background()

	configFile := flag.String("config", "/etc/calendar/sender.yml", "path to configuration file")
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

	var qm rabbitmq.Queue

	url := rabbitmq.QueryBuildAMQP(&conf.Queue)

	if err := qm.ConnectAndCreate(url, conf.Queue.Name); err != nil {
		slog.Error("New queue connection: " + err.Error())
		return
	}
	defer qm.Close()
	slog.Info("Сonnected to queue")

	chConsume, err := qm.GetConsumeChan()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT /*(Control-C)*/, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	for {
		select {
		case msg, ok := <-chConsume:
			if !ok {
				return
			}
			alertEvent, err := queue.DecodeAlertEvent(msg.Body)
			if err != nil {
				slog.Error(err.Error())
				continue
			}
			slog.Info("Take alert message from queue", slog.Int64("Event id", alertEvent.EventID))
			emailMsg := email.Message{To: alertEvent.UserEmail, Subject: alertEvent.Title, Description: alertEvent.Description}
			sendEvent(&conf.Email, &emailMsg)
		case <-ctx.Done():
			slog.Info("Stop sender")
			return
		}
	}
}

func sendEvent(conf *email.Conf, msg *email.Message) {
	if err := email.Alert(conf, msg); err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info(msg.Subject + "'%s' event notification sent to '%s'\n" + msg.To)
}
