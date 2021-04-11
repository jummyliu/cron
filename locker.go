package cron

import (
	"runtime"
	"sync/atomic"
)

// SpinLock
type SpinLock uint32

// Lock SpinLock
func (sl *SpinLock) Lock() {
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		runtime.Gosched()
	}
}

// Unlock SpinLock
func (sl *SpinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
	runtime.Gosched()
}

// NewSpinLock new SpinLock
func NewSpinLock() *SpinLock {
	var lock SpinLock
	return &lock
}
