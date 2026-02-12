package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/mrvin/calendar/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

func (q *Queue) ConnectAndCreate(url, name string) error {
	var err error

	q.conn, err = amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	q.ch, err = q.conn.Channel()
	if err != nil {
		return fmt.Errorf("connect channel: %w", err)
	}

	q.queue, err = q.ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("create queue: %w", err)
	}

	return nil
}

func (q *Queue) SendMsg(ctx context.Context, body []byte) error {
	err := q.ch.PublishWithContext(ctx,
		"",           // exchange
		q.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			Body: body,
		})
	if err != nil {
		return err
	}
	return nil
}

func (q *Queue) GetConsumeChan() (<-chan amqp.Delivery, error) {
	return q.ch.Consume(
		q.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
}

func (q *Queue) Close() error {
	err := q.conn.Close()
	if err != nil {
		log.Printf("close connect: %v", err)
	}

	err = q.ch.Close()

	return err
}

func QueryBuildAMQP(conf *queue.Conf) string {
	query := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(conf.UserName, conf.Password),
		Host:   fmt.Sprintf("%s:%d", conf.Host, conf.Port),
	}

	return query.String()
}
