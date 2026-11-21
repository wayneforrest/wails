//go:build windows

package w32

import (
	"sync"
	"testing"
)

// enumCallbackStressIterations is well past Go's per-process limit on Windows
// callbacks (about 2000, see cb_max in runtime/syscall_windows.go). Callback
// slots are never freed, and every distinct closure handed to
// syscall.NewCallback takes a new one, so an enumeration that builds its
// callback per call dies with a runtime fatal error ("too many callback
// functions") long before this many calls. The fatal error cannot be
// recovered, so a regression kills the test binary.
const enumCallbackStressIterations = 3000

// TestGetAllScreensDoesNotExhaustCallbacks is the regression test for the
// production crash where WM_DISPLAYCHANGE / WM_SETTINGCHANGE re-listed the
// screens thousands of times (docks, display churn) and each GetAllScreens
// call consumed a new callback slot until the runtime threw.
func TestGetAllScreensDoesNotExhaustCallbacks(t *testing.T) {
	firstScreens, firstErr := GetAllScreens()

	for i := 0; i < enumCallbackStressIterations; i++ {
		screens, err := GetAllScreens()
		if (err == nil) != (firstErr == nil) {
			t.Fatalf("iteration %d: error changed from %v to %v", i, firstErr, err)
		}
		if len(screens) != len(firstScreens) {
			t.Fatalf("iteration %d: got %d screens, first call got %d", i, len(screens), len(firstScreens))
		}
	}
}

// TestEnumChildWindowsDoesNotExhaustCallbacks covers the same leak in
// EnumChildWindows. The callback captures a per-iteration variable so every
// call passes a distinct closure, which is how real callers use it and what
// used to take a new callback slot each time.
func TestEnumChildWindowsDoesNotExhaustCallbacks(t *testing.T) {
	desktop := GetDesktopWindow()

	for i := 0; i < enumCallbackStressIterations; i++ {
		visited := 0
		EnumChildWindows(desktop, func(hwnd HWND, lparam LPARAM) LRESULT {
			visited++
			return 0 // Stop after the first child
		})
		if visited > 1 {
			t.Fatalf("iteration %d: callback ran %d times after asking to stop", i, visited)
		}
	}
}

// TestEnumChildWindowsPassesZeroLParam pins the existing contract: the
// caller's callback has always received an LPARAM of 0.
func TestEnumChildWindowsPassesZeroLParam(t *testing.T) {
	EnumChildWindows(GetDesktopWindow(), func(hwnd HWND, lparam LPARAM) LRESULT {
		if lparam != 0 {
			t.Errorf("callback got lparam %d, want 0", lparam)
		}
		return 0
	})
}

// TestGetAllScreensConcurrent checks that concurrent enumerations from
// different goroutines each get their own, complete result.
func TestGetAllScreensConcurrent(t *testing.T) {
	want, wantErr := GetAllScreens()

	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				screens, err := GetAllScreens()
				if (err == nil) != (wantErr == nil) || len(screens) != len(want) {
					t.Errorf("concurrent call got %d screens (err %v), want %d (err %v)", len(screens), err, len(want), wantErr)
					return
				}
			}
		}()
	}
	wg.Wait()
}
