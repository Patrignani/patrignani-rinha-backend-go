package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/repositories"
	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/clients"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/config"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/models"
)

type PaymentService interface {
	ExecuteDefault(ctx context.Context, msg workers.Message) error
	ExecuteFallback(ctx context.Context, msg workers.Message) error
}

type PaymentServiceImp struct {
	httpRequest *http.Client
	waitingRoom workers.QueueWorker
	memoryCache cache.AtomicCache
	repo        repositories.PaymentRepository
}

func NewPaymentService(httpRequest *http.Client, waitingRoom workers.QueueWorker, memoryCache cache.AtomicCache, repo repositories.PaymentRepository) PaymentService {
	return &PaymentServiceImp{
		httpRequest: httpRequest,
		waitingRoom: waitingRoom,
		memoryCache: memoryCache,
		repo:        repo,
	}
}

func (p *PaymentServiceImp) ExecuteDefault(ctx context.Context, msg workers.Message) error {
	url := fmt.Sprintf("%s/payments", config.Env.DefaultUrl)

	statusCode, err := p.postPayment(ctx, msg, url)

	if err != nil && statusCode != 422 {
		log.Printf("ExecuteDefault - error %v \n", err)
		if !config.Env.EnableCheckHealthCheck {
			p.memoryCache.SetHealthDeafultApi(true)
		}

		p.waitingRoom.Send(msg)

		return err
	}

	if !config.Env.EnableCheckHealthCheck {
		p.memoryCache.SetHealthDeafultApi(false)
	}

	if statusCode != 422 {
		if err := p.repo.Insert(ctx, models.PaymentDb{
			CorrelationId: msg.CorrelationId,
			Amount:        msg.Amount,
			Fallback:      false,
			CreatedAt:     time.Now().UTC(),
		}); err != nil {
			log.Printf("ExecuteDefault - insert %v \n", err)
			return err
		}
	}

	log.Printf("ExecuteDefault - inseriu \n")
	return nil
}

func (p *PaymentServiceImp) ExecuteFallback(ctx context.Context, msg workers.Message) error {
	url := fmt.Sprintf("%s/payments", config.Env.FallbackUrl)

	statusCode, err := p.postPayment(ctx, msg, url)

	if err != nil && statusCode != 422 {
		log.Printf("ExecuteFallback - error %v \n", err)
		if !config.Env.EnableCheckHealthCheck {
			p.memoryCache.SetHealthFallbackApi(true)
		}

		p.waitingRoom.Send(msg)

		return err
	}

	if !config.Env.EnableCheckHealthCheck {
		p.memoryCache.SetHealthFallbackApi(false)
	}

	if statusCode != 422 {
		if err := p.repo.Insert(ctx, models.PaymentDb{
			CorrelationId: msg.CorrelationId,
			Amount:        msg.Amount,
			Fallback:      true,
			CreatedAt:     time.Now().UTC(),
		}); err != nil {
			log.Printf("ExecuteFallback - insert %v \n", err)
			return err
		}
	}

	log.Printf("ExecuteFallback - inseriu \n")

	return nil
}

func (p *PaymentServiceImp) postPayment(ctx context.Context, msg workers.Message, url string) (int, error) {

	reqBody := models.PaymentRequest{
		CorrelationId: msg.CorrelationId,
		Amount:        msg.Amount,
		RequestedAt:   time.Now().UTC(),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatal("Erro ao serializar o corpo:", err)
	}

	resp, err := clients.Do[any](p.httpRequest, clients.RequestParams{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: bodyBytes,
		Ctx:  ctx,
	}, nil)

	return resp.StatusCode, err
}
