package trylock

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestMutexLock(t *testing.T) {
	mu := New()

	mu.Lock()
	mu.Unlock()
	mu.Lock()
	mu.Unlock()

	mu.RLock()
	mu.RUnlock()
	mu.RLock()
	mu.RUnlock()

	mu.TryLock(0)
	mu.Unlock()
	mu.TryLock(5 * time.Second)
	mu.Unlock()

	mu.RTryLock(0)
	mu.RUnlock()
	mu.RTryLock(5 * time.Second)
	mu.RUnlock()
}

func TestMutexLockTryLock(t *testing.T) {
	mu := New()

	if ok := mu.TryLock(0); !ok {
		t.Errorf("cannot Lock !!!")
	}
	if ok := mu.TryLock(0); ok {
		t.Errorf("cannot Lock twice !!!")
	}

	mu.Unlock()
}

func TestMutexLockAfterUnlock(t *testing.T) {
	mu := New()
	mu.Lock()

	go func() {
		time.Sleep(50 * time.Millisecond)
		mu.Unlock()
	}()

	mu.Lock()
	mu.Unlock()
}

func TestMutexLockAfterRUnlock(t *testing.T) {
	mu := New()
	mu.RLock()

	go func() {
		time.Sleep(50 * time.Millisecond)
		mu.RUnlock()
	}()

	mu.Lock()
	mu.Unlock()
}

func TestMutexLockTryLockTimeout(t *testing.T) {
	mu := New()
	mu.Lock()

	if ok := mu.TryLock(10 * time.Millisecond); ok {
		t.Errorf("should not Lock in 10ms !!!")
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		mu.Unlock()
	}()
	if ok := mu.TryLock(200 * time.Millisecond); !ok {
		t.Errorf("cannot Lock after 200ms !!!")
	}

	mu.Unlock()
}

func TestMutexLockRTryLockTimeout(t *testing.T) {
	mu := New()
	mu.Lock()

	if ok := mu.RTryLock(10 * time.Millisecond); ok {
		t.Errorf("should not Lock in 10ms !!!")
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		mu.Unlock()
	}()
	if ok := mu.RTryLock(200 * time.Millisecond); !ok {
		t.Errorf("cannot Lock after 200ms !!!")
	}
	mu.RUnlock()
}

func TestMutexLockUnLockTwice(t *testing.T) {
	mu := New()
	mu.Lock()
	defer func() {
		if x := recover(); x != nil {
			if x != "Unlock() failed" {
				t.Errorf("unexpect panic")
			}
		} else {
			t.Errorf("should panic after unlock twice")
		}
	}()
	mu.Unlock()
	mu.Unlock()
}

func TestMutexLockRLockTwice(t *testing.T) {
	mu := New()
	mu.RLock()
	mu.RLock()
	mu.RUnlock()
	mu.RUnlock()
}

func TestMutexLockUnLockInvalid(t *testing.T) {
	mu := New()
	mu.Lock()
	defer func() {
		if x := recover(); x != nil {
			if x != "RUnlock() failed" {
				t.Errorf("unexpect panic")
			}
		} else {
			t.Errorf("should panic after RUnlock a write lock")
		}
	}()
	mu.RUnlock()
}

func TestMutexLockBroadcast(t *testing.T) {
	mu := New()
	mu.Lock()

	done := int32(0)
	for i := 0; i < 3; i++ {
		go func() {
			mu.RLock()
			atomic.AddInt32(&done, 1)
			mu.RUnlock()
		}()
	}

	time.Sleep(10 * time.Millisecond)

	mu.Unlock()

	time.Sleep(10 * time.Millisecond)

	if done != 3 {
		t.Fatal("Broadcast is failed")
	}
}
