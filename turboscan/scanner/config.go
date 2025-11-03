package scanner

import "time"

// Config defines scanner configuration parameters.
type Config struct {
	BaseURL           string
	Threads           int
	Timeout           int
	StatusCodes       []int
	Verbose           bool
	MaxRetries        int
	RateLimit         int
	Recursive         bool
	MaxDepth          int
	Extensions        []string
	EnableSmartFilter bool
}

// Result represents a single scan result.
type Result struct {
	URL              string        `json:"url"`
	StatusCode       int           `json:"statusCode"`
	Size             int64         `json:"size"`
	Time             time.Duration `json:"time"`
	RedirectLocation string        `json:"redirectLocation,omitempty"`
}

// Stats holds aggregated scan statistics.
type Stats struct {
	Total     int
	Success   int
	Failed    int
	StartTime time.Time
	EndTime   time.Time
}
