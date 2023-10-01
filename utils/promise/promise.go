package promise

import (
	"context"
	"sync"
)

type Promise struct {
	ctx      context.Context
	cancel   context.CancelFunc
	resultCh chan interface{}
	errCh    chan error
}

func NewPromise() *Promise {
	ctx, cancel := context.WithCancel(context.Background())
	return &Promise{
		ctx:      ctx,
		cancel:   cancel,
		resultCh: make(chan interface{}, 1),
		errCh:    make(chan error, 1),
	}
}

func (p *Promise) Resolve(value interface{}) {
	select {
	case <-p.ctx.Done():
		return
	case p.resultCh <- value:
	}
}

func (p *Promise) Reject(err error) {
	select {
	case <-p.ctx.Done():
		return
	case p.errCh <- err:
	}
}

func (p *Promise) Wait() (interface{}, error) {
	select {
	case <-p.ctx.Done():
		return nil, p.ctx.Err()
	case result := <-p.resultCh:
		return result, nil
	case err := <-p.errCh:
		return nil, err
	}
}

func All(promiseFuncs ...func(*Promise)) *Promise {
	numPromises := len(promiseFuncs)
	promiseAll := NewPromise()

	results := make([]interface{}, numPromises)
	count := 0
	for i, promiseFunc := range promiseFuncs {
		go func(i int, promiseFunc func(*Promise)) {
			promise := NewPromise()
			promiseFunc(promise)
			result, err := promise.Wait()
			if err != nil {
				promiseAll.Reject(err)
				return
			}
			results[i] = result
			count++
			if count == numPromises {
				promiseAll.Resolve(results)
			}
		}(i, promiseFunc)
	}

	return promiseAll
}

func AllSettled(promiseFuncs ...func(*Promise)) *Promise {
	numPromises := len(promiseFuncs)
	promiseAll := NewPromise()

	results := make([]interface{}, numPromises)
	count := 0
	for i, promiseFunc := range promiseFuncs {
		go func(i int, promiseFunc func(*Promise)) {
			promise := NewPromise()
			promiseFunc(promise)
			result, _ := promise.Wait()

			results[i] = result
			count++
			if count == numPromises {
				promiseAll.Resolve(results)
			}
		}(i, promiseFunc)
	}

	return promiseAll
}

func Parallel(promiseFuncs ...func()) {
	numPromises := len(promiseFuncs)
	promiseAll := NewPromise()

	results := make([]interface{}, numPromises)
	count := 0
	resolveOnce := sync.Once{}

	for i, promiseFunc := range promiseFuncs {
		go func(i int, promiseFunc func()) {
			promiseFunc()

			count++
			if count == numPromises {
				// Use sync.Once to ensure Resolve is called only once
				resolveOnce.Do(func() {
					promiseAll.Resolve(results)
				})
			}
		}(i, promiseFunc)
	}

	promiseAll.Wait()
}
