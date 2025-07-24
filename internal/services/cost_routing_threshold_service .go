package services

import (
	"context"
	"database/sql"
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
	var avg, stddev sql.NullFloat64
	var count int64

	err := c.pg.QueryRow(ctx, buildCostRoutingThreshold()).Scan(&avg, &stddev, &count)
	if err != nil {
		return fmt.Errorf("erro ao executar query no Postgres: %w", err)
	}

	if count < 10 || !avg.Valid || !stddev.Valid {
		log.Printf("Registros insuficientes ou dados inválidos: count=%d, avg valid=%v, stddev valid=%v",
			count, avg.Valid, stddev.Valid)
		return nil
	}

	avgDecimal := decimal.NewFromFloat(avg.Float64)
	stddevDecimal := decimal.NewFromFloat(stddev.Float64)

	threshold := avgDecimal.Sub(c.kFactor.Mul(stddevDecimal))

	c.memoryCache.Set(threshold)

	log.Printf("Threshold atualizado: registros=%d μ=%s σ=%s → threshold=%s",
		count, avgDecimal.String(), stddevDecimal.String(), threshold.String())

	return nil
}
