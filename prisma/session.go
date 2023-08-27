package prisma

import (
	"context"
	"errors"
	"github.com/jt05610/petri/control"
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

func (c *SessionClient) ListSessions(ctx context.Context, runID string) ([]db.SessionModel, error) {
	return c.Session.FindMany(
		db.Session.Run.Link(
			db.Run.ID.Equals(runID),
		),
	).Exec(ctx)
}

func (c *SessionClient) ActiveSessions(ctx context.Context) ([]db.SessionModel, error) {
	return c.Session.FindMany(
		db.Session.State.Equals(db.SessionStateRUNNING),
	).Exec(ctx)
}

func (c *SessionClient) CreateSession(ctx context.Context, runID, userID string, instances []string) (*db.SessionModel, error) {
	session, err := c.Session.CreateOne(
		db.Session.User.Link(db.User.ID.Equals(userID)),
		db.Session.Run.Link(db.Run.ID.Equals(runID)),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	for _, instance := range instances {
		_, err := c.Instance.FindUnique(
			db.Instance.ID.Equals(instance),
		).Update(
			db.Instance.Sessions.Link(
				db.Session.ID.Equals(session.ID),
			),
		).Exec(ctx)

		if err != nil {
			return nil, err
		}
	}
	return session, nil
}

func (c *SessionClient) AddData(ctx context.Context, sessionID string, event *control.Event) (*db.DatumModel, error) {
	return c.Datum.CreateOne(
		db.Datum.Instance.Link(
			db.Instance.ID.Equals(event.From),
		),
		db.Datum.Session.Link(
			db.Session.ID.Equals(sessionID),
		),
		db.Datum.Event.Link(
			db.Event.ID.Equals(event.Event.ID),
		),
	).Exec(ctx)
}

func (c *SessionClient) StartSession(ctx context.Context, sessionID string, timestamp db.DateTime) (*db.SessionModel, error) {
	return c.Session.UpsertOne(
		db.Session.ID.Equals(sessionID),
	).Update(
		db.Session.State.Set(db.SessionStateRUNNING),
		db.Session.StartedAt.Set(timestamp),
	).Exec(ctx)
}

func (c *SessionClient) StopSession(ctx context.Context, sessionID string, timestamp db.DateTime) (*db.SessionModel, error) {
	return c.Session.UpsertOne(
		db.Session.ID.Equals(sessionID),
	).Update(
		db.Session.State.Set(db.SessionStateSTOPPED),
		db.Session.StoppedAt.Set(timestamp),
	).Exec(ctx)
}

func (c *SessionClient) SessionData(ctx context.Context, sessionID string) ([]db.DatumModel, error) {
	return c.Datum.FindMany(
		db.Datum.Session.Link(
			db.Session.ID.Equals(sessionID),
		),
	).Exec(ctx)
}

func (c *SessionClient) PauseSession(ctx context.Context, sessionID string, timestamp db.DateTime) (*db.SessionModel, error) {
	return c.Session.UpsertOne(
		db.Session.ID.Equals(sessionID),
	).Update(
		db.Session.State.Set(db.SessionStatePAUSED),
		db.Session.PausedAt.Push([]db.DateTime{timestamp}),
	).Exec(ctx)
}

func (c *SessionClient) ResumeSession(ctx context.Context, sessionID string, timestamp db.DateTime) (*db.SessionModel, error) {
	return c.Session.UpsertOne(
		db.Session.ID.Equals(sessionID),
	).Update(
		db.Session.State.Set(db.SessionStateRUNNING),
		db.Session.ResumedAt.Push([]db.DateTime{timestamp}),
	).Exec(ctx)
}
