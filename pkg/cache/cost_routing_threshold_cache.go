package cache

import (
	"sync/atomic"

	"github.com/shopspring/decimal"
)

var zero = decimal.Zero

type CostRoutingThresholdCache interface {
	Set(value decimal.Decimal)
	Get() decimal.Decimal
}

type CostRoutingThresholdCacheImp struct {
	counter atomic.Int64
}

func NewCostRoutingThresholdCache() CostRoutingThresholdCache {
	return &CostRoutingThresholdCacheImp{}
}

func (c *CostRoutingThresholdCacheImp) Set(value decimal.Decimal) {
	if value.Cmp(zero) == 1 {
		intVal := value.Mul(decimal.NewFromInt(100)).IntPart()
		c.counter.Store(intVal)
	}
}

func (c *CostRoutingThresholdCacheImp) Get() decimal.Decimal {
	intVal := c.counter.Load()
	if intVal > 0 {
		return decimal.NewFromInt(intVal).Div(decimal.NewFromInt(100))
	}

	return zero
}
