package model

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
)

type JSON map[string]interface{}

func MarshalJSON(v JSON) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(v)
		if err != nil {
			panic(err)
		}
	})
}

func unmarshalJSON(b []byte) (JSON, error) {
	var v JSON
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func UnmarshalJSON(v interface{}) (JSON, error) {
	switch v := v.(type) {
	case string:
		return unmarshalJSON([]byte(v))
	case []byte:
		return unmarshalJSON(v)
	case map[string]interface{}:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported type for JSON: %T", v)
	}
}
