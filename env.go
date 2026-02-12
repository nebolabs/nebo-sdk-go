package nebo

import "os"

// AppEnv provides typed access to NEBO_APP_* environment variables.
// These are set by Nebo's sandbox when launching your app.
type AppEnv struct {
	Dir      string // NEBO_APP_DIR — app's installation directory
	SockPath string // NEBO_APP_SOCK — Unix socket path to listen on
	ID       string // NEBO_APP_ID — app ID from manifest
	Name     string // NEBO_APP_NAME — app name from manifest
	Version  string // NEBO_APP_VERSION — app version from manifest
	DataDir  string // NEBO_APP_DATA — path to app's data/ directory
}

func loadEnv() *AppEnv {
	return &AppEnv{
		Dir:      os.Getenv("NEBO_APP_DIR"),
		SockPath: os.Getenv("NEBO_APP_SOCK"),
		ID:       os.Getenv("NEBO_APP_ID"),
		Name:     os.Getenv("NEBO_APP_NAME"),
		Version:  os.Getenv("NEBO_APP_VERSION"),
		DataDir:  os.Getenv("NEBO_APP_DATA"),
	}
}
