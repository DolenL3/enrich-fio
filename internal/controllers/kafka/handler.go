package kafka

import (
	"context"
	"encoding/json"
	"enrich-fio/internal/config"
	enrichfio "enrich-fio/internal/enrich-fio"
	"fmt"

	"github.com/pkg/errors"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	_topicInput  = "FIO"
	_topicFailed = "FIO_FAILED"
)

// kafkaHandler is a kafka handler.
type KafkaHandler struct {
	reader  *Reader
	writer  *Writer
	service *enrichfio.Service
}

// NewHandler returns kafkaHandler.
func NewHandler(service *enrichfio.Service, config *config.KafkaConfig) *KafkaHandler {
	logger := zap.L()
	logger.Info(fmt.Sprintf("kafka is up and running on %s", config.Host))
	return &KafkaHandler{
		reader:  NewKafkaReader(_topicInput, config.Host),
		writer:  NewKafkaWriter(_topicFailed, config.Host),
		service: service,
	}
}

// request is expected request from kafka.
type request struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

// Start starts kafka handler.
func (h *KafkaHandler) Start(ctx context.Context) error {
	messages := make(chan kafkago.Message, 100)
	invalidMessages := make(chan kafkago.Message, 100)
	messageCommitChan := make(chan kafkago.Message, 100)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return h.reader.FetchMessage(ctx, invalidMessages, messages)
	})

	g.Go(func() error {
		return h.writer.WriteMessages(ctx, invalidMessages, messageCommitChan)
	})

	g.Go(func() error {
		return h.reader.CommitMessages(ctx, messageCommitChan)
	})

	g.Go(func() error {
		return h.fetchValidMessage(ctx, messages)
	})

	err := g.Wait()
	if err != nil {
		return errors.Wrap(err, "kafka handler")
	}
	return nil
}

// validMessage checks if message is valid.
func validMessage(msg *kafkago.Message) bool {
	person := request{}
	err := json.Unmarshal(msg.Value, &person)
	if err != nil {
		msg.WriterData = fmt.Sprintf("Invalid request: %v\nInvalid format\nError: %v", string(msg.Value), err.Error())
		return false

	}
	if person.Name == "" {
		msg.WriterData = fmt.Sprintf("Invalid request: %v\nName required", string(msg.Value))
		return false
	}
	if person.Surname == "" {
		msg.WriterData = fmt.Sprintf("Invalid request: %v\nSurname required", string(msg.Value))
	}
	return true
}

// fetchValidMessage fetches valid message.
func (h *KafkaHandler) fetchValidMessage(ctx context.Context, messageChan <-chan kafkago.Message) error {
	for {
		select {
		case message := <-messageChan:
			err := h.AddPerson(ctx, message)
			if err != nil {
				return errors.Wrap(err, "fetch valid message")
			}
		}
	}
}

// AddPerson sends person to buisness logic of service.
func (h *KafkaHandler) AddPerson(ctx context.Context, msg kafkago.Message) error {
	person := request{}
	err := json.Unmarshal(msg.Value, &person)
	if err != nil {
		return errors.Wrap(err, "unmarshal request")
	}
	err = h.service.AddPerson(ctx, person.Name, person.Surname, person.Patronymic)
	if err != nil {
		return errors.Wrap(err, "add person")
	}
	return nil
}
