package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/iden3/go-service-template/pkg/router/http/handlers"
	"github.com/iden3/go-service-template/pkg/router/http/middleware"
)

type Handlers struct {
	systemHandler         handlers.SystemHandler
	authenticationHandler handlers.AuthenticationHandlers
	issuerHandler         handlers.IssuerHandlers
}

func NewHandlers(
	systemHandler handlers.SystemHandler,
	authHendler handlers.AuthenticationHandlers,
	issuerHandler handlers.IssuerHandlers,
) Handlers {
	return Handlers{
		systemHandler:         systemHandler,
		authenticationHandler: authHendler,
		issuerHandler:         issuerHandler,
	}
}

func (h *Handlers) NewRouter(opts ...Option) http.Handler {
	r := chi.NewRouter()

	for _, opt := range opts {
		opt(r)
	}

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestLog)
	r.Use(chimiddleware.Recoverer)

	h.basicRouters(r)
	h.authRouters(r)
	h.apiRouters(r)

	return r
}

func (h Handlers) basicRouters(r *chi.Mux) {
	r.Get("/readiness", h.systemHandler.Readiness)
	r.Get("/liveness", h.systemHandler.Liveness)
}

func (h Handlers) authRouters(r *chi.Mux) {
	r.Get("/api/v1/requests/auth", h.authenticationHandler.CreateAuthenticationRequest)
	r.Post("/api/v1/callback", h.authenticationHandler.Callback)
	r.Get("/api/v1/status", h.authenticationHandler.AuthenticationRequestStatus)
}

func (h Handlers) apiRouters(r *chi.Mux) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/issuers", h.issuerHandler.GetIssuersList)
	})
}
