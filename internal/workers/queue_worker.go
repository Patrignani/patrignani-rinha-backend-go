package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Message struct {
	Id                      string
	CorrelationId           string
	Amount                  decimal.Decimal
	EnqueueAt               time.Time
	ReprocessedHowManyTimes int
}

type QueueWorker interface {
	Send(msg Message)
	RetryFallback()
	Consume(ctx context.Context, workers int, process func(context.Context, Message) error)
	CountFallback() int
}

type QueueWorkerImp struct {
	channel  chan Message
	fallback []Message
	mu       sync.Mutex
}

func NewQueueWorker(buffer int) QueueWorker {
	return &QueueWorkerImp{
		channel:  make(chan Message, buffer),
		fallback: []Message{},
	}
}

func (q *QueueWorkerImp) Send(msg Message) {
	select {
	case q.channel <- msg:
	default:
		q.mu.Lock()
		q.fallback = append(q.fallback, msg)
		q.mu.Unlock()
		fmt.Println("Fila cheia, mensagem salva no fallback:", msg.Id)
	}
}

func (q *QueueWorkerImp) RetryFallback() {
	q.mu.Lock()
	defer q.mu.Unlock()

	newFallback := q.fallback[:0]

	for _, msg := range q.fallback {
		select {
		case q.channel <- msg:
			fmt.Println("Mensagem reprocessada do fallback:", msg.Id)
		default:
			newFallback = append(newFallback, msg)
		}
	}

	q.fallback = newFallback
}

func (q *QueueWorkerImp) Consume(ctx context.Context, workers int, process func(context.Context, Message) error) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			close(q.channel)
			fmt.Println("Consumo encerrado")
			return
		case msg, ok := <-q.channel:
			if !ok {
				wg.Wait()
				fmt.Println("Canal fechado e todas mensagens processadas")
				return
			}

			sem <- struct{}{}
			wg.Add(1)

			go func(m Message) {
				defer wg.Done()
				defer func() { <-sem }()
				if err := process(ctx, m); err != nil {
					fmt.Printf("Erro ao processar mensagem %s: %v\n", m.Id, err)
				}
			}(msg)
		}
	}
}

func (q *QueueWorkerImp) CountFallback() int {
	return len(q.fallback)
}
