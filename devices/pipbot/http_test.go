package PipBot

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func drainAndClose(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		})
}

func TestMethods_ServeHTTP(t *testing.T) {
	d := NewPipBot(Autosampler, []int{3, 4}, 0, nil)
	serveMux := d.Mux()
	mux := drainAndClose(serveMux)

	cases := []struct {
		path     string
		method   string
		data     interface{}
		response string
		code     int
	}{
		{"http://test/volume", http.MethodGet, nil, "[]\n", http.StatusOK},
		{"http://test/volume", http.MethodPost, VolumeRequest{
			"2": {
				"A1": 1000,
			},
		}, "[]\n", http.StatusOK},
	}

	for i, c := range cases {
		buf := new(bytes.Buffer)
		err := EncodeJSON(buf, c.data)
		if err != nil {
			panic(err)
		}
		r := httptest.NewRequest(c.method, c.path, buf)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		resp := w.Result()

		if resp.StatusCode != c.code {
			t.Errorf("%d: expected %d, got %d", i, c.code, resp.StatusCode)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		_ = resp.Body.Close()

		if string(b) != c.response {
			t.Errorf("%d: expected %s, got %s", i, c.response, string(b))
		}
	}
}
