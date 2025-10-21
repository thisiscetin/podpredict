package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thisiscetin/podpredict/internal/metrics"
	"github.com/thisiscetin/podpredict/internal/model"
	"github.com/thisiscetin/podpredict/internal/store"
)

type mockModel struct {
	trainedWith []metrics.Daily
	fe          model.FEPods
	be          model.BEPods
	err         error
}

func (m *mockModel) Train(ds []metrics.Daily) error {
	m.trainedWith = ds
	return nil
}
func (m *mockModel) Predict(_ *model.Features) (model.FEPods, model.BEPods, error) {
	return m.fe, m.be, m.err
}

type mockFetcher struct {
	out []metrics.Daily
	err error
}

func (f mockFetcher) Fetch() ([]metrics.Daily, error) { return f.out, f.err }

type mockStore struct {
	items []store.Prediction
	aerr  error
	lerr  error
}

func (s *mockStore) Append(_ context.Context, r store.Prediction) error {
	if s.aerr != nil {
		return s.aerr
	}
	s.items = append(s.items, r)
	return nil
}
func (s *mockStore) List(_ context.Context) ([]store.Prediction, error) {
	if s.lerr != nil {
		return nil, s.lerr
	}
	out := make([]store.Prediction, len(s.items))
	copy(out, s.items)
	return out, nil
}

func TestNew_TrainsModel(t *testing.T) {
	mm := &mockModel{}
	ff := mockFetcher{out: []metrics.Daily{}}
	ss := &mockStore{}

	h, err := New(mm, ff, ss, 2*time.Second)
	require.NoError(t, err)
	require.NotNil(t, h)
	assert.NotNil(t, mm.trainedWith)
}

func TestPredict_Success(t *testing.T) {
	mm := &mockModel{fe: 7, be: 3}
	ff := mockFetcher{out: []metrics.Daily{}}
	ss := &mockStore{}

	h, err := New(mm, ff, ss, time.Second)
	require.NoError(t, err)

	srv := httptest.NewServer(Routes(h))
	defer srv.Close()

	body := []byte(`{"gmv":1000,"users":50,"marketing_cost":12}`)
	resp, err := http.Post(srv.URL+"/predict", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var got store.Prediction
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, 7, got.FEPods)
	assert.Equal(t, 3, got.BEPods)
	assert.Equal(t, 1000.0, got.Input.GMV)
	assert.Equal(t, 50.0, got.Input.Users)
	assert.Equal(t, 12.0, got.Input.MarketingCost)

	// Ensure it was stored
	items, err := ss.List(context.Background())
	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestPredict_BadJSON(t *testing.T) {
	mm := &mockModel{}
	ff := mockFetcher{out: []metrics.Daily{}}
	ss := &mockStore{}

	h, _ := New(mm, ff, ss, time.Second)
	srv := httptest.NewServer(Routes(h))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/predict", "application/json", bytes.NewReader([]byte(`{`)))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPredict_ModelError(t *testing.T) {
	mm := &mockModel{err: errors.New("boom")}
	ff := mockFetcher{out: []metrics.Daily{}}
	ss := &mockStore{}

	h, _ := New(mm, ff, ss, time.Second)
	srv := httptest.NewServer(Routes(h))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/predict", "application/json",
		bytes.NewReader([]byte(`{"gmv":1,"users":1,"marketing_cost":1}`)))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestPredict_StoreError(t *testing.T) {
	mm := &mockModel{fe: 2, be: 1}
	ff := mockFetcher{out: []metrics.Daily{}}
	ss := &mockStore{aerr: errors.New("db down")}

	h, _ := New(mm, ff, ss, time.Second)
	srv := httptest.NewServer(Routes(h))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/predict", "application/json",
		bytes.NewReader([]byte(`{"gmv":1,"users":1,"marketing_cost":1}`)))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestListPredictions_Success(t *testing.T) {
	mm := &mockModel{fe: 5, be: 4}
	ff := mockFetcher{out: []metrics.Daily{}}
	ss := &mockStore{}

	h, _ := New(mm, ff, ss, time.Second)
	mux := Routes(h)

	// seed
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/predict",
		bytes.NewReader([]byte(`{"gmv":10,"users":2,"marketing_cost":0}`)))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// list
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/predictions", nil)
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var items []store.Prediction
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&items))
	require.Len(t, items, 1)
	assert.Equal(t, 5, items[0].FEPods)
	assert.Equal(t, 4, items[0].BEPods)
	assert.Equal(t, 10.0, items[0].Input.GMV)
}
