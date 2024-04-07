package petri

import (
	"bytes"
	"context"
	"github.com/jt05610/petri"
	"io"
	"sync"
)

type TokenService interface {
	PutToken(ctx context.Context, placeId string, tok petri.Token) error
	PopToken(ctx context.Context, placeId string) (petri.Token, error)
	GetMarking(ctx context.Context, withValue bool) (petri.Marking, error)
	GetToken(ctx context.Context, id string) (petri.Token, error)
	Monitor(ctx context.Context, in <-chan RawMarking) (<-chan RawMarking, error)
}

type RawToken struct {
	ID    string
	Value []byte
}

type RawMarking map[string][]RawToken

func (r RawMarking) Marking() petri.Marking {
	m := make(petri.Marking)
	for k, v := range r {
		for _, tok := range v {
			m[k] = append(m[k], tok.Token(true))
		}
	}
	return m
}

func MakeRawMarking(m petri.Marking) RawMarking {
	r := make(RawMarking)
	for k, v := range m {
		for _, tok := range v {
			bb, err := io.ReadAll(tok.Value)
			if err != nil {
				continue
			}
			r[k] = append(r[k], RawToken{
				ID:    tok.ID,
				Value: bb,
			})
		}
	}
	return r
}

func (r RawToken) Bytes() []byte {
	return bytes.Clone(r.Value)
}

func (r RawToken) Token(withValue bool) petri.Token {
	if !withValue {
		return petri.Token{
			ID: r.ID,
		}
	}
	return petri.Token{
		ID:    r.ID,
		Value: bytes.NewBuffer(r.Bytes()),
	}
}

type LocalTokenService struct {
	tokens *sync.Map
	updCh  chan struct{}
}

func DefaultTokenService() *LocalTokenService {
	return &LocalTokenService{
		tokens: new(sync.Map),
	}
}

func MakeRawToken(tok petri.Token) RawToken {
	bb, err := io.ReadAll(tok.Value)
	if err != nil {
		return RawToken{}
	}
	return RawToken{
		ID:    tok.ID,
		Value: bb,
	}
}

func (l *LocalTokenService) PutToken(ctx context.Context, placeId string, tok petri.Token) error {
	v, ok := l.tokens.Load(placeId)
	if !ok {
		v = []RawToken{}
	}
	v = append(v.([]RawToken), MakeRawToken(tok))
	l.tokens.Store(placeId, v)
	select {
	case l.updCh <- struct{}{}:
	default:
	}
	return nil
}

func (l *LocalTokenService) PopToken(ctx context.Context, placeId string) (petri.Token, error) {
	v, ok := l.tokens.Load(placeId)
	if !ok {
		return petri.Token{}, nil
	}
	tokens := v.([]petri.Token)
	if len(tokens) == 0 {
		return petri.Token{}, nil
	}
	tok := tokens[0]
	l.tokens.Store(placeId, tokens[1:])
	select {
	case l.updCh <- struct{}{}:
	default:
	}
	return tok, nil
}

func (l *LocalTokenService) RawMarking() map[string][]RawToken {
	m := make(map[string][]RawToken)
	l.tokens.Range(func(key, value interface{}) bool {
		for _, tok := range value.([]RawToken) {
			m[key.(string)] = append(m[key.(string)], tok)
		}
		return true
	})
	return m
}

func (l *LocalTokenService) GetMarking(_ context.Context, withValue bool) (petri.Marking, error) {
	m := make(petri.Marking)
	l.tokens.Range(func(key, value interface{}) bool {
		for _, t := range value.([]RawToken) {
			tok := t.Token(withValue)
			m[key.(string)] = append(m[key.(string)], tok)
		}
		return true
	})
	return m, nil
}

func (l *LocalTokenService) GetToken(ctx context.Context, id string) (petri.Token, error) {
	v, ok := l.tokens.Load(id)
	if !ok {
		return petri.Token{}, nil
	}
	tok := v.(RawToken).Token(true)
	return tok, nil
}

func markingIDs(m petri.Marking) map[string][]string {
	ids := make(map[string][]string)
	for k, v := range m {
		for _, tok := range v {
			ids[k] = append(ids[k], tok.ID)
		}
	}
	return ids
}

func rawIDs(m map[string][]RawToken) map[string][]string {
	ids := make(map[string][]string)
	for k, v := range m {
		for _, tok := range v {
			ids[k] = append(ids[k], tok.ID)
		}
	}
	return ids
}

func (l *LocalTokenService) SyncMarking(m map[string][]RawToken) {
	for k, v := range rawIDs(m) {
		newLen := len(v)
		old, found := l.tokens.Swap(k, v)
		if !found {
			continue
		}
		ot := old.([]string)
		if len(ot) == newLen {
			continue
		}
		l.updCh <- struct{}{}
	}
}

func (l *LocalTokenService) Monitor(ctx context.Context, in <-chan RawMarking) (<-chan RawMarking, error) {
	l.updCh = make(chan struct{})
	ch := make(chan RawMarking)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case <-l.updCh:
				m := l.RawMarking()
				ch <- m

			case m := <-in:
				l.SyncMarking(m)
			}
		}
	}()
	return ch, nil
}

var _ TokenService = (*LocalTokenService)(nil)
