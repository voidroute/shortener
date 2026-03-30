package cache

import (
	"sync"
	"time"
)

type Cleaner struct {
	interval   time.Duration
	stop       chan struct{}
	stopped    bool
	deleteFunc func()
	mu         sync.Mutex
}

func NewCleaner(interval time.Duration, deleteFunc func()) *Cleaner {
	return &Cleaner{
		interval:   interval,
		stop:       make(chan struct{}),
		deleteFunc: deleteFunc,
	}
}

func (c *Cleaner) Start() {
	go c.run()
}

func (c *Cleaner) run() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.deleteFunc()
		case <-c.stop:
			return
		}
	}
}

func (c *Cleaner) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.stopped {
		close(c.stop)
		c.stopped = true
	}
}
