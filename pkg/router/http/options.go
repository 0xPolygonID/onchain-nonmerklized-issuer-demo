package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type Option func(r *chi.Mux)

func WithOrigins(origins []string) Option {
	return func(r *chi.Mux) {
		c := cors.New(cors.Options{
			AllowedOrigins: origins,
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type",
				"X-CSRF-Token"},
			AllowCredentials: true,
		})
		r.Use(c.Handler)
	}
}

// nolint: gocritic // no sense to pass cors.Options by pointer, since we use value after
func WithCors(corsOption cors.Options) Option {
	return func(r *chi.Mux) {
		c := cors.New(corsOption)
		r.Use(c.Handler)
	}
}
