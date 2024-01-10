package petri

type Selector[T any] struct {
	Equals              T   `json:"$eq,omitempty"`
	GreaterThan         T   `json:"$gt,omitempty"`
	GreaterThanOrEquals T   `json:"$gte,omitempty"`
	LessThan            T   `json:"$lt,omitempty"`
	LessThanOrEquals    T   `json:"$lte,omitempty"`
	In                  []T `json:"$in,omitempty"`
}

type StringSelector Selector[string]
type IntSelector Selector[int64]
type FloatSelector Selector[float64]
type BooleanSelector Selector[bool]
