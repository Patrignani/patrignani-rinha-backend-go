package workers

import (
	"context"
	"fmt"
	"sync"
)

type Message[T any] struct {
	Id     string
	Object T
}

type QueueWorker[T any] interface {
	Send(msg Message[T])
	RetryFallback()
	Consume(ctx context.Context, workers int, process func(context.Context, Message[T]) error)
	CountFallback() int
}

type QueueWorkerImp[T any] struct {
	channel  chan Message[T]
	fallback map[string]Message[T]
	mu       sync.Mutex
}

func NewQueueWorker[T any](buffer int) QueueWorker[T] {
	return &QueueWorkerImp[T]{
		channel:  make(chan Message[T], buffer),
		fallback: make(map[string]Message[T]),
	}
}

func (q *QueueWorkerImp[T]) Send(msg Message[T]) {
	select {
	case q.channel <- msg:
	default:
		q.mu.Lock()
		q.fallback[msg.Id] = msg
		q.mu.Unlock()
		fmt.Println("Fila cheia, mensagem salva no fallback:", msg.Id)
	}
}

func (q *QueueWorkerImp[T]) RetryFallback() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for id, msg := range q.fallback {
		select {
		case q.channel <- msg:
			delete(q.fallback, id)
			fmt.Println("Mensagem reprocessada do fallback:", id)
		default:
			return
		}
	}
}

func (q *QueueWorkerImp[T]) Consume(ctx context.Context, workers int, process func(context.Context, Message[T]) error) {
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

			go func(m Message[T]) {
				defer wg.Done()
				defer func() { <-sem }()
				if err := process(ctx, m); err != nil {
					fmt.Printf("Erro ao processar mensagem %s: %v\n", m.Id, err)
				}
			}(msg)
		}
	}
}

func (q *QueueWorkerImp[T]) CountFallback() int {
	return len(q.fallback)
}
