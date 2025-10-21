package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/thisiscetin/podpredict/internal/api"
	"github.com/thisiscetin/podpredict/internal/fetcher"
	"github.com/thisiscetin/podpredict/internal/fetcher/mock"
	"github.com/thisiscetin/podpredict/internal/metrics"
	"github.com/thisiscetin/podpredict/internal/model"
	"github.com/thisiscetin/podpredict/internal/model/linreg"
	"github.com/thisiscetin/podpredict/internal/store"
	"github.com/thisiscetin/podpredict/internal/store/inmemory"
)

func makeDay(date time.Time, gmv float64, users int, mc float64, fe, be int) metrics.Daily {
	d, err := metrics.NewDaily(date, gmv, users, mc, &fe, &be)
	if err != nil {
		panic(err)
	}
	return d
}

var base = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
var rows = []metrics.Daily{
	makeDay(base.AddDate(0, 0, 0), 100, 10, 5, 5, 2),
	makeDay(base.AddDate(0, 0, 1), 150, 12, 6, 6, 3),
	makeDay(base.AddDate(0, 0, 2), 200, 20, 8, 7, 4),
	makeDay(base.AddDate(0, 0, 3), 250, 25, 9, 8, 5),
	makeDay(base.AddDate(0, 0, 4), 300, 30, 10, 9, 6),
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mf := &mock.MockFetcher{}
	var last = makeDay(base.AddDate(0, 0, 5), 5, 10, 10, 0, 0)
	last.FEPods = nil
	last.BEPods = nil
	rows = append(rows, last)

	mf.On("Fetch").Return(rows, nil)

	var mdl model.Model = linreg.NewModel()
	var f fetcher.Fetcher = mf
	var st store.Store = inmemory.NewStore()

	mtr, err := f.Fetch()
	if err != nil {
		log.Fatal("fetch error", err)
	}

	trData := make([]metrics.Daily, 0)
	for _, m := range mtr {
		if m.HasPods() {
			trData = append(trData, m)
		}
	}
	if err := mdl.Train(trData); err != nil {
		log.Fatal("train error", err)
	}

	for _, m := range mtr {
		fePods, bePods := model.FEPods(0), model.BEPods(0)
		features := model.FeaturesFromDaily(m)

		if !m.HasPods() {
			fp, bp, err := mdl.Predict(&features)
			if err != nil {
				log.Fatal("prediction error", err)
			}
			fePods = fp
			bePods = bp
		} else {
			fePods, bePods = model.FEPods(*m.FEPods), model.BEPods(*m.BEPods)
		}

		err := st.Append(ctx, store.Prediction{
			ID:        uuid.New().String(),
			Timestamp: m.Date,
			Input:     features,
			FEPods:    int(fePods),
			BEPods:    int(bePods),
		})
		if err != nil {
			log.Fatal("prediction storage", err)
		}
	}

	h, err := api.New(mdl, f, st, 5*time.Second)
	if err != nil {
		log.Fatalf("api init failed: %v", err)
	}
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           api.Routes(h),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	_ = srv.Shutdown(ctx)
}

//spreadSheetID := os.Getenv("GOOGLE_SHEETS_SPREADSHEET_ID")
//if spreadSheetID == "" {
//	log.Fatal("GOOGLE_SHEETS_SPREADSHEET_ID is not set")
//}
//
//jsonCreds := os.Getenv("GOOGLE_SHEETS_CREDENTIALS")
//if jsonCreds == "" {
//	log.Fatal("GOOGLE_SHEETS_CREDENTIALS is not set")
//}
//
//f, err := gsheets.NewFetcher(ctx, []byte(jsonCreds), spreadSheetID)
//if err != nil {
//	log.Fatalf("failed to create gsheets fetcher: %v", err)
//}
//
//fr, err := f.Fetch()
//if err != nil {
//	log.Fatalf("failed to fetch data: %v", err)
//}
