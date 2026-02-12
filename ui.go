package nebo

import (
	"context"

	pb "github.com/nebolabs/nebo-sdk-go/pb"
)

// View represents a complete renderable UI view.
type View struct {
	ViewID string
	Title  string
	Blocks []UIBlock
}

// UIBlock is a single renderable element in a view.
type UIBlock struct {
	BlockID     string
	Type        string // text, heading, input, button, select, toggle, divider, image
	Text        string
	Value       string
	Placeholder string
	Hint        string
	Variant     string // button: primary/secondary/ghost/error; heading: h1/h2/h3
	Src         string // image source URL
	Alt         string // image alt text
	Disabled    bool
	Options     []SelectOption
	Style       string // compact, full-width
}

// SelectOption is a choice for select-type blocks.
type SelectOption struct {
	Label string
	Value string
}

// UIEvent represents a user interaction with a UI block.
type UIEvent struct {
	ViewID  string
	BlockID string
	Action  string // "click", "change", "submit"
	Value   string
}

// UIEventResult is the response to a UI event.
type UIEventResult struct {
	View  *View  // Updated view (optional)
	Error string // Error message (optional)
	Toast string // Toast notification (optional)
}

// UIHandler is the interface for UI capability apps.
// Implement this to render structured panels in Nebo's web interface.
type UIHandler interface {
	GetView(ctx context.Context, viewContext string) (*View, error)
	OnEvent(ctx context.Context, event UIEvent) (*UIEventResult, error)
}

// UIHandlerWithStreaming is an optional extension for apps that push live UI updates.
type UIHandlerWithStreaming interface {
	UIHandler
	StreamUpdates(ctx context.Context) (<-chan *View, error)
}

// ViewBuilder constructs a View with a fluent API.
type ViewBuilder struct {
	viewID string
	title  string
	blocks []UIBlock
}

// NewView creates a new ViewBuilder.
func NewView(viewID, title string) *ViewBuilder {
	return &ViewBuilder{viewID: viewID, title: title}
}

// Heading adds a heading block.
func (v *ViewBuilder) Heading(blockID, text, variant string) *ViewBuilder {
	v.blocks = append(v.blocks, UIBlock{BlockID: blockID, Type: "heading", Text: text, Variant: variant})
	return v
}

// Text adds a text block.
func (v *ViewBuilder) Text(blockID, text string) *ViewBuilder {
	v.blocks = append(v.blocks, UIBlock{BlockID: blockID, Type: "text", Text: text})
	return v
}

// Button adds a button block.
func (v *ViewBuilder) Button(blockID, text, variant string) *ViewBuilder {
	v.blocks = append(v.blocks, UIBlock{BlockID: blockID, Type: "button", Text: text, Variant: variant})
	return v
}

// Input adds an input block.
func (v *ViewBuilder) Input(blockID, value, placeholder string) *ViewBuilder {
	v.blocks = append(v.blocks, UIBlock{BlockID: blockID, Type: "input", Value: value, Placeholder: placeholder})
	return v
}

// Select adds a select block.
func (v *ViewBuilder) Select(blockID, value string, options []SelectOption) *ViewBuilder {
	v.blocks = append(v.blocks, UIBlock{BlockID: blockID, Type: "select", Value: value, Options: options})
	return v
}

// Toggle adds a toggle block.
func (v *ViewBuilder) Toggle(blockID, text string, on bool) *ViewBuilder {
	val := "false"
	if on {
		val = "true"
	}
	v.blocks = append(v.blocks, UIBlock{BlockID: blockID, Type: "toggle", Text: text, Value: val})
	return v
}

// Divider adds a divider block.
func (v *ViewBuilder) Divider(blockID string) *ViewBuilder {
	v.blocks = append(v.blocks, UIBlock{BlockID: blockID, Type: "divider"})
	return v
}

// Image adds an image block.
func (v *ViewBuilder) Image(blockID, src, alt string) *ViewBuilder {
	v.blocks = append(v.blocks, UIBlock{BlockID: blockID, Type: "image", Src: src, Alt: alt})
	return v
}

// Build returns the constructed View.
func (v *ViewBuilder) Build() *View {
	return &View{
		ViewID: v.viewID,
		Title:  v.title,
		Blocks: v.blocks,
	}
}

// uiBridge adapts a UIHandler to the pb.UIServiceServer gRPC interface.
type uiBridge struct {
	pb.UnimplementedUIServiceServer
	handler     UIHandler
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

func (b *uiBridge) GetView(ctx context.Context, req *pb.GetViewRequest) (*pb.UIView, error) {
	view, err := b.handler.GetView(ctx, req.Context)
	if err != nil {
		return nil, err
	}
	return viewToProto(view), nil
}

func (b *uiBridge) SendEvent(ctx context.Context, req *pb.UIEvent) (*pb.UIEventResponse, error) {
	result, err := b.handler.OnEvent(ctx, UIEvent{
		ViewID:  req.ViewId,
		BlockID: req.BlockId,
		Action:  req.Action,
		Value:   req.Value,
	})
	if err != nil {
		return &pb.UIEventResponse{Error: err.Error()}, nil
	}
	resp := &pb.UIEventResponse{
		Error: result.Error,
		Toast: result.Toast,
	}
	if result.View != nil {
		resp.View = viewToProto(result.View)
	}
	return resp, nil
}

func (b *uiBridge) StreamUpdates(_ *pb.Empty, stream pb.UIService_StreamUpdatesServer) error {
	streamer, ok := b.handler.(UIHandlerWithStreaming)
	if !ok {
		// Not a streaming handler — block until context is done
		<-stream.Context().Done()
		return nil
	}
	ch, err := streamer.StreamUpdates(stream.Context())
	if err != nil {
		return err
	}
	for {
		select {
		case view, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(viewToProto(view)); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (b *uiBridge) Configure(_ context.Context, req *pb.SettingsMap) (*pb.Empty, error) {
	if b.onConfigure != nil {
		b.onConfigure(req.Values)
	}
	return &pb.Empty{}, nil
}

func viewToProto(v *View) *pb.UIView {
	pv := &pb.UIView{
		ViewId: v.ViewID,
		Title:  v.Title,
	}
	for _, blk := range v.Blocks {
		pblk := &pb.UIBlock{
			BlockId:     blk.BlockID,
			Type:        blk.Type,
			Text:        blk.Text,
			Value:       blk.Value,
			Placeholder: blk.Placeholder,
			Hint:        blk.Hint,
			Variant:     blk.Variant,
			Src:         blk.Src,
			Alt:         blk.Alt,
			Disabled:    blk.Disabled,
			Style:       blk.Style,
		}
		for _, opt := range blk.Options {
			pblk.Options = append(pblk.Options, &pb.SelectOption{
				Label: opt.Label,
				Value: opt.Value,
			})
		}
		pv.Blocks = append(pv.Blocks, pblk)
	}
	return pv
}
