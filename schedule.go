package nebo

import (
	"context"

	pb "github.com/nebolabs/nebo-sdk-go/pb"
)

// Schedule represents a single scheduled task.
type Schedule = pb.Schedule

// ScheduleTrigger is emitted when a schedule fires.
type ScheduleTrigger = pb.ScheduleTrigger

// ScheduleHistoryEntry represents one execution of a schedule.
type ScheduleHistoryEntry = pb.ScheduleHistoryEntry

// ScheduleHandler is the interface for schedule capability apps.
// Implement this to replace Nebo's built-in cron scheduler.
type ScheduleHandler interface {
	Create(ctx context.Context, req *pb.CreateScheduleRequest) (*pb.Schedule, error)
	Get(ctx context.Context, name string) (*pb.Schedule, error)
	List(ctx context.Context, limit, offset int32, enabledOnly bool) ([]*pb.Schedule, int64, error)
	Update(ctx context.Context, req *pb.UpdateScheduleRequest) (*pb.Schedule, error)
	Delete(ctx context.Context, name string) error
	Enable(ctx context.Context, name string) (*pb.Schedule, error)
	Disable(ctx context.Context, name string) (*pb.Schedule, error)
	Trigger(ctx context.Context, name string) (bool, string, error)
	History(ctx context.Context, name string, limit, offset int32) ([]*pb.ScheduleHistoryEntry, int64, error)
	Triggers(ctx context.Context) (<-chan *pb.ScheduleTrigger, error)
}

// scheduleBridge adapts a ScheduleHandler to the pb.ScheduleServiceServer gRPC interface.
type scheduleBridge struct {
	pb.UnimplementedScheduleServiceServer
	handler     ScheduleHandler
	onConfigure func(map[string]string)
	env         *AppEnv
}

func (b *scheduleBridge) HealthCheck(_ context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Healthy: true,
		Name:    b.env.Name,
		Version: b.env.Version,
	}, nil
}

func (b *scheduleBridge) Create(ctx context.Context, req *pb.CreateScheduleRequest) (*pb.ScheduleResponse, error) {
	sched, err := b.handler.Create(ctx, req)
	if err != nil {
		return &pb.ScheduleResponse{Error: err.Error()}, nil
	}
	return &pb.ScheduleResponse{Schedule: sched}, nil
}

func (b *scheduleBridge) Get(ctx context.Context, req *pb.GetScheduleRequest) (*pb.ScheduleResponse, error) {
	sched, err := b.handler.Get(ctx, req.Name)
	if err != nil {
		return &pb.ScheduleResponse{Error: err.Error()}, nil
	}
	return &pb.ScheduleResponse{Schedule: sched}, nil
}

func (b *scheduleBridge) List(ctx context.Context, req *pb.ListSchedulesRequest) (*pb.ListSchedulesResponse, error) {
	schedules, total, err := b.handler.List(ctx, req.Limit, req.Offset, req.EnabledOnly)
	if err != nil {
		return &pb.ListSchedulesResponse{}, nil
	}
	return &pb.ListSchedulesResponse{Schedules: schedules, Total: total}, nil
}

func (b *scheduleBridge) Update(ctx context.Context, req *pb.UpdateScheduleRequest) (*pb.ScheduleResponse, error) {
	sched, err := b.handler.Update(ctx, req)
	if err != nil {
		return &pb.ScheduleResponse{Error: err.Error()}, nil
	}
	return &pb.ScheduleResponse{Schedule: sched}, nil
}

func (b *scheduleBridge) Delete(ctx context.Context, req *pb.DeleteScheduleRequest) (*pb.DeleteScheduleResponse, error) {
	if err := b.handler.Delete(ctx, req.Name); err != nil {
		return &pb.DeleteScheduleResponse{Error: err.Error()}, nil
	}
	return &pb.DeleteScheduleResponse{Success: true}, nil
}

func (b *scheduleBridge) Enable(ctx context.Context, req *pb.ScheduleNameRequest) (*pb.ScheduleResponse, error) {
	sched, err := b.handler.Enable(ctx, req.Name)
	if err != nil {
		return &pb.ScheduleResponse{Error: err.Error()}, nil
	}
	return &pb.ScheduleResponse{Schedule: sched}, nil
}

func (b *scheduleBridge) Disable(ctx context.Context, req *pb.ScheduleNameRequest) (*pb.ScheduleResponse, error) {
	sched, err := b.handler.Disable(ctx, req.Name)
	if err != nil {
		return &pb.ScheduleResponse{Error: err.Error()}, nil
	}
	return &pb.ScheduleResponse{Schedule: sched}, nil
}

func (b *scheduleBridge) Trigger(ctx context.Context, req *pb.ScheduleNameRequest) (*pb.TriggerResponse, error) {
	success, output, err := b.handler.Trigger(ctx, req.Name)
	if err != nil {
		return &pb.TriggerResponse{Error: err.Error()}, nil
	}
	return &pb.TriggerResponse{Success: success, Output: output}, nil
}

func (b *scheduleBridge) History(ctx context.Context, req *pb.ScheduleHistoryRequest) (*pb.ScheduleHistoryResponse, error) {
	entries, total, err := b.handler.History(ctx, req.Name, req.Limit, req.Offset)
	if err != nil {
		return &pb.ScheduleHistoryResponse{}, nil
	}
	return &pb.ScheduleHistoryResponse{Entries: entries, Total: total}, nil
}

func (b *scheduleBridge) Triggers(_ *pb.Empty, stream pb.ScheduleService_TriggersServer) error {
	ch, err := b.handler.Triggers(stream.Context())
	if err != nil {
		return err
	}
	for {
		select {
		case trigger, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(trigger); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (b *scheduleBridge) Configure(_ context.Context, req *pb.SettingsMap) (*pb.Empty, error) {
	if b.onConfigure != nil {
		b.onConfigure(req.Values)
	}
	return &pb.Empty{}, nil
}
