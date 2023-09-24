package enrichfio

import (
	"context"
	"enrich-fio/internal/models"
	"sync"

	"github.com/segmentio/kafka-go"
)

func Start() error {
	// kafka.Start()   ||
	// rest.Start()    || msgBrokers.Start()
	// graphQL.Start() ||

	// msgBrokers.Start() should return
	// TODO add logger to ctx
	ctx := context.Background()
	messages := make(chan models.Message, 1000)
	messagesCommitChan := make(chan models.Message, 1000)

	wg := &sync.WaitGroup{}

	for {
		select {
		case msg := <-messages:
			wg.Add(1)
			go func() {
				defer wg.Done()
				ok, msgInvalid := validateMessage(ctx, msg)
				if ok {

				} else {
					messagesCommitChan <- msgInvalid
					// brokers.Kafka.Writer.WriteMessages(ctx, messages, messagesCommitChan)
				}
			}()
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		}
	}
}

func validateMessage(ctx context.Context, message models.Message) (ok bool, msgInvalid *kafka.Message) {
	return true, nil
}
