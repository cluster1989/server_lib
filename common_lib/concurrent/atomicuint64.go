package concurrent

import (
	"fmt"
	"sync/atomic"
)

type AtomicUint64 uint64

func NewAtomicUint64(val uint64) *AtomicUint64 {
	a := AtomicUint64(val)
	return &a
}

// 得到该值
func (a *AtomicUint64) Get() uint64 {
	return uint64(*a)
}

// 将值设置进去
func (a *AtomicUint64) Set(val uint64) {
	atomic.StoreUint64((*uint64)(a), val)
}

// 对比并且设置，原子操作
func (a *AtomicUint64) CompareAndSet(expect, update uint64) bool {
	return atomic.CompareAndSwapUint64((*uint64)(a), expect, update)
}

// 设置新值，并返回旧值
func (a *AtomicUint64) GetAndSet(val uint64) uint64 {
	for {
		current := a.Get()
		if a.CompareAndSet(current, val) {
			return current
		}
	}
}

func (a *AtomicUint64) GetAndIncrement() uint64 {

	for {
		current := a.Get()
		next := current + 1
		if a.CompareAndSet(current, next) {
			return current
		}
	}
}

func (a *AtomicUint64) GetAndDecrement() uint64 {
	for {
		current := a.Get()
		next := current - 1
		if a.CompareAndSet(current, next) {
			return current
		}
	}
}

func (a *AtomicUint64) GetAndAdd(val uint64) uint64 {
	for {
		current := a.Get()
		next := current + val
		if a.CompareAndSet(current, next) {
			return current
		}
	}
}

func (a *AtomicUint64) IncrementAndGet() uint64 {
	for {
		current := a.Get()
		next := current + 1
		if a.CompareAndSet(current, next) {
			return next
		}
	}
}

func (a *AtomicUint64) DecrementAndGet() uint64 {
	for {
		current := a.Get()
		next := current - 1
		if a.CompareAndSet(current, next) {
			return next
		}
	}
}

func (a *AtomicUint64) AddAndGet(val uint64) uint64 {
	for {
		current := a.Get()
		next := current + val
		if a.CompareAndSet(current, next) {
			return next
		}
	}
}

func (a *AtomicUint64) String() string {
	return fmt.Sprintf("%d", a.Get())
}
