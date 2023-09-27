package kafka

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"
)

// Writer is kafka writer implementation.
type Writer struct {
	Writer *kafkago.Writer
}

// NewKafkaWriter returns writer implementation.
func NewKafkaWriter(topic string, host string) *Writer {
	writer := &kafkago.Writer{
		Addr:  kafkago.TCP(host),
		Topic: topic,
	}
	return &Writer{
		Writer: writer,
	}
}

// WriteMessages writes messages to kafka.
func (k *Writer) WriteMessages(ctx context.Context, invalidMessages chan kafkago.Message, invalidMessageCommitChan chan kafkago.Message) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-invalidMessages:
			info := msg.WriterData.(string)
			err := k.Writer.WriteMessages(ctx, kafkago.Message{
				Value: append(msg.Value, []byte(info)...),
			})
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
			case invalidMessageCommitChan <- msg:
			}
		}
	}
}
