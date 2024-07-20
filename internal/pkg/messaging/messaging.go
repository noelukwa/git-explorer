package messaging

import (
	"fmt"

	"github.com/noelukwa/git-explorer/internal/events"
)

type Event struct {
	Subject events.EventKind
	Data    []byte
}

type Client interface {
	Publish(event Event) error
	// PublishWithRetry(subject string, data []byte, maxRetries int, backoffInitial, backoffMax time.Duration) error
	Subscribe(subject events.EventKind, handler func([]byte) error) error
	// SubscribeWithRetry(subject string, handler func([]byte) error, maxRetries int, backoffInitial, backoffMax time.Duration) error
	Close() error
}

func NewMessagingClient(provider, url string) (Client, error) {
	switch provider {
	case "nats":
		return NewNATSClient(url)
	default:
		return nil, fmt.Errorf("unsupported messaging provider: %s", provider)
	}
}
