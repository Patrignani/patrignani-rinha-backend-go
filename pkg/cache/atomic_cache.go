package cache

import (
	"sync/atomic"

	"github.com/shopspring/decimal"
)

var zero = decimal.Zero

type AtomicCache interface {
	Set(value decimal.Decimal)
	Get() decimal.Decimal
	SetHealthAPIs(defaultAPIOn, fallbackAPIOn bool)
	GetHealthAPIs() (defaultAPIOn, fallbackAPIOn bool)
}

type AtomicCacheImp struct {
	counter       atomic.Int64
	defaultAPIOn  atomic.Bool
	fallbackAPIOn atomic.Bool
}

func NewCostRoutingThresholdCache() AtomicCache {
	return &AtomicCacheImp{}
}

func (c *AtomicCacheImp) Set(value decimal.Decimal) {
	if value.Cmp(zero) == 1 {
		intVal := value.Mul(decimal.NewFromInt(100)).IntPart()
		c.counter.Store(intVal)
	}
}

func (c *AtomicCacheImp) Get() decimal.Decimal {
	intVal := c.counter.Load()
	if intVal > 0 {
		return decimal.NewFromInt(intVal).Div(decimal.NewFromInt(100))
	}

	return zero
}

func (c *AtomicCacheImp) SetHealthAPIs(defaultAPIOn, fallbackAPIOn bool) {
	c.defaultAPIOn.Store(defaultAPIOn)
	c.fallbackAPIOn.Store(fallbackAPIOn)
}

func (c *AtomicCacheImp) GetHealthAPIs() (defaultAPIOn, fallbackAPIOn bool) {
	return c.defaultAPIOn.Load(), c.fallbackAPIOn.Load()
}
