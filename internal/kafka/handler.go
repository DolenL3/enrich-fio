package kafka

import (
	"context"
	"encoding/json"
	"enrich-fio/internal/models"
	"fmt"

	"github.com/pkg/errors"
	kafkago "github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

type KafkaHandler struct {
	Reader *Reader
	Writer *Writer
}

func NewHandler() *KafkaHandler {
	return &KafkaHandler{
		Reader: NewKafkaReader("FIO"),
		Writer: NewKafkaWriter("FIO_FAILED"),
	}
}

// func (h *KafkaHandler) HandlePerson(msg kafkago.Message) error {
// 	h.ValidateMessage(msg)
// 	return nil
// }

func (h *KafkaHandler) Start(ctx context.Context) error {
	messages := make(chan kafkago.Message, 100)
	invalidMessages := make(chan kafkago.Message, 100)
	messageCommitChan := make(chan kafkago.Message, 100)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return h.Reader.FetchMessage(ctx, invalidMessages, messages)
	})

	g.Go(func() error {
		return h.Writer.WriteMessages(ctx, invalidMessages, messageCommitChan)
	})

	g.Go(func() error {
		return h.Reader.CommitMessages(ctx, messageCommitChan)
	})

	err := g.Wait()
	if err != nil {
		return errors.Wrap(err, "kafka handler")
	}
	return nil
}

func validMessage(msg *kafkago.Message) bool {
	var person models.Person
	err := json.Unmarshal(msg.Value, &person)
	if err != nil {
		msg.WriterData = fmt.Sprintf("Invalid request: %v\nInvalid format\nError: %v", string(msg.Value), err.Error())
		return false

	}
	if person.Name == "" || person.Surname == "" {
		msg.WriterData = fmt.Sprintf("Invalid request: %v\nName and surname are required", string(msg.Value))
		return false
	}
	return true
}

// func (h *KafkaHandler) AddUser(msg kafkago.Message) error
