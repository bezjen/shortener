// Package pool provides sync.Pool wrapper with Resetter interface implementation
package pool

import (
	"sync"
)

// Resetter is an interface for types that can be reset to their initial state.
type Resetter interface {
	Reset()
}

// Pool is a generic container that stores and reuses objects of type T.
type Pool[T Resetter] struct {
	internal *sync.Pool
}

// New creates and returns a pointer to a new Pool instance for type T.
func New[T Resetter]() *Pool[T] {
	return &Pool[T]{
		internal: &sync.Pool{
			New: func() interface{} {
				var t T
				return t
			},
		},
	}
}

// Get returns an object from the pool.
func (p *Pool[T]) Get() T {
	obj := p.internal.Get()
	if obj == nil {
		var zero T
		return zero
	}
	return obj.(T)
}

// Put places an object back into the pool after resetting its state.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.internal.Put(obj)
}
