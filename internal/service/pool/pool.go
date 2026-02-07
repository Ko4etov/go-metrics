package pool

import (
	"sync"
)

type Resetable interface {
	Reset()
}

type Pool[T Resetable] struct {
	pool sync.Pool
}

func New[T Resetable](newFunc func() T) *Pool[T] {
	p := &Pool[T]{}
	p.pool.New = func() any {
		return newFunc()
	}
	return p
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.pool.Put(x)
}