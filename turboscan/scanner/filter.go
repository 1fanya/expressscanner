package scanner

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
)

type SmartFilter struct {
	mu             sync.RWMutex
	notFoundSize   int64
	notFoundDigest string
	ready          bool
}

func NewSmartFilter() *SmartFilter {
	return &SmartFilter{}
}

func (f *SmartFilter) Calibrate(client *HTTPClient, baseURL string) {
	if client == nil {
		return
	}

	randomPath := generateRandomPath()
	url := strings.TrimRight(baseURL, "/") + "/" + randomPath
	resp, err := client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return
	}

	f.mu.Lock()
	f.notFoundSize = resp.ContentLength
	f.notFoundDigest = hashBytes(body)
	f.ready = true
	f.mu.Unlock()
}

func (f *SmartFilter) IsReal(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	f.mu.RLock()
	ready := f.ready
	size := f.notFoundSize
	digest := f.notFoundDigest
	f.mu.RUnlock()

	if !ready {
		return true
	}

	if size >= 0 && resp.ContentLength == size {
		return false
	}

	if digest == "" {
		return true
	}

	body, err := prepareBodyForReuse(resp)
	if err != nil {
		return true
	}
	return hashBytes(body) != digest
}

func generateRandomPath() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "turboscan_random"
	}
	return hex.EncodeToString(b)
}

func hashBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	sum := uint32(2166136261)
	for _, b := range data {
		sum ^= uint32(b)
		sum *= 16777619
	}
	return hex.EncodeToString([]byte{
		byte(sum >> 24),
		byte(sum >> 16),
		byte(sum >> 8),
		byte(sum),
	})
}

func restoreBody(resp *http.Response, data []byte) {
	resp.Body = io.NopCloser(bytes.NewReader(data))
	resp.ContentLength = int64(len(data))
}

// prepareBodyForReuse reads the body and restores it for later use.
func prepareBodyForReuse(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, errors.New("nil response")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, err
	}
	restoreBody(resp, body)
	return body, nil
}
