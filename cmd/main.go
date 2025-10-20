package main

import (
	"context"
	"log"
	"os"

	"github.com/thisiscetin/podpredict/internal/data/fetcher/gsheets"
)

const spreadSheetID = "1fMGuVN9FY5jWt_YHH0MPoIiriYW4AJjMaHoDW-2Jdm4"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	for _, d := range fr {
		log.Printf("%+v\n", d)
	}
}
