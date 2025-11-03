package scanner

import (
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

type Scanner struct {
	config      Config
	client      *HTTPClient
	filter      *SmartFilter
	rateLimiter *RateLimiter

	results []Result
	stats   Stats

	mutex      sync.Mutex
	filterOnce sync.Once
}

func NewScanner(config Config) *Scanner {
	if config.Threads <= 0 {
		config.Threads = 10
	}
	if config.Timeout <= 0 {
		config.Timeout = 10
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 0
	}
	if config.MaxDepth <= 0 {
		config.MaxDepth = 1
	}

	s := &Scanner{
		config:      config,
		client:      NewHTTPClient(config),
		rateLimiter: NewRateLimiter(config.RateLimit),
		results:     make([]Result, 0),
	}

	if config.EnableSmartFilter {
		s.filter = NewSmartFilter()
	}

	return s
}

func (s *Scanner) ensureCalibrated() {
	if s.filter == nil {
		return
	}
	s.filterOnce.Do(func() {
		s.filter.Calibrate(s.client, s.config.BaseURL)
	})
}

func (s *Scanner) Scan(words []string) []Result {
	return s.scanBase(s.config.BaseURL, words)
}

func (s *Scanner) ScanWithExtensions(words []string, extensions []string) []Result {
	expanded := expandWords(words, extensions)
	return s.scanBase(s.config.BaseURL, expanded)
}

func (s *Scanner) ScanRecursive(words []string, maxDepth int) []Result {
	visited := make(map[string]struct{})
	base := strings.TrimRight(s.config.BaseURL, "/")
	visited[base] = struct{}{}
	results := s.scanRecursive(base, words, 0, maxDepth, visited)
	sort.Slice(results, func(i, j int) bool {
		if results[i].StatusCode == results[j].StatusCode {
			return results[i].URL < results[j].URL
		}
		return results[i].StatusCode < results[j].StatusCode
	})
	return results
}

func (s *Scanner) scanRecursive(base string, words []string, depth, maxDepth int, visited map[string]struct{}) []Result {
	currentResults := s.scanBase(base, words)
	if depth >= maxDepth-1 {
		return currentResults
	}

	var recursiveResults []Result
	for _, result := range currentResults {
		if !isRedirect(result.StatusCode) {
			continue
		}
		location := result.RedirectLocation
		if location == "" {
			location = result.URL
		}
		nextURL := resolveURL(base, location)
		if nextURL == "" {
			continue
		}
		if _, ok := visited[nextURL]; ok {
			continue
		}
		visited[nextURL] = struct{}{}

		childResults := s.scanRecursive(nextURL, words, depth+1, maxDepth, visited)
		recursiveResults = append(recursiveResults, childResults...)
	}

	combined := append(currentResults, recursiveResults...)
	return combined
}

func (s *Scanner) scanBase(baseURL string, words []string) []Result {
	if len(words) == 0 {
		return nil
	}

	s.ensureCalibrated()

	start := time.Now()
	s.trackTotal(len(words), start)

	jobs := make(chan string)
	resultsCh := make(chan Result, len(words))

	var wg sync.WaitGroup
	for i := 0; i < s.config.Threads; i++ {
		wg.Add(1)
		go s.worker(baseURL, jobs, resultsCh, &wg)
	}

	go func() {
		for _, word := range words {
			jobs <- word
		}
		close(jobs)
		wg.Wait()
		close(resultsCh)
	}()

	localResults := make([]Result, 0)
	for result := range resultsCh {
		localResults = append(localResults, result)
		s.appendResult(result)
		fmt.Printf("[+] %d - %s [Size: %d] [Time: %v]\n", result.StatusCode, result.URL, result.Size, result.Time)
	}

	s.finishStats(time.Now())

	return localResults
}

func (s *Scanner) worker(baseURL string, jobs <-chan string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for word := range jobs {
		if s.rateLimiter != nil {
			s.rateLimiter.Wait()
		}

		target := buildURL(baseURL, word)
		start := time.Now()
		resp, err := s.client.GetWithRetry(target, s.config.MaxRetries)
		duration := time.Since(start)

		if err != nil {
			s.incrementFailed()
			if s.config.Verbose {
				fmt.Printf("[-] %s - Error: %v\n", target, err)
			}
			continue
		}

		redirectLocation := resp.Header.Get("Location")
		shouldReport := s.shouldReport(resp.StatusCode)
		isReal := true
		if s.filter != nil {
			isReal = s.filter.IsReal(resp)
		}

		size := resp.ContentLength
		if size < 0 {
			if body, err := prepareBodyForReuse(resp); err == nil {
				size = int64(len(body))
			}
		} else if s.filter != nil {
			// Smart filtering can adjust ContentLength after body restoration.
			size = resp.ContentLength
		}

		resp.Body.Close()

		if shouldReport && isReal {
			result := Result{
				URL:              target,
				StatusCode:       resp.StatusCode,
				Size:             size,
				Time:             duration,
				RedirectLocation: redirectLocation,
			}
			results <- result
			s.incrementSuccess()
		} else {
			s.incrementFailed()
		}
	}
}

func (s *Scanner) shouldReport(statusCode int) bool {
	if len(s.config.StatusCodes) == 0 {
		return true
	}
	for _, code := range s.config.StatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

func (s *Scanner) incrementSuccess() {
	s.mutex.Lock()
	s.stats.Success++
	s.mutex.Unlock()
}

func (s *Scanner) incrementFailed() {
	s.mutex.Lock()
	s.stats.Failed++
	s.mutex.Unlock()
}

func (s *Scanner) trackTotal(count int, start time.Time) {
	s.mutex.Lock()
	if s.stats.StartTime.IsZero() {
		s.stats.StartTime = start
	}
	s.stats.Total += count
	s.mutex.Unlock()
}

func (s *Scanner) finishStats(now time.Time) {
	s.mutex.Lock()
	s.stats.EndTime = now
	s.mutex.Unlock()
}

func (s *Scanner) appendResult(result Result) {
	s.mutex.Lock()
	s.results = append(s.results, result)
	s.mutex.Unlock()
}

func (s *Scanner) PrintStats() {
	s.mutex.Lock()
	stats := s.stats
	s.mutex.Unlock()

	duration := stats.EndTime.Sub(stats.StartTime)
	if duration <= 0 {
		duration = time.Since(stats.StartTime)
	}
	reqPerSec := 0.0
	if duration > 0 {
		reqPerSec = float64(stats.Total) / duration.Seconds()
	}

	fmt.Println("\n[*] Scan Statistics:")
	fmt.Printf("    Total Requests:  %d\n", stats.Total)
	fmt.Printf("    Successful:      %d\n", stats.Success)
	fmt.Printf("    Failed:          %d\n", stats.Failed)
	fmt.Printf("    Duration:        %v\n", duration)
	fmt.Printf("    Req/sec:         %.2f\n", reqPerSec)
}

func expandWords(words []string, extensions []string) []string {
	if len(extensions) == 0 {
		return append([]string{}, words...)
	}

	expanded := make([]string, 0, len(words)*(len(extensions)+1))
	for _, word := range words {
		expanded = append(expanded, word)
		for _, ext := range extensions {
			if strings.HasSuffix(word, ext) {
				continue
			}
			expanded = append(expanded, word+ext)
		}
	}
	return expanded
}

func isRedirect(status int) bool {
	return status == 301 || status == 302 || status == 307 || status == 308
}

func buildURL(base, segment string) string {
	cleanedBase := strings.TrimRight(base, "/")
	if segment == "" {
		return cleanedBase
	}
	return cleanedBase + "/" + strings.TrimLeft(segment, "/")
}

func resolveURL(base string, location string) string {
	if location == "" {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	loc, err := url.Parse(location)
	if err != nil {
		return ""
	}
	resolved := baseURL.ResolveReference(loc)
	resolved.Path = path.Clean(resolved.Path)
	return strings.TrimRight(resolved.String(), "/")
}
