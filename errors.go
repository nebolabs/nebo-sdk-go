package nebo

import "errors"

var (
	// ErrNoSockPath is returned when NEBO_APP_SOCK is not set.
	ErrNoSockPath = errors.New("NEBO_APP_SOCK environment variable is not set")

	// ErrNoHandlers is returned when Run() is called with no registered handlers.
	ErrNoHandlers = errors.New("no capability handlers registered — register at least one handler before calling Run()")
)
