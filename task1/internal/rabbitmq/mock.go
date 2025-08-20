package queue

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/service"
)

type MemoryQueue struct {
	messages    chan []byte
	mu          sync.Mutex
	subscribers []func([]byte) error
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		messages: make(chan []byte, 1000), // Буферизованный канал
	}
}

func (mq *MemoryQueue) PublishWithDelay(ctx context.Context, queueName string, message interface{}, delay time.Duration) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	go func() {
		time.Sleep(delay)
		select {
		case mq.messages <- data:
		case <-ctx.Done():
			log.Printf("Context cancelled while publishing message")
		}
	}()

	return nil
}

func (mq *MemoryQueue) Consume(ctx context.Context, queueName string) (<-chan []byte, error) {
	output := make(chan []byte)

	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-mq.messages:
				output <- msg
			}
		}
	}()

	return output, nil
}

func (mq *MemoryQueue) Subscribe(handler func([]byte) error) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.subscribers = append(mq.subscribers, handler)
}

// Implement Queue interface
var _ service.Queue = (*MemoryQueue)(nil)
