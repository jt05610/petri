package prisma

import (
	"context"
	"github.com/jt05610/petri/prisma/db"
	"github.com/jt05610/petri/sequence"
)

var _ sequence.Service = (*RunClient)(nil)

type RunClient struct {
	*db.PrismaClient
}

func (c *RunClient) Load(ctx context.Context, id string) (*db.RunModel, error) {
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

func (c *RunClient) List(ctx context.Context) ([]db.RunModel, error) {
	return c.Run.FindMany().With().Exec(ctx)
}
