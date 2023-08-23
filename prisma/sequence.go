package prisma

import (
	"context"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma/db"
	"github.com/jt05610/petri/sequence"
)

var _ sequence.Service = (*RunClient)(nil)

type RunClient struct {
	*db.PrismaClient
}

func (c *RunClient) Load(ctx context.Context, id string) (*sequence.Sequence, error) {
	r, err := c.load(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToSequence(r), nil
}

func (c *RunClient) List(ctx context.Context) ([]*sequence.ListItem, error) {
	runs, err := c.list(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*sequence.ListItem, len(runs))
	for i, r := range runs {
		list[i] = &sequence.ListItem{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
		}
	}
	return list, nil
}

func (c *RunClient) load(ctx context.Context, id string) (*db.RunModel, error) {
	return c.Run.FindUnique(
		db.Run.ID.Equals(id),
	).With(
		db.Run.Net.Fetch().With(
			db.Net.Places.Fetch(),
		).With(
			db.Net.Transitions.Fetch(),
		).With(
			db.Net.Arcs.Fetch(),
		).With(
			db.Net.Children.Fetch(),
		),
	).With(
		db.Run.Steps.Fetch().With(
			db.Step.Action.Fetch().With(
				db.Action.Constants.Fetch(),
			).With(
				db.Action.Device.Fetch().With(
					db.Device.Nets.Fetch(),
				).With(
					db.Device.Actions.Fetch().With(),
				).With(
					db.Device.Instances.Fetch(),
				),
			).With(
				db.Action.Event.Fetch().With(
					db.Event.Fields.Fetch(),
				).With(
					db.Event.Transitions.Fetch(),
				),
			),
		).OrderBy(
			db.Step.Order.Order(db.SortOrderAsc),
		),
	).Exec(ctx)
}

func (c *RunClient) list(ctx context.Context) ([]db.RunModel, error) {
	return c.Run.FindMany().Exec(ctx)
}

func ToEvent(e *db.EventModel) *labeled.Event {
	fields := make([]*labeled.Field, len(e.Fields()))
	for i, f := range e.Fields() {
		fields[i] = ToField(&f)
	}
	return &labeled.Event{
		Name:   e.Name,
		Fields: fields,
	}
}

func ToInstance(i *db.InstanceModel) *device.Instance {
	return &device.Instance{
		ID:       i.ID,
		Name:     i.Name,
		Language: string(i.Language),
		Address:  i.Addr,
	}
}

func ToInstances(instances []db.InstanceModel) []*device.Instance {
	ret := make([]*device.Instance, len(instances))
	for i, inst := range instances {
		ret[i] = ToInstance(&inst)
	}
	return ret
}

func ToDevice(d *db.DeviceModel) *device.Device {
	return &device.Device{
		ID:        d.ID,
		Name:      d.Name,
		Instances: ToInstances(d.Instances()),
	}
}

func ToField(f *db.FieldModel) *labeled.Field {
	return &labeled.Field{
		Name: f.Name,
		Type: labeled.FieldType(f.Type),
	}
}

func ToAction(a *db.ActionModel) *sequence.Action {
	constants := make([]*sequence.Constant, len(a.Constants()))
	for i, c := range a.Constants() {
		constants[i] = &sequence.Constant{
			Field: ToField(c.Field()),
			Value: c.Value,
		}
	}
	return &sequence.Action{
		Constants: constants,
		Device:    ToDevice(a.Device()),
		Event:     ToEvent(a.Event()),
	}
}

func ToStep(s *db.StepModel) *sequence.Step {
	return &sequence.Step{
		Action: ToAction(s.Action()),
	}
}

func ToSequence(r *db.RunModel) *sequence.Sequence {
	steps := make([]*sequence.Step, len(r.Steps()))
	for i, step := range r.Steps() {
		steps[i] = ToStep(&step)
	}
	return &sequence.Sequence{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Steps:       steps,
	}
}
