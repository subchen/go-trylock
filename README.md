# go-trylock

[![GoDoc](https://godoc.org/github.com/subchen/go-trylock?status.svg)](https://godoc.org/github.com/subchen/go-trylock)
[![Build Status](https://travis-ci.org/subchen/go-trylock.svg?branch=master)](https://travis-ci.org/subchen/go-trylock)
[![Coverage Status](https://coveralls.io/repos/github/subchen/go-trylock/badge.svg?branch=master)](https://coveralls.io/github/subchen/go-trylock?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/subchen/go-trylock)](https://goreportcard.com/report/github.com/subchen/go-trylock)
[![License](http://img.shields.io/badge/License-Apache_2-red.svg?style=flat)](http://www.apache.org/licenses/LICENSE-2.0)

TryLock support on read-write lock for Golang

## Interface

`go-trylock` implements [`sync.Locker`](https://golang.org/src/sync/mutex.go?s=881:924#L21).

Have same interfaces with [`sync.RWMutex`](https://golang.org/src/sync/rwmutex.go?s=987:1319#L18)

Documentation can be found at [Godoc](https://godoc.org/github.com/subchen/go-trylock)

## Examples

```go
import (
    "time"
    "errors"
    "github.com/subchen/go-trylock"
)

var mu = trylock.New()

func goroutineWrite() error {
    if ok := mu.TryLock(1 * time.Second); !ok {
    	return errors.New("timeout, cannot TryLock !!!")
    }
    defer mu.Unlock()
    
    // write something
}

func goroutineRead() {
    if ok := mu.RTryLock(1 * time.Second); !ok {
    	return errors.New("timeout, cannot RTryLock !!!")
    }
    defer mu.RUnlock()
    
    // read something
}
```

## LICENSE

Apache 2.0
