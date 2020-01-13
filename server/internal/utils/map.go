package utils

import (
	"sync"
	"sync/atomic"
)

type CountedSyncMap struct {
	sync.Map
	len uint64
}

func (m *CountedSyncMap) CountedDelete(key interface{}) {
	m.Delete(key)
	atomic.AddUint64(&m.len, ^uint64(0))
}

func (m *CountedSyncMap) CountedStore(key, value interface{}) {
	m.Store(key, value)
	atomic.AddUint64(&m.len, uint64(1))
}

func (m *CountedSyncMap) CountedLen() uint64 {
	return atomic.LoadUint64(&m.len)
}
