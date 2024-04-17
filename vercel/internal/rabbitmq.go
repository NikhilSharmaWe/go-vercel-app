package internal

import (
	"context"
	"fmt"

	"github.com/NikhilSharmaWe/go-vercel-app/vercel/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitClient struct {
	conn *amqp.Connection

	ch *amqp.Channel
}

func ConnectRabbitMQ(username, password, host, vhost string) (*amqp.Connection, error) {
	return amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/%s", username, password, host, vhost))
}

// we should recreate a channel of each concurrent tasks but reuse a connection
// 1 connection for a service and spawn channels for each concurrent thingy
// each client should use a personal channel for concurrent events
func NewRabbitMQClient(conn *amqp.Connection) (*RabbitClient, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := ch.Confirm(false); err != nil {
		return nil, err
	}

	return &RabbitClient{
		conn: conn,
		ch:   ch,
	}, nil
}

func (rc RabbitClient) Close() error {
	return rc.ch.Close()
}

func (rc RabbitClient) createQueue(queueName string, durable bool, autoDelete bool) (*amqp.Queue, error) {
	q, err := rc.ch.QueueDeclare(queueName, durable, autoDelete, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &q, nil
}

func CreateNewQueueReturnClient(conn *amqp.Connection, queueName string, durable bool, autoDelete bool) (*RabbitClient, error) {
	client, err := NewRabbitMQClient(conn)
	if err != nil {
		return nil, err
	}

	if _, err = client.createQueue(queueName, durable, autoDelete); err != nil {
		return nil, err
	}

	return client, nil
}

func (rc RabbitClient) CreateBinding(name, binding, exchange string) error {
	return rc.ch.QueueBind(name, binding, exchange, false, nil)
}

func (rc RabbitClient) Send(ctx context.Context, exchange, routingKey string, options amqp.Publishing) error {
	return rc.ch.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		true,
		false,
		options,
	)
}

func (rc RabbitClient) SendWithConfirmingPublish(ctx context.Context, exchange, routingKey string, options amqp.Publishing) error {
	confirmation, err := rc.ch.PublishWithDeferredConfirmWithContext(
		ctx,
		exchange,
		routingKey,
		true,
		false,
		options,
	)
	if err != nil {
		return err
	}

	done := make(chan bool)
	go func() {
		res := confirmation.Wait()
		done <- res
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return models.ErrConfirmationTimeout
	}
}

func (rc RabbitClient) Consume(queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	return rc.ch.Consume(queue, consumer, autoAck, false, false, false, nil)
}

func (rc RabbitClient) ApplyQualtyOfService(count, size int, global bool) error {
	return rc.ch.Qos(count, size, global)
}
