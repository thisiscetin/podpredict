package config

import (
	"fmt"
	"os"
	"time"
)

const (
	DefaultAddr          = ":7000"
	DefaultEnvVarCreds   = "GOOGLE_SHEETS_CREDENTIALS"
	DefaultEnvVarSheetID = "GOOGLE_SHEETS_SPREADSHEET_ID"
)

type Config struct {
	Addr          string
	FetchTimeout  time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	CredsJSON     []byte
	SpreadsheetID string
}

func Load() (Config, error) {
	creds := os.Getenv(DefaultEnvVarCreds)
	if creds == "" {
		return Config{}, fmt.Errorf("%s is required", DefaultEnvVarCreds)
	}
	id := os.Getenv(DefaultEnvVarSheetID)
	if id == "" {
		return Config{}, fmt.Errorf("%s is required", DefaultEnvVarSheetID)
	}

	return Config{
		Addr:          DefaultAddr,
		FetchTimeout:  5 * time.Second,
		ReadTimeout:   5 * time.Second,
		WriteTimeout:  10 * time.Second,
		IdleTimeout:   60 * time.Second,
		CredsJSON:     []byte(creds),
		SpreadsheetID: id,
	}, nil
}
