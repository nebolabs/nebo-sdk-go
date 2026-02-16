package nebo

import (
	"os"
	"testing"
)

func TestNewRequiresSockPath(t *testing.T) {
	os.Unsetenv("NEBO_APP_SOCK")
	_, err := New()
	if err != ErrNoSockPath {
		t.Fatalf("expected ErrNoSockPath, got %v", err)
	}
}

func TestNewWithSockPath(t *testing.T) {
	t.Setenv("NEBO_APP_SOCK", "/tmp/test-nebo.sock")
	t.Setenv("NEBO_APP_NAME", "test-app")

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if app.env.SockPath != "/tmp/test-nebo.sock" {
		t.Errorf("SockPath = %q, want /tmp/test-nebo.sock", app.env.SockPath)
	}
	if app.env.Name != "test-app" {
		t.Errorf("Name = %q, want test-app", app.env.Name)
	}
}

func TestRunWithNoHandlers(t *testing.T) {
	t.Setenv("NEBO_APP_SOCK", "/tmp/test-nebo.sock")

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	err = app.Run()
	if err != ErrNoHandlers {
		t.Fatalf("expected ErrNoHandlers, got %v", err)
	}
}
