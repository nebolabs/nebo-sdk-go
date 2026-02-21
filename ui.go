package nebo

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	pb "github.com/neboloop/nebo-sdk-go/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// uiBridge adapts HandleFunc/Handle to the pb.UIServiceServer gRPC interface.
type uiBridge struct {
	pb.UnimplementedUIServiceServer
	mux         *http.ServeMux
	onConfigure func(map[string]string)
	env         *AppEnv
}

func (b *uiBridge) HealthCheck(_ context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Healthy: true,
		Name:    b.env.Name,
		Version: b.env.Version,
	}, nil
}

func (b *uiBridge) Configure(_ context.Context, req *pb.SettingsMap) (*pb.Empty, error) {
	if b.onConfigure != nil {
		b.onConfigure(req.Values)
	}
	return &pb.Empty{}, nil
}

// HandleRequest dispatches a proxied HTTP request through the app's http.ServeMux.
func (b *uiBridge) HandleRequest(ctx context.Context, req *pb.HttpRequest) (*pb.HttpResponse, error) {
	if b.mux == nil {
		return nil, status.Error(codes.Unimplemented, "no HTTP handlers registered")
	}

	// Build path with query string
	uri := req.Path
	if req.Query != "" {
		uri += "?" + req.Query
	}

	// Build *http.Request from proto
	var body io.Reader
	if len(req.Body) > 0 {
		body = strings.NewReader(string(req.Body))
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, uri, body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "build request: %v", err)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Dispatch through standard net/http
	rec := httptest.NewRecorder()
	b.mux.ServeHTTP(rec, httpReq)

	// Convert response
	result := rec.Result()
	respBody, _ := io.ReadAll(result.Body)
	result.Body.Close()

	headers := make(map[string]string, len(result.Header))
	for k := range result.Header {
		headers[k] = result.Header.Get(k)
	}

	return &pb.HttpResponse{
		StatusCode: int32(result.StatusCode),
		Headers:    headers,
		Body:       respBody,
	}, nil
}
