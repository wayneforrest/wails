package w32

import "sync"

// callbackRegistry hands per-call Go state to a Windows callback that is
// created only once.
//
// Go never frees a callback made by syscall.NewCallback and allows only about
// 2000 per process, so a callback must not be built from a fresh closure on
// every call. Instead the callback is a package-level function, and the state
// it needs travels through the API's LPARAM as an id from this registry.
// The id is registered before the synchronous Windows call and removed after
// it returns. Safe for concurrent use.
type callbackRegistry[T any] struct {
	mu     sync.Mutex
	lastID uintptr
	items  map[uintptr]T
}

// add stores value and returns its id. The id is never 0.
func (r *callbackRegistry[T]) add(value T) uintptr {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.items == nil {
		r.items = make(map[uintptr]T)
	}
	r.lastID++
	if r.lastID == 0 {
		r.lastID++
	}
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
