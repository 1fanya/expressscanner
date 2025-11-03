package scanner

import "time"

// Config bundles the runtime options that drive UltraScan's behaviour.
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

// Result captures the details for a single HTTP request during enumeration.
type Result struct {
	URL              string        `json:"url"`
	StatusCode       int           `json:"statusCode"`
	Size             int64         `json:"size"`
	Time             time.Duration `json:"time"`
	RedirectLocation string        `json:"redirectLocation,omitempty"`
}

// Stats aggregates counters and timing information across a scan session.
type Stats struct {
	Total     int
	Success   int
	Failed    int
	StartTime time.Time
	EndTime   time.Time
}
