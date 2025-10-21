package fetcher

import (
	"github.com/thisiscetin/podpredict/internal/metrics"
)

// Fetcher defines the interface for loading daily business metrics.
// A Fetcher is expected to provide all data required for model training
// and inference, returning a slice of metrics.Daily records sorted by date.
// Implementations should ensure deterministic, repeatable results where possible.
// Example implementations include:
//   - A Google Sheets fetcher that reads data from a spreadsheet range.
//   - A mock fetcher used in tests that returns static data.
type Fetcher interface {
	// Fetch retrieves a complete set of daily metric records.
	// The returned slice should contain one metrics.Daily value per day
	// with any necessary validation already performed by the implementation.
	// If the data source cannot be reached or the data cannot be parsed,
	// an error is returned.
	Fetch() ([]metrics.Daily, error)
}
