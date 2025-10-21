package main

import (
	"context"
	"log"
	"os"

	"github.com/thisiscetin/podpredict/internal/data/fetcher/gsheets"
	"github.com/thisiscetin/podpredict/internal/model"
	"github.com/thisiscetin/podpredict/internal/model/linreg"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spreadSheetID := os.Getenv("GOOGLE_SHEETS_SPREADSHEET_ID")
	if spreadSheetID == "" {
		log.Fatal("GOOGLE_SHEETS_SPREADSHEET_ID is not set")
	}

	jsonCreds := os.Getenv("GOOGLE_SHEETS_CREDENTIALS")
	if jsonCreds == "" {
		log.Fatal("GOOGLE_SHEETS_CREDENTIALS is not set")
	}

	f, err := gsheets.NewFetcher(ctx, []byte(jsonCreds), spreadSheetID)
	if err != nil {
		log.Fatalf("failed to create gsheets fetcher: %v", err)
	}

	fr, err := f.Fetch()
	if err != nil {
		log.Fatalf("failed to fetch data: %v", err)
	}

	m := linreg.NewModel()
	if err := m.Train(fr); err != nil {
		log.Fatalf("failed to train model: %v", err)
	}

	for _, d := range fr {
		if !d.HasPods() {
			f := &model.Features{
				GMV:           d.GMV,
				Users:         d.Users,
				MarketingCost: d.MarketingCost,
			}

			fePods, bePods, err := m.Predict(f)

			log.Println(d, fePods, bePods, err)
		} else {
			log.Println(d)
		}
	}
}
