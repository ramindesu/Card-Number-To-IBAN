package handler

import "github.com/go-chi/chi/v5"

// RegisterRoutes registers the endpoints owned by the Card-to-IBAN slice.
func (h *CardToIBANHandler) RegisterRoutes(router chi.Router) {
	router.Post("/card-to-iban", h.Convert)
}
