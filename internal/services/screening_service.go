package services

import (
	"context"
	"math/rand"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
)

type ScreeningService interface {
	Redirect(ctx context.Context, msg workers.Message) error
}

type ScreeningServiceImp struct {
	memoryCache       cache.AtomicCache
	highPriorityQueue workers.QueueWorker
	lowPriorityQueue  workers.QueueWorker
	waitingRoom       workers.QueueWorker
}

func NewScreeningService(memoryCache cache.AtomicCache, highPriorityQueue workers.QueueWorker, lowPriorityQueue workers.QueueWorker, waitingRoom workers.QueueWorker) ScreeningService {
	return &ScreeningServiceImp{
		memoryCache:       memoryCache,
		highPriorityQueue: highPriorityQueue,
		lowPriorityQueue:  lowPriorityQueue,
		waitingRoom:       waitingRoom,
	}
}

func (s *ScreeningServiceImp) Redirect(ctx context.Context, msg workers.Message) error {
	defaultStatusFail := s.memoryCache.GetHealthDeafultApi()
	fallbackStatusFail := s.memoryCache.GetHealthFallbackApi()

	if !defaultStatusFail {
		s.lowPriorityQueue.Send(msg)
		return nil
	}

	if !fallbackStatusFail {
		s.highPriorityQueue.Send(msg)
		return nil
	}

	if msg.ReprocessedHowManyTimes > 0 {
		baseChance := 30
		increment := 10
		chance := baseChance + increment*msg.ReprocessedHowManyTimes

		if chance > 80 {
			chance = 80
		}

		randValue := rand.Intn(100)

		if randValue < chance {
			s.calcRedirect(msg, 50)
			return nil
		}
	}

	s.waitingRoom.Send(msg)

	return nil
}

// func (s *ScreeningServiceImp) Redirect(ctx context.Context, msg workers.Message) error {
// 	defaultStatusFail := s.memoryCache.GetHealthDeafultApi()
// 	fallbackStatusFail := s.memoryCache.GetHealthFallbackApi()

// 	if !defaultStatusFail && !fallbackStatusFail {
// 		s.calcRedirect(msg, config.Env.CalcRedirect)
// 		return nil
// 	}

// 	if defaultStatusFail && !fallbackStatusFail {
// 		s.highPriorityQueue.Send(msg)
// 		return nil
// 	}

// 	if !defaultStatusFail && fallbackStatusFail {
// 		s.lowPriorityQueue.Send(msg)
// 		return nil
// 	}

// 	if msg.ReprocessedHowManyTimes > 0 {
// 		baseChance := 30
// 		increment := 10
// 		chance := baseChance + increment*msg.ReprocessedHowManyTimes

// 		if chance > 80 {
// 			chance = 80
// 		}

// 		randValue := rand.Intn(100)

// 		if randValue < chance {
// 			s.calcRedirect(msg, 50)
// 			return nil
// 		}
// 	}

// 	s.waitingRoom.Send(msg)

// 	return nil
// }

func (s *ScreeningServiceImp) calcRedirect(msg workers.Message, chance int) {
	randValue := rand.Intn(100)

	if randValue < chance {
		s.highPriorityQueue.Send(msg)
		return
	}

	s.lowPriorityQueue.Send(msg)
}
