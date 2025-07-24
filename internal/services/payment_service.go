package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
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
}

func NewPaymentService(httpRequest *http.Client, waitingRoom workers.QueueWorker) PaymentService {
	return &PaymentServiceImp{
		httpRequest: httpRequest,
		waitingRoom: waitingRoom,
	}
}

func (p *PaymentServiceImp) ExecuteDefault(ctx context.Context, msg workers.Message) error {
	url := fmt.Sprintf("%s/payments", config.Env.DefaultUrl)

	err := p.postPayment(ctx, msg, url)

	if err != nil {
		p.waitingRoom.Send(msg)
	}

	return err
}

func (p *PaymentServiceImp) ExecuteFallback(ctx context.Context, msg workers.Message) error {
	url := fmt.Sprintf("%s/payments", config.Env.FallbackUrl)

	err := p.postPayment(ctx, msg, url)

	if err != nil {
		p.waitingRoom.Send(msg)
	}

	return err
}

func (p *PaymentServiceImp) postPayment(ctx context.Context, msg workers.Message, url string) error {

	reqBody := models.PaymentRequest{
		CorrelationId: msg.CorrelationId,
		Amount:        msg.Amount,
		RequestedAt:   time.Now().UTC(),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatal("Erro ao serializar o corpo:", err)
	}

	_, err = clients.Do[any](p.httpRequest, clients.RequestParams{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: bodyBytes,
		Ctx:  ctx,
	}, nil)

	return err
}
