package services

import (
	"context"
	"fmt"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
)

type ScreeningService interface {
	Redirect(ctx context.Context, msg workers.Message[any]) error
}

type ScreeningServiceImp struct {
	memoryCache       cache.CostRoutingThresholdCache
	highPriorityQueue workers.QueueWorker[any]
	lowPriorityQueue  workers.QueueWorker[any]
}

func NewScreeningService(memoryCache cache.CostRoutingThresholdCache, highPriorityQueue workers.QueueWorker[any], lowPriorityQueue workers.QueueWorker[any]) ScreeningService {
	return &ScreeningServiceImp{
		memoryCache:       memoryCache,
		highPriorityQueue: highPriorityQueue,
		lowPriorityQueue:  lowPriorityQueue,
	}
}

func (s *ScreeningServiceImp) Redirect(ctx context.Context, msg workers.Message[any]) error {

	threshold := s.memoryCache.Get()

	value, ok := msg.Object.(float64)
	if !ok {
		return fmt.Errorf("valor inválido, esperado float64, recebido: %T", msg.Value)
	}

	if value > threshold {
		s.lowPriorityQueue.Send(value)
		return nil
	}

	s.highPriorityQueue.Send(value)

	return nil
}
