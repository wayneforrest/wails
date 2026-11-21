//go:build windows

package w32

import "sync"

// callbackRegistry passes per-call state to a callback created once. Go never
// frees a syscall.NewCallback callback and allows about 2000 per process, so a
// per-call closure must travel through the LPARAM as an id from this registry.
type callbackRegistry[T any] struct {
	mu     sync.Mutex
	lastID uintptr
	items  map[uintptr]T
}

// add stores value and returns its id.
func (r *callbackRegistry[T]) add(value T) uintptr {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.items == nil {
		r.items = make(map[uintptr]T)
	}
	r.lastID++
	r.items[r.lastID] = value
	return r.lastID
}

// get returns the value stored under id, if any.
func (r *callbackRegistry[T]) get(id uintptr) (T, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.items[id]
	return value, ok
}

// remove forgets id.
func (r *callbackRegistry[T]) remove(id uintptr) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
}
