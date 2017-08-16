package concurrent

import "sync/atomic"
import "fmt"

type AtomicBoolean int32

func NewAtomicBoolean(val bool) *AtomicBoolean {
	var a AtomicBoolean
	if val {
		a = AtomicBoolean(1)
	} else {
		a = AtomicBoolean(0)
	}
	return &a
}

// 得到该值
func (a *AtomicBoolean) Get() bool {
	return atomic.LoadInt32((*int32)(a)) != 0
}

// 将值设置进去
func (a *AtomicBoolean) Set(val bool) {
	if val {
		atomic.StoreInt32((*int32)(a), 1)
	} else {
		atomic.StoreInt32((*int32)(a), 0)
	}
}

// 对比并且设置，原子操作
func (a *AtomicBoolean) CompareAndSet(expect, update bool) bool {
	var (
		o int32
		n int32
	)

	if expect {
		o = 1
	} else {
		o = 0
	}

	if update {
		n = 1
	} else {
		o = 0
	}

	return atomic.CompareAndSwapInt32((*int32)(a), o, n)
}

// 设置新值，并返回旧值
func (a *AtomicBoolean) GetAndSet(val bool) bool {
	for {
		current := a.Get()
		if a.CompareAndSet(current, val) {
			return current
		}
	}
}

func (a *AtomicBoolean) String() string {
	return fmt.Sprintf("%t", a.Get())
}
