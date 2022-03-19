package Rxorm

import (
	"sync"
	"reflect"
)

type Call struct {
	wg sync.WaitGroup
	val interface{} //返回值
	isCache bool
	err error
}

type SingleFlight struct {
	mu sync.Mutex
	flights map[string]*Call
}

type Request func(bean interface{}) (bool, error)

// 放在engine里面
func (s *SingleFlight) Do(key string, bean interface{}, fn Request)(bool, error){
	s.mu.Lock()
	if s.flights == nil{
		s.flights = make(map[string]*Call)
	}
	// 1.如果存在已有请求
	c, ok := s.flights[key]
	if ok{
		s.mu.Unlock()
		c.wg.Wait()
		// 这里直接赋值指针的话不太恰当，使用reflect更合适
		reflect.ValueOf(bean).Elem().Set(reflect.ValueOf(c.val).Elem())
		return true, c.err
	} 

	// 2.未有请求，访问缓存
	call := &Call{val: bean}
	call.wg.Add(1)
	s.flights[key] = call
	s.mu.Unlock()

	call.isCache, call.err = fn(call.val)
	call.wg.Done()

	s.mu.Lock()
	delete(s.flights, key)
	s.mu.Unlock()

	return call.isCache, call.err
}