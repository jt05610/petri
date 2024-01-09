package couch

import (
	"context"
	_ "github.com/go-kivik/couchdb/v3"
	"github.com/go-kivik/kivik/v3"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri"
	"os"
)

var _ petri.Service[petri.Object, petri.Input, petri.Filter, petri.Update] = (*Service[petri.Object, petri.Input, petri.Filter, petri.Update])(nil)

type Service[T petri.Object, U petri.Input, V petri.Filter, W petri.Update] struct {
	cancel func()
	db     *kivik.DB
	revMap map[string]string
}

type Config struct {
	User    string
	Pass    string
	Address string
	Port    string
}

func (c *Config) URI() string {
	return "http://" + c.User + ":" + c.Pass + "@" + c.Address + ":" + c.Port
}

func lookupKey(key string, into *string) {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic("missing env var: " + key)
	}
	*into = value
}

func LoadConfig(envFile ...string) *Config {
	var config Config
	err := godotenv.Load(envFile...)
	if err != nil {
		panic(err)
	}
	keys := []struct {
		key  string
		into *string
	}{
		{"COUCHDB_USER", &config.User},
		{"COUCHDB_PASSWORD", &config.Pass},
		{"COUCHDB_HOST", &config.Address},
		{"COUCHDB_PORT", &config.Port},
	}

	for _, k := range keys {
		lookupKey(k.key, k.into)
	}
	return &config
}

func Open[T petri.Object, U petri.Input, V petri.Filter, W petri.Update](uri string, name string) (*Service[T, U, V, W], error) {
	client, err := kivik.New("couch", uri)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	dbs, err := client.AllDBs(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	found := false
	for _, db := range dbs {
		if db == name {
			found = true
			break
		}
	}
	if !found {
		err = client.CreateDB(ctx, name)
		if err != nil {
			cancel()
			return nil, err
		}
	}
	db := client.DB(ctx, name)
	if err != nil {
		cancel()
		return nil, err
	}
	return &Service[T, U, V, W]{
		cancel: cancel,
		db:     db,
		revMap: make(map[string]string),
	}, nil
}

func (s *Service[T, U, V, W]) Close() error {
	s.cancel()
	return nil
}

func (s *Service[T, U, V, W]) Update(ctx context.Context, id string, update W) (T, error) {
	o, err := s.Get(context.Background(), id)
	if err != nil {
		var zero T
		return zero, err
	}
	err = o.Update(update)
	if err != nil {
		var zero T
		return zero, err
	}
	doc := o.Document()
	doc["_rev"] = s.revMap[id]
	rev, err := s.db.Put(ctx, id, doc)
	if err != nil {
		var zero T
		return zero, err
	}
	s.revMap[id] = rev
	return o, nil
}

func (s *Service[T, U, V, W]) Get(ctx context.Context, id string) (T, error) {
	var ret T
	var zero T
	row := s.db.Get(ctx, id)
	err := row.ScanDoc(&ret)
	if err != nil {
		return zero, err
	}
	s.revMap[id] = row.Rev
	return ret, nil
}

func (s *Service[T, U, V, W]) List(ctx context.Context, f V) ([]T, error) {
	ret := make([]T, 0)
	rows, err := s.db.Find(ctx, map[string]interface{}{
		"selector": f,
	}, kivik.Options{})
	if err != nil {
		return ret, err
	}
	for rows.Next() {
		var row T
		err := rows.ScanDoc(&row)
		if err != nil {
			return ret, err
		}
		ret = append(ret, row)
	}
	return ret, nil
}

func (s *Service[T, U, V, W]) Add(ctx context.Context, input U) (T, error) {
	var zero T
	o, ok := input.Object().(T)
	if !ok {
		return zero, petri.ErrWrongInput
	}
	doc := o.Document()
	doc["_id"] = o.Identifier()
	rev, err := s.db.Put(ctx, o.Identifier(), doc)
	if err != nil {
		return zero, err
	}
	s.revMap[o.Identifier()] = rev
	return o, nil
}

func (s *Service[T, U, V, W]) Remove(ctx context.Context, id string) (T, error) {
	var zero T
	o, err := s.Get(ctx, id)
	if err != nil {
		return zero, err
	}
	rev, err := s.db.Delete(ctx, id, s.revMap[id])
	if err != nil {
		return zero, err
	}
	s.revMap[id] = rev
	return o, nil
}
func TokenService(uri string) (petri.Service[*petri.TokenSchema, *petri.TokenInput, *petri.TokenFilter, *petri.TokenUpdate], error) {
	return Open[*petri.TokenSchema, *petri.TokenInput, *petri.TokenFilter, *petri.TokenUpdate](uri, "tokens")
}

func PlaceService(uri string) (*Service[*petri.Place, *petri.PlaceInput, *petri.PlaceFilter, *petri.PlaceUpdate], error) {
	return Open[*petri.Place, *petri.PlaceInput, *petri.PlaceFilter, *petri.PlaceUpdate](uri, "places")
}

func TransitionService(uri string) (*Service[*petri.Transition, *petri.TransitionInput, *petri.TransitionFilter, *petri.TransitionUpdate], error) {
	return Open[*petri.Transition, *petri.TransitionInput, *petri.TransitionFilter, *petri.TransitionUpdate](uri, "transitions")
}

func ArcService(uri string) (*Service[*petri.Arc, *petri.ArcInput, *petri.ArcFilter, *petri.ArcUpdate], error) {
	return Open[*petri.Arc, *petri.ArcInput, *petri.ArcFilter, *petri.ArcUpdate](uri, "arcs")
}

func NetService(uri string) (*Service[*petri.Net, *petri.NetInput, *petri.NetFilter, *petri.NetUpdate], error) {
	return Open[*petri.Net, *petri.NetInput, *petri.NetFilter, *petri.NetUpdate](uri, "nets")
}
