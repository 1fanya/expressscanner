package scanner

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"
)

type HTTPClient struct {
	client *http.Client
	config Config
}

func NewHTTPClient(config Config) *HTTPClient {
	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 500,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // For security testing environments
		},
		DisableCompression:     false,
		DisableKeepAlives:      false,
		ForceAttemptHTTP2:      true,
		MaxResponseHeaderBytes: 4096,
		ResponseHeaderTimeout:  time.Duration(config.Timeout) * time.Second,
		TLSHandshakeTimeout:    10 * time.Second,
		ExpectContinueTimeout:  1 * time.Second,
		Proxy:                  http.ProxyFromEnvironment,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(config.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &HTTPClient{client: client, config: config}
}

func (c *HTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "ExpressScan/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *HTTPClient) GetWithRetry(url string, maxRetries int) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.Get(url)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if !isTemporaryError(err) {
			break
		}

		backoff := time.Duration(attempt+1) * time.Second
		time.Sleep(backoff)
	}
	return nil, lastErr
}

func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}
	return false
}
