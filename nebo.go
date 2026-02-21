// Package nebo provides the SDK for building Nebo apps.
//
// A Nebo app is a compiled binary that communicates with Nebo over gRPC via Unix sockets.
// The SDK handles all gRPC server setup, signal handling, and protocol bridging — you just
// implement handler interfaces for the capabilities your app provides.
//
// Quick start:
//
//	app, _ := nebo.New()
//	app.RegisterTool(&MyTool{})
//	log.Fatal(app.Run())
package nebo

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/neboloop/nebo-sdk-go/pb"
	"google.golang.org/grpc"
)

// App is the main entry point for a Nebo app. It manages the gRPC server,
// capability registration, and lifecycle.
type App struct {
	env         *AppEnv
	server      *grpc.Server
	onConfigure func(map[string]string)
	hasHandlers bool
	mux         *http.ServeMux
}

// New creates a new Nebo App. It reads NEBO_APP_* environment variables and
// prepares the gRPC server. Returns ErrNoSockPath if NEBO_APP_SOCK is not set.
func New() (*App, error) {
	env := loadEnv()
	if env.SockPath == "" {
		return nil, ErrNoSockPath
	}
	return &App{
		env:    env,
		server: grpc.NewServer(),
	}, nil
}

// Env returns the app's environment variables.
func (a *App) Env() *AppEnv {
	return a.env
}

// OnConfigure sets a callback that is called when Nebo pushes settings updates.
// This is shared across all capability handlers.
func (a *App) OnConfigure(fn func(map[string]string)) {
	a.onConfigure = fn
}

// RegisterTool registers a ToolHandler capability.
func (a *App) RegisterTool(h ToolHandler) {
	pb.RegisterToolServiceServer(a.server, &toolBridge{
		handler:     h,
		onConfigure: a.onConfigure,
		env:         a.env,
	})
	a.hasHandlers = true
}

// RegisterChannel registers a ChannelHandler capability.
func (a *App) RegisterChannel(h ChannelHandler) {
	pb.RegisterChannelServiceServer(a.server, &channelBridge{
		handler:     h,
		onConfigure: a.onConfigure,
		env:         a.env,
	})
	a.hasHandlers = true
}

// RegisterGateway registers a GatewayHandler capability.
func (a *App) RegisterGateway(h GatewayHandler) {
	pb.RegisterGatewayServiceServer(a.server, &gatewayBridge{
		handler:     h,
		onConfigure: a.onConfigure,
		env:         a.env,
	})
	a.hasHandlers = true
}

// HandleFunc registers an HTTP handler function for the given pattern.
// Identical to http.ServeMux.HandleFunc — same name, same signature.
// The app's HTTP handlers are called when the browser makes fetch() requests
// to /apps/{id}/api/{path}.
func (a *App) HandleFunc(pattern string, handler http.HandlerFunc) {
	if a.mux == nil {
		a.mux = http.NewServeMux()
	}
	a.mux.HandleFunc(pattern, handler)
	a.hasHandlers = true
}

// Handle registers an HTTP handler for the given pattern.
// Accepts any http.Handler — chi routers, gorilla/mux, middleware chains.
func (a *App) Handle(pattern string, handler http.Handler) {
	if a.mux == nil {
		a.mux = http.NewServeMux()
	}
	a.mux.Handle(pattern, handler)
	a.hasHandlers = true
}

// RegisterComm registers a CommHandler capability.
func (a *App) RegisterComm(h CommHandler) {
	pb.RegisterCommServiceServer(a.server, &commBridge{
		handler:     h,
		onConfigure: a.onConfigure,
		env:         a.env,
	})
	a.hasHandlers = true
}

// RegisterSchedule registers a ScheduleHandler capability.
func (a *App) RegisterSchedule(h ScheduleHandler) {
	pb.RegisterScheduleServiceServer(a.server, &scheduleBridge{
		handler:     h,
		onConfigure: a.onConfigure,
		env:         a.env,
	})
	a.hasHandlers = true
}

// Run starts the gRPC server on the Unix socket and blocks until SIGTERM/SIGINT.
// It removes any stale socket file before listening.
func (a *App) Run() error {
	if !a.hasHandlers {
		return ErrNoHandlers
	}

	// Register UI service if HandleFunc/Handle was called
	if a.mux != nil {
		pb.RegisterUIServiceServer(a.server, &uiBridge{
			mux:         a.mux,
			onConfigure: a.onConfigure,
			env:         a.env,
		})
	}

	// Remove stale socket from previous run
	os.Remove(a.env.SockPath)

	listener, err := net.Listen("unix", a.env.SockPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", a.env.SockPath, err)
	}

	// Graceful shutdown on SIGTERM/SIGINT
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		a.server.GracefulStop()
	}()

	fmt.Fprintf(os.Stderr, "[%s] listening on %s\n", a.env.Name, a.env.SockPath)
	return a.server.Serve(listener)
}
