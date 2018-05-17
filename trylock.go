package trylock

import (
	"sync/atomic"
	"time"
)

// MutexLock is a simple sync.RWMutex + ability to try to Lock.
type MutexLock struct {
	// if v == 0, no lock
	// if v == -1, write lock
	// if v > 0, read lock, and v is the number of readers
	v *int32

	// wakeup channel
	// it is only to wake up one of writelock waiters
	// it is not for readlock waiters because of they are need to be waked up together.
	ch chan struct{}
}

// New returns a new MutexLock
func New() *MutexLock {
	v := int32(0)
	ch := make(chan struct{}, 1)
	return &MutexLock{&v, ch}
}

// TryLock tries to lock for writing. It returns true in case of success, false if timeout.
// A negative timeout means no timeout. If timeout is 0 that means try at once and quick return.
// If the lock is currently held by another goroutine, TryLock will wait until it has a chance to acquire it.
func (m *MutexLock) TryLock(timeout time.Duration) bool {
	// deadline for timeout
	deadline := time.Now().Add(timeout)

	for {
		if atomic.CompareAndSwapInt32(m.v, 0, -1) {
			return true
		}

		// Waiting for wake up before trying again.
		if timeout < 0 {
			<-m.ch
		} else {
			elapsed := deadline.Sub(time.Now())
			if elapsed <= 0 {
				// timeout
				return false
			}

			select {
			case <-m.ch:
				// wake up to try again
			case <-time.After(elapsed):
				// timeout
				return false
			}
		}
	}
}

// TryRLock tries to lock for reading. It returns true in case of success, false if timeout.
// A negative timeout means no timeout. If timeout is 0 that means try at once and quick return.
func (m *MutexLock) TryRLock(timeout time.Duration) bool {
	start := time.Now()

	sleepInterval := 1 * time.Millisecond

	// compute max sleep interval (1..64 ms)
	maxSleepInterval := timeout / 25
	if maxSleepInterval < 0 {
		maxSleepInterval = 64 * time.Millisecond // no timeout
	} else if maxSleepInterval < 1*time.Millisecond {
		maxSleepInterval = 1 * time.Millisecond
	} else if maxSleepInterval > 64*time.Millisecond {
		maxSleepInterval = 64 * time.Millisecond
	}

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

		// progressive sleep interval
		if sleepInterval < maxSleepInterval {
			sleepInterval *= 2
		}
		time.Sleep(sleepInterval)
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

	select {
	case m.ch <- struct{}{}:
		// to wake up waiters
	default:
		// ch is full, skip
	}
}

// RUnlock unlocks for reading. It is a panic if m is not locked for reading on entry to Unlock.
func (m *MutexLock) RUnlock() {
	if n := atomic.AddInt32(m.v, -1); n < 0 {
		panic("RUnlock() failed")
	}

	select {
	case m.ch <- struct{}{}:
		// to wake up waiters
	default:
		// ch is full, skip
	}
}
