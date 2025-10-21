package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/thisiscetin/podpredict/internal/api"
	"github.com/thisiscetin/podpredict/internal/config"
	"github.com/thisiscetin/podpredict/internal/fetcher/gsheets"
	"github.com/thisiscetin/podpredict/internal/metrics"
	"github.com/thisiscetin/podpredict/internal/model"
	"github.com/thisiscetin/podpredict/internal/model/linreg"
	"github.com/thisiscetin/podpredict/internal/store"
	"github.com/thisiscetin/podpredict/internal/store/inmemory"
)

func main() {
	// Context cancels on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load config once.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config error: ", err)
	}

	// Deps
	mdl := linreg.NewModel()
	st := inmemory.NewStore()

	// Fetcher
	ftc, err := gsheets.NewFetcher(ctx, cfg.CredsJSON, cfg.SpreadsheetID)
	if err != nil {
		log.Fatal("fetcher init error: ", err)
	}

	// Fetch → Train
	mtr, err := ftc.Fetch()
	if err != nil {
		log.Fatal("fetch error: ", err)
	}
	if err := mdl.Train(filterDaysWithPods(mtr)); err != nil {
		log.Fatal("train error: ", err)
	}

	// Predict missing pods → Store
	if err := upsertPredictions(ctx, mdl, st, mtr); err != nil {
		log.Fatal("prediction storage error: ", err)
	}

	// API & Server
	h, err := api.New(mdl, ftc, st, cfg.FetchTimeout)
	if err != nil {
		log.Fatal("api init failed: ", err)
	}
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.Routes(h),
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	// Wait for signal or server error, then graceful shutdown.
	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil {
			log.Printf("server error: %v", err)
		}
	}

	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shCtx)
}

func filterDaysWithPods(ms []metrics.Daily) []metrics.Daily {
	out := make([]metrics.Daily, 0, len(ms))
	for _, m := range ms {
		if m.HasPods() {
			out = append(out, m)
		}
	}
	return out
}

func upsertPredictions(ctx context.Context, mdl model.Model, st store.Store, ms []metrics.Daily) error {
	for _, m := range ms {
		fePods, bePods := model.FEPods(0), model.BEPods(0)
		features := model.FeaturesFromDaily(m)

		if !m.HasPods() {
			fp, bp, err := mdl.Predict(&features)
			if err != nil {
				return err
			}
			fePods, bePods = fp, bp
		} else {
			fePods, bePods = model.FEPods(ptrVal(m.FEPods)), model.BEPods(ptrVal(m.BEPods))
		}

		if err := st.Append(ctx, store.Prediction{
			ID:        uuid.New().String(),
			Timestamp: m.Date,
			Input:     features,
			FEPods:    int(fePods),
			BEPods:    int(bePods),
		}); err != nil {
			return err
		}
	}
	return nil
}

func ptrVal(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
