//go:build windows

package w32

import (
	"sync"
	"testing"
)

func TestCallbackRegistryAddGetRemove(t *testing.T) {
	var r callbackRegistry[string]

	a := r.add("a")
	b := r.add("b")
	if a == b {
		t.Fatalf("ids must be unique, both were %d", a)
	}

	if got, ok := r.get(a); !ok || got != "a" {
		t.Errorf("get(a) = %q, %v; want \"a\", true", got, ok)
	}

	r.remove(a)
	if _, ok := r.get(a); ok {
		t.Errorf("get(a) after remove still found a value")
	}
	if got, ok := r.get(b); !ok || got != "b" {
		t.Errorf("removing a disturbed b: get(b) = %q, %v", got, ok)
	}
	if len(r.items) != 1 {
		t.Errorf("registry holds %d items, want 1", len(r.items))
	}
}

func TestCallbackRegistryUnknownID(t *testing.T) {
	var r callbackRegistry[int]
	if _, ok := r.get(0); ok {
		t.Errorf("get(0) on an empty registry found a value")
	}
	r.remove(42) // Must not panic.
}

func TestCallbackRegistryConcurrent(t *testing.T) {
	var r callbackRegistry[int]
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				value := g*1000 + i
				id := r.add(value)
				if got, ok := r.get(id); !ok || got != value {
					t.Errorf("get(%d) = %d, %v; want %d, true", id, got, ok, value)
					return
				}
				r.remove(id)
			}
		}(g)
	}
	wg.Wait()
	if len(r.items) != 0 {
		t.Errorf("registry leaked %d items", len(r.items))
	}
}
