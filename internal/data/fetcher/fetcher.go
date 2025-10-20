package fetcher

import "github.com/thisiscetin/podpredict/internal/data/metrics"

type Fetcher interface {
	Fetch() ([]metrics.Daily, error)
}
