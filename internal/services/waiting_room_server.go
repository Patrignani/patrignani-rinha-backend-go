package services

import (
	"context"
	"log"
	"time"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/config"
)

type WaitingRoomServer interface {
	Delay(ctx context.Context, msg workers.Message) error
}

type WaitingRoomServerImp struct {
	screeningQueue workers.QueueWorker
}

func NewWaitingRoomServer(screeningQueue workers.QueueWorker) WaitingRoomServer {
	return &WaitingRoomServerImp{
		screeningQueue: screeningQueue,
	}
}

func (w *WaitingRoomServerImp) Delay(ctx context.Context, msg workers.Message) error {
	time.Sleep(config.Env.WaitingRoomSleepTime)
	msg.ReprocessedHowManyTimes++
	log.Printf("WaitingRoom msg: %s, ReprocessedHowManyTimes: %d", msg.CorrelationId, msg.ReprocessedHowManyTimes)
	w.screeningQueue.Send(msg)
	return nil
}
