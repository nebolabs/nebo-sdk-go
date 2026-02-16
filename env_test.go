package nebo

import (
	"os"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	t.Setenv("NEBO_APP_DIR", "/apps/test")
	t.Setenv("NEBO_APP_SOCK", "/tmp/test.sock")
	t.Setenv("NEBO_APP_ID", "com.example.test")
	t.Setenv("NEBO_APP_NAME", "Test App")
	t.Setenv("NEBO_APP_VERSION", "1.2.3")
	t.Setenv("NEBO_APP_DATA", "/apps/test/data")

	env := loadEnv()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Dir", env.Dir, "/apps/test"},
		{"SockPath", env.SockPath, "/tmp/test.sock"},
		{"ID", env.ID, "com.example.test"},
		{"Name", env.Name, "Test App"},
		{"Version", env.Version, "1.2.3"},
		{"DataDir", env.DataDir, "/apps/test/data"},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestLoadEnvMissing(t *testing.T) {
	// Clear all env vars
	for _, key := range []string{"NEBO_APP_DIR", "NEBO_APP_SOCK", "NEBO_APP_ID", "NEBO_APP_NAME", "NEBO_APP_VERSION", "NEBO_APP_DATA"} {
		os.Unsetenv(key)
	}

	env := loadEnv()

	if env.SockPath != "" {
		t.Errorf("SockPath = %q, want empty", env.SockPath)
	}
	if env.Name != "" {
		t.Errorf("Name = %q, want empty", env.Name)
	}
}
