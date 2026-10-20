//go:build !windows

package teacher

import "testing"

func TestExitCodeUnixSignal(t *testing.T) {
	code, err := OSRunner{}.Run("sh", []string{"-c", "kill -TERM $$"}, EnvFromOS())
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if code != 143 {
		t.Errorf("got code %d, want 143", code)
	}
}
