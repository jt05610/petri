package handlers

import (
	"io"
	"net/http"
	"sort"
	"strings"
)

type Methods map[string]http.Handler

func (m Methods) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(r io.ReadCloser) {
		_, _ = io.Copy(io.Discard, r)
		_ = r.Close()
	}(r.Body)

	if h, ok := m[r.Method]; ok {
		if h == nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		} else {
			h.ServeHTTP(w, r)
		}
		return
	}
	w.Header().Add("Allow", m.allowedMethods())
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (m Methods) allowedMethods() string {
	var methods []string
	for method := range m {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	return strings.Join(methods, ", ")
}

func NetSVG() Methods {
	return Methods{
		http.MethodGet: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Not Found", http.StatusNotFound)
		}),
	}
}

func MarkedNetSVG() Methods {
	return Methods{
		http.MethodGet: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Not Found", http.StatusNotFound)
		}),
	}
}
