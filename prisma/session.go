package prisma

import (
	"context"
	"errors"
	"github.com/jt05610/petri/prisma/db"
)

type SessionClient struct {
	*db.PrismaClient
}

func requireUser(ctx context.Context) (string, error) {
	uID := ctx.Value("userID")
	if uID == nil {
		return "", errors.New("user not logged in")
	}
	return uID.(string), nil
}

func (c *SessionClient) ListSessions(ctx context.Context) ([]db.SessionModel, error) {
	return c.Session.FindMany().Exec(ctx)
}

func (c *SessionClient) CreateSession(ctx context.Context, runID string) (*db.SessionModel, error) {
	uID, err := requireUser(ctx)
	if err != nil {
		return nil, err
	}
	return c.Session.CreateOne(
		db.Session.User.Link(db.User.ID.Equals(uID)),
		db.Session.Run.Link(db.Run.ID.Equals(runID)),
	).Exec(ctx)
}
