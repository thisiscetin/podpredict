package api

import "net/http"

// Routes returns a mux with API routes registered.
func Routes(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /predict", h.Predict)
	mux.HandleFunc("GET /predictions", h.ListPredictions)
	mux.HandleFunc("GET /healthz", h.HealthCheck)
	return mux
}
