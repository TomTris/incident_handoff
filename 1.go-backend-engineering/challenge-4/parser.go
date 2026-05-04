package main

import (
	"strings"
	"sync"
)

// Zero-capacity map → resizes as keys are added.
// strings.Split(p, "=") inside the loop allocates a new []string slice every iteration.
// Both cause unnecessary heap allocations.
func ParseA(line string) map[string]string {
	result := map[string]string{}
	parts := strings.Split(line, "|")
	for _, p := range parts {
		kv := strings.Split(strings.TrimSpace(p), "=")
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

// Pre-sized map with len(pairs) → no resize.
// strings.Index instead of Split inside the loop → no slice allocation per iteration.
// Still allocates a new map every call → GC pressure remains.
func ParseB(line string) map[string]string {
	pairs := strings.Split(line, "|")
	result := make(map[string]string, len(pairs))
	for _, p := range pairs {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, "="); idx != -1 {
			result[p[:idx]] = p[idx+1:]
		}
	}
	return result
}

var bufPool = sync.Pool{
	New: func() any {
		return make(map[string]string, 10)
	},
}

func getMap() map[string]string {
	return bufPool.Get().(map[string]string)
}

func ReleaseMap(m map[string]string) {
	bufPool.Put(m)
}

// Reuses map from pool → no map allocation after the first call.
// GC can clear the pool between cycles, causing a reallocation, but only rare
func ParseC(line string) map[string]string {
	pairs := strings.Split(line, "|")
	result := getMap()
	clear(result)
	for _, p := range pairs {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, "="); idx != -1 {
			result[p[:idx]] = p[idx+1:]
		}
	}
	return result
}
