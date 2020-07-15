package trylock

import (
	"sync"
	"testing"
	"time"
)

func BenchmarkStdLock(b *testing.B) {
	mux := sync.Mutex{}
	b.Run("goroutine-1", func(_b *testing.B) {
		_b.ReportAllocs()
		for i := 0; i < _b.N; i++ {
			mux.Lock()
			mux.Unlock()
		}
	})

	b.Run("goroutine-N", func(_b *testing.B) {
		var countor = 0
		var wg sync.WaitGroup
		for i := 0; i < _b.N; i++ {
			_b.ReportAllocs()
			wg.Add(1)
			go func() {
				mux.Lock()
				countor++
				mux.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func BenchmarkLock(b *testing.B) {
	mux := New()
	b.Run("goroutine-1", func(_b *testing.B) {
		_b.ReportAllocs()
		for i := 0; i < _b.N; i++ {
			mux.Lock()
			mux.Unlock()
		}
	})

	b.Run("goroutine-N", func(_b *testing.B) {
		var countor = 0
		var wg sync.WaitGroup
		for i := 0; i < _b.N; i++ {
			_b.ReportAllocs()
			wg.Add(1)
			go func() {
				mux.Lock()
				countor++
				mux.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()
		if countor != _b.N {
			b.Fatalf("countor(%d) != _b.N(%d)", countor, _b.N)
		}
	})
}

func BenchmarkTryLock(b *testing.B) {
	mux := New()
	b.Run("goroutine-1", func(_b *testing.B) {
		_b.ReportAllocs()
		for i := 0; i < _b.N; i++ {
			if !mux.TryLock(nil) {
				_b.FailNow()
			}
			mux.Unlock()
		}
	})

	b.Run("goroutine-N", func(_b *testing.B) {
		var countor = 0
		var wg sync.WaitGroup
		for i := 0; i < _b.N; i++ {
			_b.ReportAllocs()
			wg.Add(1)
			go func() {
				if !mux.TryLock(nil) {
					_b.FailNow()
				}
				countor++
				mux.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()
		if countor != _b.N {
			b.Fatalf("countor(%d) != _b.N(%d)", countor, _b.N)
		}
	})
}

func BenchmarkTryLockTimeout(b *testing.B) {
	mux := New()
	b.Run("goroutine-1", func(_b *testing.B) {
		_b.ReportAllocs()
		for i := 0; i < _b.N; i++ {
			if !mux.TryLockTimeout(1 * time.Millisecond) {
				_b.FailNow()
			}
			mux.Unlock()
		}
	})

	b.Run("goroutine-N", func(_b *testing.B) {
		var countor = 0
		var wg sync.WaitGroup
		for i := 0; i < _b.N; i++ {
			_b.ReportAllocs()
			wg.Add(1)
			go func() {
				if !mux.TryLockTimeout(1000 * time.Millisecond) {
					_b.FailNow()
				}
				countor++
				mux.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()
		if countor != _b.N {
			b.Fatalf("countor(%d) != _b.N(%d)", countor, _b.N)
		}
	})
}
