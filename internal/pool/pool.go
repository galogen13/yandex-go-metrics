package pool

import (
	"fmt"
	"sync"
)

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	pool    sync.Pool
	newFunc func() T
}

func New[T Resetter](newFunc func() T) (*Pool[T], error) {
	if newFunc == nil {
		return nil, fmt.Errorf("newFunc cannot be nil")
	}

	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
		newFunc: newFunc,
	}, nil
}

func (p *Pool[T]) Get() T {
	item := p.pool.Get()
	if item == nil {
		return p.newFunc()
	}

	return item.(T)
}

func (p *Pool[T]) Put(item T) {
	item.Reset()

	p.pool.Put(item)
}
