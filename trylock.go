package trylock

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// TryLocker is a RWMutex with trylock support
type TryLocker interface {
	// TryLock acquires the write lock without blocking.
	// On success, returns true. On failure or context cancellation,
	// returns false.
	// A nil Context means try and return at once.
	TryLock(context.Context) bool

	// TryLockTimeout acquires the write lock without blocking.
	// On success, returns true. On failure or timeout, returns false.
	// A negative timeout means no timeout.
	// A zero timeout means try and return at once.
	TryLockTimeout(time.Duration) bool

	// Lock locks for writing.
	// If the lock is already locked for reading or writing, Lock blocks until the lock is available.
	Lock()

	// Unlock unlocks for writing.
	// It is a panic if is not locked for writing before.
	Unlock()

	// RTryLock acquires the read lock without blocking.
	// On success, returns true. On failure or timeout, returns false.
	// A nil Context means try and return at once.
	RTryLock(context.Context) bool

	// RTryLockTimeout acquires the read lock without blocking.
	// On success, returns true. On failure or timeout, returns false.
	// A negative timeout means no timeout.
	// A zero timeout means try and return at once.
	RTryLockTimeout(time.Duration) bool

	// RLock locks for reading.
	// If the lock is already locked for writing, RLock blocks until the lock is available.
	RLock()

	// RUnlock unlocks for reading.
	// It is a panic if is not locked for reading before.
	RUnlock()
}

// trylocker implements TryLocker interface
type trylocker struct {
	// lock state
	// if state == 0, no lock holds
	// if state == -1, write lock holds
	// if state > 0, read lock holds, and the value is the number of readers
	state *int32

	// a broadcast channel
	ch chan struct{}
	// a locker for acquires broadcast channel
	lock sync.Mutex
}

// confirm trylocker implements sync.Locker on compiling phase
var _ sync.Locker = &trylocker{}

// New create a new TryLocker instance
func New() TryLocker {
	return &trylocker{
		state: new(int32),
		ch:    make(chan struct{}, 1),
	}
}

func (m *trylocker) TryLock(ctx context.Context) bool {
	for {
		if atomic.CompareAndSwapInt32(m.state, 0, -1) {
			// acquire OK
			return true
		}
		if ctx == nil {
			return false
		}

		// get broadcast channel
		ch := m.channel()

		// waiting for broadcast signal or timeout
		select {
		case <-ch:
			// wake up to try again
		case <-ctx.Done():
			// timeout
			return false
		}
	}
}
func (m *trylocker) TryLockTimeout(d time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return m.TryLock(ctx)
}

func (m *trylocker) RTryLock(ctx context.Context) bool {
	for {
		n := atomic.LoadInt32(m.state)
		if n >= 0 {
			if atomic.CompareAndSwapInt32(m.state, n, n+1) {
				// acquire OK
				return true
			}
		}

		// get broadcast channel
		ch := m.channel()

		select {
		case <-ch:
			// wake up to try again
		case <-ctx.Done():
			// timeout
			return false
		}
	}
}
func (m *trylocker) RTryLockTimeout(d time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return m.RTryLock(ctx)
}

func (m *trylocker) Lock() {
	m.TryLock(context.Background())
}

func (m *trylocker) RLock() {
	m.RTryLock(context.Background())
}

func (m *trylocker) Unlock() {
	if ok := atomic.CompareAndSwapInt32(m.state, -1, 0); !ok {
		panic("Unlock() failed")
	}

	m.broadcast()
}

func (m *trylocker) RUnlock() {
	n := atomic.AddInt32(m.state, -1)
	if n < 0 {
		panic("RUnlock() failed")
	}

	if n == 0 {
		m.broadcast()
	}
}

// get broadcast channel
func (m *trylocker) channel() chan struct{} {
	m.lock.Lock()
	ch := m.ch
	m.lock.Unlock()

	return ch
}

// send broadcast signal
func (m *trylocker) broadcast() {
	newCh := make(chan struct{}, 1)

	m.lock.Lock()
	ch := m.ch
	m.ch = newCh
	m.lock.Unlock()

	// send broadcast signal
	close(ch)
}
