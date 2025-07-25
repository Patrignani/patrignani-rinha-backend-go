package cache

import (
	"sync/atomic"

	"github.com/shopspring/decimal"
)

var zero = decimal.Zero

type AtomicCache interface {
	SetHealthDeafultApi(defaultAPIOn bool)
	GetHealthDeafultApi() bool
	SetHealthFallbackApi(fallbackAPIOn bool)
	GetHealthFallbackApi() (fallbackAPIOn bool)
}

type AtomicCacheImp struct {
	defaultAPIOn  atomic.Bool
	fallbackAPIOn atomic.Bool
}

func NewCostRoutingThresholdCache() AtomicCache {
	return &AtomicCacheImp{}
}

func (c *AtomicCacheImp) SetHealthDeafultApi(defaultAPIOn bool) {
	c.defaultAPIOn.Store(defaultAPIOn)
}

func (c *AtomicCacheImp) GetHealthDeafultApi() bool {
	return c.defaultAPIOn.Load()
}

func (c *AtomicCacheImp) SetHealthFallbackApi(fallbackAPIOn bool) {
	c.fallbackAPIOn.Store(fallbackAPIOn)
}

func (c *AtomicCacheImp) GetHealthFallbackApi() (fallbackAPIOn bool) {
	return c.fallbackAPIOn.Load()
}
