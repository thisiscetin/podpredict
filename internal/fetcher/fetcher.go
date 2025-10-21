package fetcher

import "github.com/thisiscetin/podpredict/internal/metrics"

type Fetcher interface {
	Fetch() ([]metrics.Daily, error)
}
