package kafka

import (
	"context"
	"enrich-fio/internal/config"
	"log"

	"github.com/pkg/errors"
	kafkago "github.com/segmentio/kafka-go"
)

// Reader is kafka reader implementation.
type Reader struct {
	Reader *kafkago.Reader
}

// NewKafkaReader returns reader implementation.
func NewKafkaReader(topic string) *Reader {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{config.Conf.KafkaHost},
		Topic:   topic,
		GroupID: "group",
	})

	return &Reader{
		Reader: reader,
	}
}

// FetchMessage fetches message from kafka.
func (k *Reader) FetchMessage(ctx context.Context, invalidMessages chan<- kafkago.Message, messages chan<- kafkago.Message) error {
	for {
		message, err := k.Reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		if !validMessage(&message) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case invalidMessages <- message:
				log.Println(message.WriterData)
			}
		} else {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case messages <- message:
				log.Printf("message is valid: %v \n", string(message.Value))
			}
		}
	}
}

// CommitMessages commits messages from kafka, so they wouldn't be sent again.
func (k *Reader) CommitMessages(ctx context.Context, invalidMessageCommitChan <-chan kafkago.Message) error {
	for {
		select {
		case <-ctx.Done():
		case msg := <-invalidMessageCommitChan:
			err := k.Reader.CommitMessages(ctx, msg)
			if err != nil {
				return errors.Wrap(err, "Reader.CommitMessages")
			}
			log.Printf("committed msg: %v \n", string(msg.Value))
		}
	}
}
