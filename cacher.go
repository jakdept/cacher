package cacher

import (
	"sync"
	"time"
)

type cacher struct {
	data  interface{}
	err   error
	isSet bool

	delay time.Duration
	timer *time.Timer

	lock sync.Mutex

	populate func() (interface{}, error)
}

func New(delay time.Duration, populate func() (interface{}, error)) cacher {
	return cacher{
		delay:    delay,
		lock:     sync.Mutex{},
		populate: populate,
	}
}

func (c *cacher) Clear() {
	// maybe stop some extra calls
	c.timer.Stop()

	// make sure nobody is clearing
	c.lock.Lock()
	defer c.lock.Unlock()

	// clear all values
	c.data, c.err, c.isSet = nil, nil, false
}

func (c *cacher) Get() (interface{}, error) {
	// make sure nobody is clearing
	c.lock.Lock()
	defer c.lock.Unlock()

	// if it's already populated, return cached results
	if c.isSet {
		return c.data, c.err
	}

	// if there's no delay set, just run it once and return the results
	if c.delay == 0 {
		return c.populate()
	}

	// if the values aren't populated, populate them
	c.data, c.err = c.populate()
	c.isSet = true

	// start a new timer to stop things after the delay
	c.timer = time.AfterFunc(c.delay, c.Clear)

	return c.data, c.err
}

func (c *cacher) ChangeDelay(d time.Duration) {
	c.delay = d
	c.Clear()
}
