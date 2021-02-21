package gpool

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPoolG(t *testing.T) {
	pool := NewPool(10)

	g := pool.G()
	defer g.Release()

	var wg sync.WaitGroup
	var called bool

	wg.Add(1)
	g.Exec(func() { called = true; wg.Done() })

	wg.Wait()

	assert.True(t, called)
}

func TestPoolStats(t *testing.T) {
	pool := NewPool(10)

	g := pool.G()
	defer g.Release()

	stats := pool.Stats()

	assert.Equal(t, 1, stats.Active)
	assert.Equal(t, 9, stats.Idle)
}
