package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NATSClient struct {
	conn   *nats.Conn
	js     jetstream.JetStream
	stream jetstream.Stream
}

func NewNATSClient(ctx context.Context, url string) (*NATSClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "GITEXPLORER",
		Subjects:  []string{"intent.*"},
		Retention: jetstream.WorkQueuePolicy,
	})
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}
	return &NATSClient{conn: nc, js: js, stream: stream}, nil
}

func (nc *NATSClient) Close() error {
	nc.conn.Close()
	return nil
}

func (nc *NATSClient) Conn() *nats.Conn {
	return nc.conn
}

func (nc *NATSClient) Publish(ctx context.Context, event Event) error {
	key := fmt.Sprintf("intent.%s", event.Subject)
	_, err := nc.js.Publish(ctx, key, event.Data)
	return err
}

func (nc *NATSClient) NewConsumer(ctx context.Context, subs []string, name string) (jetstream.Consumer, error) {

	consumer, err := nc.stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:           name,
		FilterSubjects: subs,
	})
	return consumer, err
}

func (nc *NATSClient) NeKV(ctx context.Context, name string) (jetstream.KeyValue, error) {
	kv, err := nc.js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: name,
	})
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	return kv, nil
}
