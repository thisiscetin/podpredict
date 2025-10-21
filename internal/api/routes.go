package api

import "net/http"

// Routes returns a mux with API routes registered.
func Routes(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/predict", h.Predict)
	mux.HandleFunc("/predictions", h.ListPredictions)
	mux.HandleFunc("/healthz", h.HealthCheck)
	return mux
}
