package unpack

import (
	"sync"
)

var mPool = sync.Pool{
	New: func() any { return map[string]any{} },
}

func acquireMap() map[string]any {
	m := mPool.Get().(map[string]any)
	for k := range m {
		delete(m, k)
	}
	return m
}

func releaseMap(m map[string]any) {
	for k := range m {
		delete(m, k)
	}
	mPool.Put(m)
}

var mmPool = sync.Pool{
	New: func() any { return map[string]map[string]any{} },
}

func acquireMap2Map() map[string]map[string]any {
	m := mmPool.Get().(map[string]map[string]any)
	for k := range m {
		delete(m, k)
	}
	return m
}

func releaseMap2Map(m map[string]map[string]any) {
	for k := range m {
		delete(m, k)
	}
	mmPool.Put(m)
}
