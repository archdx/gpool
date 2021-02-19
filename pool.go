package gpool

import (
	"sync"
	"time"
)

type G struct {
	wg *sync.WaitGroup

	gCh  chan *G
	fnCh chan func()

	releaseCh   chan struct{}
	releaseHook func()

	stopCh chan struct{}
}

func (g *G) Exec(fn func()) { g.fnCh <- fn }

func (g *G) Release() { g.releaseCh <- struct{}{} }

func (g *G) run() {
	for {
		select {
		case fn := <-g.fnCh:
			fn()
		case <-g.releaseCh:
			g.releaseHook()
			return
		}
	}
}

func (g *G) park() {
	defer g.wg.Done()

	for {
		select {
		case g.gCh <- g:
			g.run()
		case <-g.stopCh:
			return
		}
	}
}

type Pool struct {
	wg sync.WaitGroup

	gCh chan *G

	mu           sync.RWMutex
	active       int
	idle         int
	waitDuration time.Duration

	stopCh chan struct{}
}

func (p *Pool) G() *G {
	ts := time.Now()

	g := <-p.gCh
	wait := time.Now().Sub(ts)

	p.mu.Lock()
	p.active++
	p.idle--
	p.waitDuration += wait
	p.mu.Unlock()

	return g
}

type Stats struct {
	Active       int
	Idle         int
	WaitDuration time.Duration
}

func (p *Pool) Stats() *Stats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &Stats{
		Active:       p.active,
		Idle:         p.idle,
		WaitDuration: p.waitDuration,
	}
}

func (p *Pool) Close() {
	close(p.stopCh)
	p.wg.Wait()
}

func (p *Pool) onReleaseG() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.active--
	p.idle++
}

func NewPool(size int) *Pool {
	p := Pool{
		gCh:    make(chan *G),
		stopCh: make(chan struct{}),
	}

	p.wg.Add(size)
	p.idle = size

	for i := 0; i < size; i++ {
		go (&G{
			wg:          &p.wg,
			gCh:         p.gCh,
			fnCh:        make(chan func()),
			releaseCh:   make(chan struct{}, 1),
			releaseHook: p.onReleaseG,
			stopCh:      p.stopCh,
		}).park()
	}

	return &p
}
