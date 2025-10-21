package gsheets

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/thisiscetin/podpredict/internal/fetcher"
	"github.com/thisiscetin/podpredict/internal/metrics"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	readRange  = "Sheet1!A:F"
	dateLayout = "02/01/2006" // dd/mm/yyyy
)

// impl implements the fetcher.Fetcher interface for Google Sheets.
type impl struct {
	client        *sheets.Service
	spreadsheetID string
}

// NewFetcher creates a new Google Sheets fetcher using service account credentials.
// jsonCreds should contain the raw JSON of the service account key.
func NewFetcher(ctx context.Context, jsonCreds []byte, spreadsheetID string) (fetcher.Fetcher, error) {
	config, err := google.JWTConfigFromJSON(jsonCreds, sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	httpClient := config.Client(ctx)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &impl{
		client:        srv,
		spreadsheetID: spreadsheetID,
	}, nil
}

// Fetch retrieves metrics from the Google Sheet and converts them into a slice of metrics.Daily.
// It logs errors per row but continues processing other rows.
func (i *impl) Fetch() ([]metrics.Daily, error) {
	resp, err := i.client.Spreadsheets.Values.Get(i.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sheet data: %w", err)
	}

	var results []metrics.Daily
	for rowIdx, row := range resp.Values {
		if rowIdx == 0 {
			continue // skip header
		}

		daily, err := i.parseRow(row, rowIdx+1)
		if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, daily)
	}

	return results, nil
}

// parseRow parses a single row from the sheet into metrics.Daily.
func (i *impl) parseRow(row []any, rowNum int) (metrics.Daily, error) {
	var err error

	if len(row) < 4 {
		return metrics.Daily{}, fmt.Errorf("row %d: not enough columns", rowNum)
	}

	// Date
	dateStr, ok := row[0].(string)
	if !ok {
		return metrics.Daily{}, fmt.Errorf("row %d: invalid date format: %v", rowNum, row[0])
	}
	date, err := parseDate(dateStr)
	if err != nil {
		return metrics.Daily{}, fmt.Errorf("row %d: failed to parse date: %v", rowNum, err)
	}

	// GMV
	gmvStr, ok := row[1].(string)
	if !ok {
		return metrics.Daily{}, fmt.Errorf("row %d: invalid GMV format: %v", rowNum, row[1])
	}
	gmv, err := parseFloat(gmvStr)
	if err != nil {
		return metrics.Daily{}, fmt.Errorf("row %d: failed to parse GMV: %v", rowNum, err)
	}

	// Users
	usersStr, ok := row[2].(string)
	if !ok {
		return metrics.Daily{}, fmt.Errorf("row %d: invalid Users format: %v", rowNum, row[2])
	}
	users, err := parseInt(usersStr)
	if err != nil {
		return metrics.Daily{}, fmt.Errorf("row %d: failed to parse Users: %v", rowNum, err)
	}

	// Marketing Cost
	marketingStr, ok := row[3].(string)
	if !ok {
		return metrics.Daily{}, fmt.Errorf("row %d: invalid Marketing Cost format: %v", rowNum, row[3])
	}
	marketingCost, err := parseFloat(marketingStr)
	if err != nil {
		return metrics.Daily{}, fmt.Errorf("row %d: failed to parse Marketing Cost: %v", rowNum, err)
	}

	// FE Pods (optional)
	var fePods *int
	if len(row) > 4 && row[4] != "" {
		if s, ok := row[4].(string); ok {
			v, err := parseInt(s)
			if err == nil {
				fePods = &v
			} else {
				log.Printf("row %d: failed to parse FE Pods: %v", rowNum, err)
			}
		}
	}

	// BE Pods (optional)
	var bePods *int
	if len(row) > 5 && row[5] != "" {
		if s, ok := row[5].(string); ok {
			v, err := parseInt(s)
			if err == nil {
				bePods = &v
			} else {
				log.Printf("row %d: failed to parse BE Pods: %v", rowNum, err)
			}
		}
	}

	dailyMetric, err := metrics.NewDaily(date, gmv, users, marketingCost, fePods, bePods)
	if err != nil {
		return metrics.Daily{}, fmt.Errorf("row %d: failed to create Daily metric: %v", rowNum, err)
	}

	return dailyMetric, nil
}

// parseFloat cleans a string of commas and parses it as float64.
func parseFloat(s string) (float64, error) {
	clean := strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(clean, 64)
}

// parseInt parses a string as int.
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseDate parses a date string in dd/mm/yyyy format.
func parseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}
