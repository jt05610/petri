package petri

import "context"

func (d Document) With(key string, value interface{}) Document {
	d[key] = value
	return d
}

type Input interface {
	IsInput()
	Kind() Kind
	Object() Object
}

type Mask interface {
	IsMask()
}

type Update interface {
	IsUpdate()
}

type Filter interface {
	IsFilter()
}

type Adder[T Object, U Input] interface {
	Add(ctx context.Context, input U) (T, error)
}

type Remover[T Object] interface {
	Remove(ctx context.Context, id string) (T, error)
}

type Updater[T Object, U Update] interface {
	Update(ctx context.Context, id string, update U) (T, error)
}

type Getter[T Object] interface {
	Get(ctx context.Context, id string) (T, error)
}

type Lister[T Object, U Filter] interface {
	List(ctx context.Context, f U) ([]T, error)
}

type Service[T Object, U Input, V Filter, W Update] interface {
	Adder[T, U]
	Updater[T, W]
	Getter[T]
	Lister[T, V]
	Remover[T]
}
