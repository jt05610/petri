package middleware

import (
	"log"
	"net/http"
)

func RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("handling middleware RequireUser")

	}
}
