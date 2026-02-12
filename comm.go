package nebo

import (
	"context"

	pb "github.com/nebolabs/nebo-sdk-go/pb"
)

// CommMessage represents an inter-agent communication message.
type CommMessage struct {
	ID             string
	From           string
	To             string
	Topic          string
	ConversationID string
	Type           string // "message", "mention", "proposal", "command", "info", "task"
	Content        string
	Metadata       map[string]string
	Timestamp      int64
	HumanInjected  bool
	HumanID        string
}

// CommHandler is the interface for comm capability apps.
// Implement this to provide inter-agent communication for Nebo.
type CommHandler interface {
	Name() string
	Version() string
	Connect(ctx context.Context, config map[string]string) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	Send(ctx context.Context, msg CommMessage) error
	Subscribe(ctx context.Context, topic string) error
	Unsubscribe(ctx context.Context, topic string) error
	Register(ctx context.Context, agentID string, capabilities []string) error
	Deregister(ctx context.Context) error
	Receive(ctx context.Context) (<-chan CommMessage, error)
}

// commBridge adapts a CommHandler to the pb.CommServiceServer gRPC interface.
type commBridge struct {
	pb.UnimplementedCommServiceServer
	handler     CommHandler
	onConfigure func(map[string]string)
	env         *AppEnv
}

func (b *commBridge) HealthCheck(_ context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Healthy: true,
		Name:    b.env.Name,
		Version: b.env.Version,
	}, nil
}

func (b *commBridge) Name(_ context.Context, _ *pb.Empty) (*pb.CommNameResponse, error) {
	return &pb.CommNameResponse{Name: b.handler.Name()}, nil
}

func (b *commBridge) Version(_ context.Context, _ *pb.Empty) (*pb.CommVersionResponse, error) {
	return &pb.CommVersionResponse{Version: b.handler.Version()}, nil
}

func (b *commBridge) Connect(ctx context.Context, req *pb.CommConnectRequest) (*pb.CommConnectResponse, error) {
	if err := b.handler.Connect(ctx, req.Config); err != nil {
		return &pb.CommConnectResponse{Error: err.Error()}, nil
	}
	return &pb.CommConnectResponse{}, nil
}

func (b *commBridge) Disconnect(ctx context.Context, _ *pb.Empty) (*pb.CommDisconnectResponse, error) {
	if err := b.handler.Disconnect(ctx); err != nil {
		return &pb.CommDisconnectResponse{Error: err.Error()}, nil
	}
	return &pb.CommDisconnectResponse{}, nil
}

func (b *commBridge) IsConnected(_ context.Context, _ *pb.Empty) (*pb.CommIsConnectedResponse, error) {
	return &pb.CommIsConnectedResponse{Connected: b.handler.IsConnected()}, nil
}

func (b *commBridge) Send(ctx context.Context, req *pb.CommSendRequest) (*pb.CommSendResponse, error) {
	msg := fromProtoCommMsg(req.Message)
	if err := b.handler.Send(ctx, msg); err != nil {
		return &pb.CommSendResponse{Error: err.Error()}, nil
	}
	return &pb.CommSendResponse{}, nil
}

func (b *commBridge) Subscribe(ctx context.Context, req *pb.CommSubscribeRequest) (*pb.CommSubscribeResponse, error) {
	if err := b.handler.Subscribe(ctx, req.Topic); err != nil {
		return &pb.CommSubscribeResponse{Error: err.Error()}, nil
	}
	return &pb.CommSubscribeResponse{}, nil
}

func (b *commBridge) Unsubscribe(ctx context.Context, req *pb.CommUnsubscribeRequest) (*pb.CommUnsubscribeResponse, error) {
	if err := b.handler.Unsubscribe(ctx, req.Topic); err != nil {
		return &pb.CommUnsubscribeResponse{Error: err.Error()}, nil
	}
	return &pb.CommUnsubscribeResponse{}, nil
}

func (b *commBridge) Register(ctx context.Context, req *pb.CommRegisterRequest) (*pb.CommRegisterResponse, error) {
	if err := b.handler.Register(ctx, req.AgentId, req.Capabilities); err != nil {
		return &pb.CommRegisterResponse{Error: err.Error()}, nil
	}
	return &pb.CommRegisterResponse{}, nil
}

func (b *commBridge) Deregister(ctx context.Context, _ *pb.Empty) (*pb.CommDeregisterResponse, error) {
	if err := b.handler.Deregister(ctx); err != nil {
		return &pb.CommDeregisterResponse{Error: err.Error()}, nil
	}
	return &pb.CommDeregisterResponse{}, nil
}

func (b *commBridge) Receive(_ *pb.Empty, stream pb.CommService_ReceiveServer) error {
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
			if err := stream.Send(toProtoCommMsg(msg)); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (b *commBridge) Configure(_ context.Context, req *pb.SettingsMap) (*pb.Empty, error) {
	if b.onConfigure != nil {
		b.onConfigure(req.Values)
	}
	return &pb.Empty{}, nil
}

func toProtoCommMsg(m CommMessage) *pb.CommMessage {
	return &pb.CommMessage{
		Id:             m.ID,
		From:           m.From,
		To:             m.To,
		Topic:          m.Topic,
		ConversationId: m.ConversationID,
		Type:           m.Type,
		Content:        m.Content,
		Metadata:       m.Metadata,
		Timestamp:      m.Timestamp,
		HumanInjected:  m.HumanInjected,
		HumanId:        m.HumanID,
	}
}

func fromProtoCommMsg(m *pb.CommMessage) CommMessage {
	if m == nil {
		return CommMessage{}
	}
	return CommMessage{
		ID:             m.Id,
		From:           m.From,
		To:             m.To,
		Topic:          m.Topic,
		ConversationID: m.ConversationId,
		Type:           m.Type,
		Content:        m.Content,
		Metadata:       m.Metadata,
		Timestamp:      m.Timestamp,
		HumanInjected:  m.HumanInjected,
		HumanID:        m.HumanId,
	}
}
