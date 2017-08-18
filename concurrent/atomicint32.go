package concurrent

import "sync/atomic"
import "fmt"

type AtomicInt32 int32

func NewAtomicInt32(val int32) *AtomicInt32 {
	a := AtomicInt32(val)
	return &a
}

// 得到该值
func (a *AtomicInt32) Get() int32 {
	return int32(*a)
}

// 将值设置进去
func (a *AtomicInt32) Set(val int32) {
	atomic.StoreInt32((*int32)(a), val)
}

// 对比并且设置，原子操作
func (a *AtomicInt32) CompareAndSet(expect, update int32) bool {
	return atomic.CompareAndSwapInt32((*int32)(a), expect, update)
}

// 设置新值，并返回旧值
func (a *AtomicInt32) GetAndSet(val int32) int32 {
	for {
		current := a.Get()
		if a.CompareAndSet(current, val) {
			return current
		}
	}
}

func (a *AtomicInt32) GetAndIncrement() int32 {

	for {
		current := a.Get()
		next := current + 1
		if a.CompareAndSet(current, next) {
			return current
		}
	}
}

func (a *AtomicInt32) GetAndDecrement() int32 {
	for {
		current := a.Get()
		next := current - 1
		if a.CompareAndSet(current, next) {
			return current
		}
	}
}

func (a *AtomicInt32) GetAndAdd(val int32) int32 {
	for {
		current := a.Get()
		next := current + val
		if a.CompareAndSet(current, next) {
			return current
		}
	}
}

func (a *AtomicInt32) IncrementAndGet() int32 {
	for {
		current := a.Get()
		next := current + 1
		if a.CompareAndSet(current, next) {
			return next
		}
	}
}

func (a *AtomicInt32) DecrementAndGet() int32 {
	for {
		current := a.Get()
		next := current - 1
		if a.CompareAndSet(current, next) {
			return next
		}
	}
}

func (a *AtomicInt32) AddAndGet(val int32) int32 {
	for {
		current := a.Get()
		next := current + val
		if a.CompareAndSet(current, next) {
			return next
		}
	}
}

func (a *AtomicInt32) String() string {
	return fmt.Sprintf("%d", a.Get())
}
