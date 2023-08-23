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
