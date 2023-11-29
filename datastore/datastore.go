package datastore

import "context"

type Initializer interface {
	CreateDb(ctx context.Context, name string) error
}

type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

type User interface {
	CreateUser(ctx context.Context, name, password string, role Role) error
	DeleteUser(ctx context.Context, name string) error
}

type Administrator interface {
	DeleteDb(ctx context.Context, name string) error
	AddUserToDb(ctx context.Context, db string, user User) error
	RemoveUserFromDb(ctx context.Context, db, user string) error
	ChangeUserDbRole(ctx context.Context, db, user, role string) error
}

type Loader interface {
	Load(ctx context.Context, name string) (DataStore, error)
}

// DataStore is an interface for storing event data and device configuration data.
type DataStore interface {
}
