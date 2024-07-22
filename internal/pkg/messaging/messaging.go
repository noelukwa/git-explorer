package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/noelukwa/git-explorer/internal/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *Client) Close() {
	c.channel.Close()
	c.conn.Close()
}

func (c *Client) DeclareQueue(name string) error {
	_, err := c.channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	return nil
}

func (c *Client) Publish(ctx context.Context, queueName string, event events.EventKind, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers: amqp.Table{
				"event_kind": string(event),
			},
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (c *Client) Subscribe(ctx context.Context, queueName string, handler func(context.Context, events.EventKind, []byte)) error {
	msgs, err := c.channel.ConsumeWithContext(ctx,
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			eventKind := events.EventKind(msg.Headers["event_kind"].(string))
			handler(ctx, eventKind, msg.Body)
		}
	}()

	return nil
}
