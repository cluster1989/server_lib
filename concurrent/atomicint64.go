package concurrent

import "sync/atomic"
import "fmt"

type AtomicInt64 int64

func NewAtomicInt64(val int64) *AtomicInt64 {
	a := AtomicInt64(val)
	return &a
}

// 得到该值
func (a *AtomicInt64) Get() int64 {
	return int64(*a)
}

// 将值设置进去
func (a *AtomicInt64) Set(val int64) {
	atomic.StoreInt64((*int64)(a), val)
}

// 对比并且设置，原子操作
func (a *AtomicInt64) CompareAndSet(expect, update int64) bool {
	return atomic.CompareAndSwapInt64((*int64)(a), expect, update)
}

// 设置新值，并返回旧值
func (a *AtomicInt64) GetAndSet(val int64) int64 {
	for {
		current := a.Get()
		if a.CompareAndSet(current, val) {
			return current
		}
	}
}

func (a *AtomicInt64) GetAndIncrement() int64 {

	for {
		current := a.Get()
		next := current + 1
		if a.CompareAndSet(current, next) {
			return current
		}
	}
}

func (a *AtomicInt64) GetAndDecrement() int64 {
	for {
		current := a.Get()
		next := current - 1
		if a.CompareAndSet(current, next) {
			return current
		}
	}
}

func (a *AtomicInt64) GetAndAdd(val int64) int64 {
	for {
		current := a.Get()
		next := current + val
		if a.CompareAndSet(current, next) {
			return current
		}
	}
}

func (a *AtomicInt64) IncrementAndGet() int64 {
	for {
		current := a.Get()
		next := current + 1
		if a.CompareAndSet(current, next) {
			return next
		}
	}
}

func (a *AtomicInt64) DecrementAndGet() int64 {
	for {
		current := a.Get()
		next := current - 1
		if a.CompareAndSet(current, next) {
			return next
		}
	}
}

func (a *AtomicInt64) AddAndGet(val int64) int64 {
	for {
		current := a.Get()
		next := current + val
		if a.CompareAndSet(current, next) {
			return next
		}
	}
}

func (a *AtomicInt64) String() string {
	return fmt.Sprintf("%d", a.Get())
}
