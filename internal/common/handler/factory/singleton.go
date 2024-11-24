package factory

import "sync"

type Supplier func(string) any

type Singleton struct {
	cache    map[string]any
	locker   *sync.Mutex
	supplier Supplier
}

func NewSingleton(supplier Supplier) *Singleton {
	return &Singleton{
		supplier: supplier,
		locker:   &sync.Mutex{},
		cache:    make(map[string]any),
	}
}

/*
先看缓存里面有没有，没有的话就加锁然后读取内存
*/
func (s *Singleton) Get(key string) any {
	if value, hit := s.cache[key]; hit {
		return value
	}
	s.locker.Lock()
	defer s.locker.Unlock()
	if value, hit := s.cache[key]; hit {
		return value
	}
	s.cache[key] = s.supplier(key)
	return s.cache[key]
}
