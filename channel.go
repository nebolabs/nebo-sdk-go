package nebo

import (
	"context"

	pb "github.com/neboloop/nebo-sdk-go/pb"
)

// MessageSender identifies who sent a message.
type MessageSender struct {
	Name  string
	Role  string // relationship dynamic (Friend, COO, Mentor)
	BotID string // NeboLoop bot UUID
}

// Attachment represents a file or media attachment.
type Attachment struct {
	Type     string // "image", "file", "audio", "video"
	URL      string
	Filename string
	Size     int64 // bytes
}

// MessageAction represents an interactive element (button, keyboard row).
type MessageAction struct {
	Label      string
	CallbackID string
}

// ChannelEnvelope is the v1 message envelope used for both inbound and outbound messages.
type ChannelEnvelope struct {
	MessageID    string
	ChannelID    string
	Sender       MessageSender
	Text         string
	Attachments  []Attachment
	ReplyTo      string
	Actions      []MessageAction
	PlatformData []byte
	Timestamp    string // RFC3339

	// Legacy fields (inbound only)
	UserID   string
	Metadata string // JSON-encoded
}

// ChannelHandler is the interface for channel capability apps.
// Implement this to bridge an external messaging platform to Nebo.
type ChannelHandler interface {
	ID() string
	Connect(ctx context.Context, config map[string]string) error
	Disconnect(ctx context.Context) error
	Send(ctx context.Context, env ChannelEnvelope) (string, error)
	Receive(ctx context.Context) (<-chan ChannelEnvelope, error)
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
	env := ChannelEnvelope{
		ChannelID:    req.ChannelId,
		Text:         req.Text,
		MessageID:    req.MessageId,
		ReplyTo:      req.ReplyTo,
		PlatformData: req.PlatformData,
	}
	if req.Sender != nil {
		env.Sender = MessageSender{
			Name:  req.Sender.Name,
			Role:  req.Sender.Role,
			BotID: req.Sender.BotId,
		}
	}
	for _, a := range req.Attachments {
		env.Attachments = append(env.Attachments, Attachment{
			Type:     a.Type,
			URL:      a.Url,
			Filename: a.Filename,
			Size:     a.Size,
		})
	}
	for _, a := range req.Actions {
		env.Actions = append(env.Actions, MessageAction{
			Label:      a.Label,
			CallbackID: a.CallbackId,
		})
	}

	messageID, err := b.handler.Send(ctx, env)
	if err != nil {
		return &pb.ChannelSendResponse{Error: err.Error()}, nil
	}
	return &pb.ChannelSendResponse{MessageId: messageID}, nil
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
			pbMsg := &pb.InboundMessage{
				ChannelId:    msg.ChannelID,
				UserId:       msg.UserID,
				Text:         msg.Text,
				Metadata:     msg.Metadata,
				MessageId:    msg.MessageID,
				ReplyTo:      msg.ReplyTo,
				PlatformData: msg.PlatformData,
				Timestamp:    msg.Timestamp,
			}
			if msg.Sender != (MessageSender{}) {
				pbMsg.Sender = &pb.MessageSender{
					Name:  msg.Sender.Name,
					Role:  msg.Sender.Role,
					BotId: msg.Sender.BotID,
				}
			}
			for _, a := range msg.Attachments {
				pbMsg.Attachments = append(pbMsg.Attachments, &pb.Attachment{
					Type:     a.Type,
					Url:      a.URL,
					Filename: a.Filename,
					Size:     a.Size,
				})
			}
			for _, a := range msg.Actions {
				pbMsg.Actions = append(pbMsg.Actions, &pb.MessageAction{
					Label:      a.Label,
					CallbackId: a.CallbackID,
				})
			}
			if err := stream.Send(pbMsg); err != nil {
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
