package services

import (
	"context"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
)

type ScreeningService interface {
	Redirect(ctx context.Context, msg workers.Message) error
}

type ScreeningServiceImp struct {
	memoryCache       cache.CostRoutingThresholdCache
	highPriorityQueue workers.QueueWorker
	lowPriorityQueue  workers.QueueWorker
}

func NewScreeningService(memoryCache cache.CostRoutingThresholdCache, highPriorityQueue workers.QueueWorker, lowPriorityQueue workers.QueueWorker) ScreeningService {
	return &ScreeningServiceImp{
		memoryCache:       memoryCache,
		highPriorityQueue: highPriorityQueue,
		lowPriorityQueue:  lowPriorityQueue,
	}
}

func (s *ScreeningServiceImp) Redirect(ctx context.Context, msg workers.Message) error {

	threshold := s.memoryCache.Get()

	if msg.Amount.Cmp(threshold) == 1 {
		s.lowPriorityQueue.Send(msg)
		return nil
	}

	s.highPriorityQueue.Send(msg)

	return nil
}
