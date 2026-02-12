package nebo

import (
	"context"

	pb "github.com/nebolabs/nebo-sdk-go/pb"
)

// InboundMessage represents a message received from an external platform.
type InboundMessage struct {
	ChannelID string
	UserID    string
	Text      string
	Metadata  string // JSON-encoded
}

// ChannelHandler is the interface for channel capability apps.
// Implement this to bridge an external messaging platform to Nebo.
type ChannelHandler interface {
	ID() string
	Connect(ctx context.Context, config map[string]string) error
	Disconnect(ctx context.Context) error
	Send(ctx context.Context, channelID, text string) error
	Receive(ctx context.Context) (<-chan InboundMessage, error)
}

// channelBridge adapts a ChannelHandler to the pb.ChannelServiceServer gRPC interface.
type channelBridge struct {
	pb.UnimplementedChannelServiceServer
	handler     ChannelHandler
	onConfigure func(map[string]string)
	env         *AppEnv
}

func (b *channelBridge) HealthCheck(_ context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Healthy: true,
		Name:    b.env.Name,
		Version: b.env.Version,
	}, nil
}

func (b *channelBridge) ID(_ context.Context, _ *pb.Empty) (*pb.IDResponse, error) {
	return &pb.IDResponse{Id: b.handler.ID()}, nil
}

func (b *channelBridge) Connect(ctx context.Context, req *pb.ChannelConnectRequest) (*pb.ChannelConnectResponse, error) {
	if err := b.handler.Connect(ctx, req.Config); err != nil {
		return &pb.ChannelConnectResponse{Error: err.Error()}, nil
	}
	return &pb.ChannelConnectResponse{}, nil
}

func (b *channelBridge) Disconnect(ctx context.Context, _ *pb.Empty) (*pb.ChannelDisconnectResponse, error) {
	if err := b.handler.Disconnect(ctx); err != nil {
		return &pb.ChannelDisconnectResponse{Error: err.Error()}, nil
	}
	return &pb.ChannelDisconnectResponse{}, nil
}

func (b *channelBridge) Send(ctx context.Context, req *pb.ChannelSendRequest) (*pb.ChannelSendResponse, error) {
	if err := b.handler.Send(ctx, req.ChannelId, req.Text); err != nil {
		return &pb.ChannelSendResponse{Error: err.Error()}, nil
	}
	return &pb.ChannelSendResponse{}, nil
}

func (b *channelBridge) Receive(_ *pb.Empty, stream pb.ChannelService_ReceiveServer) error {
	ch, err := b.handler.Receive(stream.Context())
	if err != nil {
		return err
	}
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(&pb.InboundMessage{
				ChannelId: msg.ChannelID,
				UserId:    msg.UserID,
				Text:      msg.Text,
				Metadata:  msg.Metadata,
			}); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (b *channelBridge) Configure(_ context.Context, req *pb.SettingsMap) (*pb.Empty, error) {
	if b.onConfigure != nil {
		b.onConfigure(req.Values)
	}
	return &pb.Empty{}, nil
}
