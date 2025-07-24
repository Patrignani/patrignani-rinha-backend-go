package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/cache"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/storage"
	"github.com/shopspring/decimal"
)

var (
	buildCostRoutingThreshold = func() string {
		columns := []string{"AVG(amount) as avg", "STDDEV_SAMP(amount) as stddev", "COUNT(*) as count"}
		return storage.BuildSelect("entry_history", columns, map[string]string{})
	}
)

type CostRoutingThresholdService interface {
	Calculation(ctx context.Context) error
}

type CostRoutingThresholdServiceImp struct {
	memoryCache cache.CostRoutingThresholdCache
	pg          storage.PostgresClient
	kFactor     decimal.Decimal
}

func NewCostRoutingThresholdService(memoryCache cache.CostRoutingThresholdCache, pg storage.PostgresClient, kFactor float32) CostRoutingThresholdService {
	return &CostRoutingThresholdServiceImp{
		memoryCache: memoryCache,
		kFactor:     decimal.NewFromFloat32(kFactor),
		pg:          pg,
	}
}

func (c *CostRoutingThresholdServiceImp) Calculation(ctx context.Context) error {
	var avg, stddev float64
	var count int64
	println(buildCostRoutingThreshold())
	err := c.pg.QueryRow(ctx, buildCostRoutingThreshold()).Scan(&avg, &stddev, &count)
	if err != nil {
		return fmt.Errorf("erro ao executar query no Postgres: %w", err)
	}

	if count < 10 {
		log.Printf("Registros insuficientes (%d < 10): pulando cálculo de threshold", count)
		return nil
	}

	avgDecimal := decimal.NewFromFloat(avg)
	stddevDecimal := decimal.NewFromFloat(stddev)

	threshold := avgDecimal.Sub(c.kFactor.Mul(stddevDecimal))

	c.memoryCache.Set(threshold)

	log.Printf("Threshold atualizado: registros=%d μ=%s σ=%s → threshold=%s", count, avgDecimal.String(), stddevDecimal.String(), threshold.String())
	return nil
}
