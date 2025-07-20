package cache

import (
	"sync/atomic"
)

type CostRoutingThresholdCache interface {
	Set(value float64)
	Get() float64
}

type CostRoutingThresholdCacheImp struct {
	counter atomic.Int64
}

func NewCostRoutingThresholdCache() CostRoutingThresholdCache {
	return &CostRoutingThresholdCacheImp{}
}

func (c *CostRoutingThresholdCacheImp) Set(value float64) {
	c.counter.Store(int64(value * 100))
}

func (c *CostRoutingThresholdCacheImp) Get() float64 {
	return float64(c.counter.Load()) / 100
}
