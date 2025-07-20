package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/storage"
)

var (
	kFactor                   float64 = 0.4
	buildCostRoutingThreshold         = func() string {
		columns := []string{"AVG(value) as avg", "STDDEV_SAMP(value) as stddev", "COUNT(*) as count"}
		return storage.BuildSelect("entry_history", columns, map[string]string{})
	}
)

type CostRoutingThresholdService interface {
	Calculation(ctx context.Context) error
}

type CostRoutingThresholdServiceImp struct {
	memoryCache cache.CostRoutingThresholdCache
	pg          storage.PostgresClient
}

func NewCostRoutingThresholdService(memoryCache cache.CostRoutingThresholdCache, pg storage.PostgresClient) CostRoutingThresholdService {
	return &CostRoutingThresholdServiceImp{
		memoryCache: memoryCache,
		pg:          pg,
	}
}

func (c *CostRoutingThresholdServiceImp) Calculation(ctx context.Context) error {

	var avg, stddev float64
	var count int64

	err := c.pg.QueryRow(ctx, buildCostRoutingThreshold()).Scan(&avg, &stddev, &count)
	if err != nil {
		return fmt.Errorf("erro ao executar query no Postgres: %w", err)
	}

	if count < 10 {
		log.Printf("Registros insuficientes (%d < 10): pulando cálculo de threshold", count)
		return nil
	}

	threshold := avg - kFactor*stddev

	c.memoryCache.Set(threshold)

	log.Printf("Threshold atualizado: registros=%d μ=%.4f σ=%.4f → threshold=%.4f", count, avg, stddev, threshold)
	return nil
}
