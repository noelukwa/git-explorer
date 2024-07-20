package messaging

import (
	"github.com/nats-io/nats.go"
	"github.com/noelukwa/git-explorer/internal/events"
)

type NATSClient struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

func NewNATSClient(url string) (*NATSClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	return &NATSClient{conn: nc, js: js}, nil
}

func (nc *NATSClient) Close() error {
	nc.conn.Close()
	return nil
}

func (nc *NATSClient) Conn() *nats.Conn {
	return nc.conn
}

func (nc *NATSClient) Publish(event Event) error {
	_, err := nc.js.Publish(string(event.Subject), event.Data)
	return err
}

func (nc *NATSClient) Subscribe(subject events.EventKind, handler func([]byte) error) error {
	_, err := nc.js.Subscribe(string(subject), func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			msg.Nak()
		} else {
			msg.Ack()
		}
	}, nats.ManualAck())
	return err
}
