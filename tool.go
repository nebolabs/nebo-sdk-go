package nebo

import (
	"context"
	"encoding/json"

	pb "github.com/nebolabs/nebo-sdk-go/pb"
)

// ToolHandler is the interface for tool capability apps.
// Implement this to give Nebo's agent a new tool.
type ToolHandler interface {
	Name() string
	Description() string
	Schema() json.RawMessage
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}

// ToolHandlerWithApproval is an optional extension for tools that require user confirmation.
type ToolHandlerWithApproval interface {
	ToolHandler
	RequiresApproval() bool
}

// toolBridge adapts a ToolHandler to the pb.ToolServiceServer gRPC interface.
type toolBridge struct {
	pb.UnimplementedToolServiceServer
	handler     ToolHandler
	onConfigure func(map[string]string)
	env         *AppEnv
}

func (b *toolBridge) HealthCheck(_ context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Healthy: true,
		Name:    b.env.Name,
		Version: b.env.Version,
	}, nil
}

func (b *toolBridge) Name(_ context.Context, _ *pb.Empty) (*pb.NameResponse, error) {
	return &pb.NameResponse{Name: b.handler.Name()}, nil
}

func (b *toolBridge) Description(_ context.Context, _ *pb.Empty) (*pb.DescriptionResponse, error) {
	return &pb.DescriptionResponse{Description: b.handler.Description()}, nil
}

func (b *toolBridge) Schema(_ context.Context, _ *pb.Empty) (*pb.SchemaResponse, error) {
	return &pb.SchemaResponse{Schema: b.handler.Schema()}, nil
}

func (b *toolBridge) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	content, err := b.handler.Execute(ctx, req.Input)
	if err != nil {
		return &pb.ExecuteResponse{Content: err.Error(), IsError: true}, nil
	}
	return &pb.ExecuteResponse{Content: content}, nil
}

func (b *toolBridge) RequiresApproval(_ context.Context, _ *pb.Empty) (*pb.ApprovalResponse, error) {
	if h, ok := b.handler.(ToolHandlerWithApproval); ok {
		return &pb.ApprovalResponse{RequiresApproval: h.RequiresApproval()}, nil
	}
	return &pb.ApprovalResponse{RequiresApproval: false}, nil
}

func (b *toolBridge) Configure(_ context.Context, req *pb.SettingsMap) (*pb.Empty, error) {
	if b.onConfigure != nil {
		b.onConfigure(req.Values)
	}
	return &pb.Empty{}, nil
}
