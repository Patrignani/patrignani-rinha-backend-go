package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/clients"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/config"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/models"
)

type CheckHealthPaymentService interface {
	SetStatusPayment(ctx context.Context) error
}

type CheckHealthPaymentServiceImp struct {
	httpRequest *http.Client
	memoryCache cache.AtomicCache
}

func NewCheckHealthPaymentService(httpRequest *http.Client, memoryCache cache.AtomicCache) CheckHealthPaymentService {
	return &CheckHealthPaymentServiceImp{
		httpRequest: httpRequest,
		memoryCache: memoryCache,
	}
}

func (c *CheckHealthPaymentServiceImp) SetStatusPayment(ctx context.Context) error {
	var wg sync.WaitGroup
	var healthDefault *models.Health
	var healthfallback *models.Health
	var errDefault error
	var errFallback error
	var fallbackFail, defaultFail bool

	wg.Add(2)
	go func() {
		defer wg.Done()
		url := fmt.Sprintf("%s/payments/service-health", config.Env.DefaultUrl)
		healthDefault, errDefault = c.getStatus(ctx, url)
		if errDefault != nil {
			log.Printf("Error get health DEFAULT: %v", errFallback)
			return
		}

		defaultFail = healthDefault != nil && healthDefault.Failing || healthDefault.MinResponseTime > config.Env.LimitTimeHealth
	}()

	go func() {
		defer wg.Done()
		url := fmt.Sprintf("%s/payments/service-health", config.Env.FallbackUrl)
		healthfallback, errFallback = c.getStatus(ctx, url)
		if errFallback != nil {
			log.Printf("Error get health fallback: %v", errFallback)
			return
		}

		fallbackFail = healthfallback != nil && healthfallback.Failing || healthfallback.MinResponseTime > config.Env.LimitTimeHealth
	}()

	wg.Wait()

	c.memoryCache.SetHealthDeafultApi(defaultFail)
	c.memoryCache.SetHealthFallbackApi(fallbackFail)

	return nil
}

func (c *CheckHealthPaymentServiceImp) getStatus(ctx context.Context, url string) (*models.Health, error) {
	var health models.Health

	_, err := clients.Do(c.httpRequest, clients.RequestParams{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
	}, &health)

	if err != nil {
		return nil, err
	}

	return &health, nil
}
