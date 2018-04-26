package trylock

import (
	"sync/atomic"
	"time"
)

const sleepInteval = 1 * time.Millisecond

// MutexLock is a simple sync.RWMutex + ability to try to Lock.
type MutexLock struct {
	// if v == 0, no lock
	// if v == -1, write lock
	// if v > 0, read lock, and v is the number of readers
	v *int32
}

// New returns a new MutexLock
func New() *MutexLock {
	v := int32(0)
	return &MutexLock{&v}
}

// TryLock tries to lock for writing. It returns true in case of success, false if timeout.
// A negative timeout means no timeout. If timeout is 0 that means try at once and quick return.
func (m *MutexLock) TryLock(timeout time.Duration) bool {
	start := time.Now()
	for {
		if atomic.CompareAndSwapInt32(m.v, 0, -1) {
			return true
		}
		if timeout >= 0 && time.Now().Sub(start) >= timeout {
			return false
		}
		time.Sleep(sleepInteval)
	}
}

// TryRLock tries to lock for reading. It returns true in case of success, false if timeout.
// A negative timeout means no timeout. If timeout is 0 that means try at once and quick return.
func (m *MutexLock) TryRLock(timeout time.Duration) bool {
	start := time.Now()
	for {
		n := atomic.LoadInt32(m.v)
		if n >= 0 {
			if atomic.CompareAndSwapInt32(m.v, n, n+1) {
				return true
			}
		}
		if timeout >= 0 && time.Now().Sub(start) >= timeout {
			return false
		}
		time.Sleep(sleepInteval)
	}
}

// Lock locks for writing. If the lock is already locked for reading or writing, Lock blocks until the lock is available.
func (m *MutexLock) Lock() {
	m.TryLock(-1)
}

// RLock locks for reading. If the lock is already locked for writing, RLock blocks until the lock is available.
func (m *MutexLock) RLock() {
	m.TryRLock(-1)
}

// Unlock unlocks for writing. It is a panic if m is not locked for writing on entry to Unlock.
func (m *MutexLock) Unlock() {
	if ok := atomic.CompareAndSwapInt32(m.v, -1, 0); !ok {
		panic("Unlock() failed")
	}
}

// RUnlock unlocks for reading. It is a panic if m is not locked for reading on entry to Unlock.
func (m *MutexLock) RUnlock() {
	if n := atomic.AddInt32(m.v, -1); n < 0 {
		panic("RUnlock() failed")
	}
}
