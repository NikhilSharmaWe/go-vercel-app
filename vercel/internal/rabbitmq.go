package internal

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitClient struct {
	conn *amqp.Connection

	ch *amqp.Channel
}

func connectRabbitMQ(username, password, host, vhost string) (*amqp.Connection, error) {
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

	// put the channel to confirm mode
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

func CreateNewQueueWithNewConnAndClient(username, password, host, vhost, queueName string, durable bool, autoDelete bool) (*amqp.Queue, error) {
	conn, err := connectRabbitMQ(username, password, host, vhost)
	if err != nil {
		return nil, err
	}

	client, err := NewRabbitMQClient(conn)
	if err != nil {
		return nil, err
	}

	return client.createQueue(queueName, durable, autoDelete)
}

// CreateBinding will bind the current channel to the given exchange using the routing key provided
func (rc RabbitClient) CreateBinding(name, binding, exchange string) error {
	// we have create the customer_events exchange through the cli
	// having nowait to false will make the channel return error if it is fail to bind
	return rc.ch.QueueBind(name, binding, exchange, false, nil)
}

// Send is used to publish payloads onto an exchange with the given routing key
func (rc RabbitClient) Send(ctx context.Context, exchange, routingKey string, options amqp.Publishing) error {
	return rc.ch.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		// Mandatory is used to determine if an error should be returned upon failure
		true,
		// Immediate
		false,
		options,
	)
}

// PublishWithDeferredConfirmWithContext is not acknowledgement, it is that server acknowledges that it has published the message on the exchange, they are different but it is good to know that the message is sent successfully
func (rc RabbitClient) SendWithConfirmingPublish(ctx context.Context, exchange, routingKey string, options amqp.Publishing) error {
	// this only works if
	confirmation, err := rc.ch.PublishWithDeferredConfirmWithContext(
		ctx,
		exchange,
		routingKey,
		// Mandatory is used to determine if an error should be returned upon failure
		true,
		// Immediate
		false,
		options,
	)

	if err != nil {
		return err
	}

	log.Println(confirmation.Wait())
	return nil
}

func (rc RabbitClient) Consume(queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	// if auto ack is set to true then the consumer will send the acknowledgement when the message is consumed but this is not a good practice for the services which can fail, because if we send the ack and the service fail that message will be lost. so we should send the ack manually after the service completes the task from the message consumed from the queue
	return rc.ch.Consume(queue, consumer, autoAck, false, false, false, nil)
}

// ApplyQOS
// prefetch count - an integer on how many unacknowledged messages the server can send
// prefetch size - is int of how many bytes
// global - detemines if the rule should be applied globally or not
func (rc RabbitClient) ApplyQualtyOfService(count, size int, global bool) error {
	return rc.ch.Qos(count, size, global)
}
