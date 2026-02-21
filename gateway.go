package nebo

import (
	"context"

	pb "github.com/neboloop/nebo-sdk-go/pb"
)

// GatewayRequest represents an LLM chat completion request from Nebo.
type GatewayRequest struct {
	RequestID   string
	Messages    []GatewayMessage
	Tools       []GatewayToolDef
	MaxTokens   int32
	Temperature float64
	System      string
	UserID      string
	UserPlan    string
	UserToken   string // Only populated if app has "user:token" permission
}

// GatewayMessage is a single message in a conversation.
type GatewayMessage struct {
	Role       string
	Content    string
	ToolCallID string
	ToolCalls  string // JSON-encoded array
}

// GatewayToolDef describes a tool available to the model.
type GatewayToolDef struct {
	Name        string
	Description string
	InputSchema []byte // JSON Schema
}

// GatewayEvent is a streamed event sent back to Nebo.
type GatewayEvent struct {
	Type      string // "text", "tool_call", "thinking", "error", "done"
	Content   string
	Model     string
	RequestID string
}

// GatewayHandler is the interface for gateway capability apps.
// Implement this to provide LLM model routing to Nebo.
type GatewayHandler interface {
	Stream(ctx context.Context, req *GatewayRequest) (<-chan GatewayEvent, error)
	Cancel(ctx context.Context, requestID string) error
}

// gatewayBridge adapts a GatewayHandler to the pb.GatewayServiceServer gRPC interface.
type gatewayBridge struct {
	pb.UnimplementedGatewayServiceServer
	handler     GatewayHandler
	onConfigure func(map[string]string)
	env         *AppEnv
}

func (b *gatewayBridge) HealthCheck(_ context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Healthy: true,
		Name:    b.env.Name,
		Version: b.env.Version,
	}, nil
}

func (b *gatewayBridge) Stream(req *pb.GatewayRequest, stream pb.GatewayService_StreamServer) error {
	gwReq := &GatewayRequest{
		RequestID:   req.RequestId,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		System:      req.System,
	}
	if req.User != nil {
		gwReq.UserID = req.User.UserId
		gwReq.UserPlan = req.User.Plan
		gwReq.UserToken = req.User.Token
	}
	for _, m := range req.Messages {
		gwReq.Messages = append(gwReq.Messages, GatewayMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallId,
			ToolCalls:  m.ToolCalls,
		})
	}
	for _, t := range req.Tools {
		gwReq.Tools = append(gwReq.Tools, GatewayToolDef{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}

	ch, err := b.handler.Stream(stream.Context(), gwReq)
	if err != nil {
		return err
	}
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(&pb.GatewayEvent{
				Type:      event.Type,
				Content:   event.Content,
				Model:     event.Model,
				RequestId: event.RequestID,
			}); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (b *gatewayBridge) Cancel(ctx context.Context, req *pb.CancelRequest) (*pb.CancelResponse, error) {
	if err := b.handler.Cancel(ctx, req.RequestId); err != nil {
		return &pb.CancelResponse{Cancelled: false}, nil
	}
	return &pb.CancelResponse{Cancelled: true}, nil
}

func (b *gatewayBridge) Poll(_ context.Context, _ *pb.PollRequest) (*pb.PollResponse, error) {
	return &pb.PollResponse{}, nil
}

func (b *gatewayBridge) Configure(_ context.Context, req *pb.SettingsMap) (*pb.Empty, error) {
	if b.onConfigure != nil {
		b.onConfigure(req.Values)
	}
	return &pb.Empty{}, nil
}
