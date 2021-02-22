# gpool
[![Build Status](https://travis-ci.org/archdx/gpool.svg?branch=main)](https://travis-ci.org/archdx/gpool)
[![codecov](https://codecov.io/gh/archdx/gpool/branch/main/graph/badge.svg)](https://codecov.io/gh/archdx/gpool)

### Description
Package gpool implements easy to use goroutines pool. Although creating goroutines is cheap, it can be useful in some highload scenarios to limit goroutine spawning and monitor consumption.

### Usage
Create new pool
```go
pool := gpool.NewPool(1024)
defer pool.Close()
```

Goroutine is acquired with **G()** method and should be released after job is done
```go
g := pool.G()
defer g.Release()

g.Exec(someFunc())
```
Also gpool can be used to create workers in cases where long lived workers are not suitable. For example when executing code in context of incoming http request.
It can be building batch proxy on top of storage with only single key access supported:
```go
func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
    var mu sync.Mutex
    var result []Object

    idCh := make(chan int)

    gr := pool.GGroup(8)

    gr.Exec(func() {
        for id := range idCh {
            obj := storage.queryObject(id)
        
            mu.Lock()
            result = append(result, obj)
            mu.Unlock()
        }
    })

    for _, id := range parseObjectIDs(r) {
        idCh <- id
    }

    close(idCh)
    gr.WaitAndRelease()
}
```
