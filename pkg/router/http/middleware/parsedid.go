package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

type DIDContextKey struct{}

func ParseDID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		didStr := chi.URLParam(r, "identifier")
		if didStr == "" {
			http.Error(w, "identifier is required in the path", http.StatusBadRequest)
			return
		}
		did, err := w3c.ParseDID(didStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				DIDContextKey{},
				did,
			),
		)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
