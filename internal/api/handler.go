package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/thisiscetin/podpredict/internal/fetcher"
	"github.com/thisiscetin/podpredict/internal/model"
	"github.com/thisiscetin/podpredict/internal/store"
)

// Handler owns HTTP endpoints and their dependencies.
type Handler struct {
	model   model.Model
	fetcher fetcher.Fetcher
	store   store.Store

	// optional: request-scoped timeout
	timeout time.Duration
}

// New wires dependencies, fetches training data via Fetcher, and trains the Model.
func New(m model.Model, f fetcher.Fetcher, st store.Store, timeout time.Duration) (*Handler, error) {
	if m == nil {
		return nil, errors.New("nil model")
	}
	if f == nil {
		return nil, errors.New("nil fetcher")
	}
	if st == nil {
		return nil, errors.New("nil store")
	}

	// Initial training
	data, err := f.Fetch()
	if err != nil {
		return nil, err
	}
	if err := m.Train(data); err != nil {
		return nil, err
	}

	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Handler{
		model:   m,
		fetcher: f,
		store:   st,
		timeout: timeout,
	}, nil
}

// POST /predict
// Body: { "gmv": <float>, "users": <float>, "marketing_cost": <float> }
// Returns: store.Prediction (with timestamp)
func (h *Handler) Predict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	var in model.Features
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	fe, be, err := h.model.Predict(&in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "prediction failed: "+err.Error())
		return
	}

	rec := store.Prediction{
		Timestamp: time.Now(),
		Input:     in,
		FEPods:    int(fe),
		BEPods:    int(be),
	}
	if err := h.store.Append(ctx, rec); err != nil {
		writeError(w, http.StatusInternalServerError, "persisting prediction failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, rec)
}

// GET /predictions
// Returns: []store.Prediction
func (h *Handler) ListPredictions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	items, err := h.store.List(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing predictions failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// GET /healthz
// Returns JSON with status info about model and store
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := struct {
		Status  string `json:"status"`
		StoreOK bool   `json:"store_ok"`
		ModelOK bool   `json:"model_ok"`
		Now     string `json:"timestamp"`
	}{
		Status:  "ok",
		ModelOK: true,
		Now:     time.Now().UTC().Format(time.RFC3339),
	}

	// Check store health
	if _, err := h.store.List(ctx); err != nil {
		status.StoreOK = false
		status.Status = "degraded"
	} else {
		status.StoreOK = true
	}

	writeJSON(w, http.StatusOK, status)
}

// Optional: expose retraining for future endpoints/CLI
func (h *Handler) Retrain(ctx context.Context) error {
	data, err := h.fetcher.Fetch()
	if err != nil {
		return err
	}
	return h.model.Train(data)
}
