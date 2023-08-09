package sequence

import (
	"context"
	"github.com/jt05610/petri/prisma/db"
)

type Service interface {
	// Load loads the run from the database
	Load(ctx context.Context, id string) (*db.RunModel, error)
	List(ctx context.Context) ([]db.RunModel, error)
}
