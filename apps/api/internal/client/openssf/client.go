// Package openssf normalizes bounded OpenSSF Scorecard REST observations.
package openssf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
)

const maximumResponseBytes = 1 << 20
const maximumConcurrentRequests = 5

// Client retrieves and caches published Scorecard analyses.
type Client struct {
	baseURL  *url.URL
	http     *http.Client
	ttl      time.Duration
	capacity int
	now      func() time.Time
	mu       sync.Mutex
	cache    map[string]cacheEntry
	slots    chan struct{}
}

type cacheEntry struct {
	snapshot issue.OpenSSFSnapshot
	expires  time.Time
	created  time.Time
}

// NewClient validates bounded request and cache settings.
func NewClient(
	baseURL *url.URL,
	timeout, ttl time.Duration,
	capacity int,
) (*Client, error) {
	if baseURL == nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("compose OpenSSF client: valid base URL is required")
	}
	if timeout <= 0 || ttl <= 0 || capacity <= 0 {
		return nil, fmt.Errorf("compose OpenSSF client: positive bounds are required")
	}
	return &Client{
		baseURL: baseURL, http: &http.Client{Timeout: timeout},
		ttl: ttl, capacity: capacity, now: time.Now,
		cache: make(map[string]cacheEntry, capacity),
		slots: make(chan struct{}, maximumConcurrentRequests),
	}, nil
}

// GetOpenSSFScorecard returns a caller-owned normalized snapshot.
func (client *Client) GetOpenSSFScorecard(
	ctx context.Context,
	owner, repositoryName string,
) (issue.OpenSSFSnapshot, error) {
	key := strings.ToLower(owner + "/" + repositoryName)
	if snapshot, found := client.cached(key); found {
		return cloneSnapshot(snapshot), nil
	}
	var lastError error
	for attempt := 0; attempt < 2; attempt++ {
		snapshot, retry, err := client.request(ctx, owner, repositoryName)
		if err == nil {
			client.store(key, snapshot)
			return cloneSnapshot(snapshot), nil
		}
		lastError = err
		if !retry {
			break
		}
	}
	return issue.OpenSSFSnapshot{}, lastError
}

func (client *Client) request(
	ctx context.Context,
	owner, repositoryName string,
) (issue.OpenSSFSnapshot, bool, error) {
	select {
	case client.slots <- struct{}{}:
		defer func() { <-client.slots }()
	case <-ctx.Done():
		return issue.OpenSSFSnapshot{}, false, ctx.Err()
	}
	endpoint := *client.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") +
		"/projects/github.com/" + owner + "/" + repositoryName
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return issue.OpenSSFSnapshot{}, false, fmt.Errorf("build OpenSSF request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := client.http.Do(request)
	if err != nil {
		return issue.OpenSSFSnapshot{}, false, fmt.Errorf("request OpenSSF Scorecard: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return issue.OpenSSFSnapshot{Warning: "No published OpenSSF Scorecard analysis was found."}, false, nil
	}
	if response.StatusCode != http.StatusOK {
		retry := response.StatusCode == http.StatusBadGateway ||
			response.StatusCode == http.StatusServiceUnavailable
		return issue.OpenSSFSnapshot{}, retry, fmt.Errorf("OpenSSF Scorecard returned status %d", response.StatusCode)
	}
	limited := io.LimitReader(response.Body, maximumResponseBytes+1)
	content, err := io.ReadAll(limited)
	if err != nil {
		return issue.OpenSSFSnapshot{}, false, fmt.Errorf("read OpenSSF Scorecard: %w", err)
	}
	if len(content) > maximumResponseBytes {
		return issue.OpenSSFSnapshot{}, false, fmt.Errorf("OpenSSF Scorecard response exceeded size limit")
	}
	var payload scorecardResponse
	if err := json.Unmarshal(content, &payload); err != nil {
		return issue.OpenSSFSnapshot{}, false, fmt.Errorf("decode OpenSSF Scorecard: %w", err)
	}
	if len(payload.Checks) > 50 {
		return issue.OpenSSFSnapshot{}, false, fmt.Errorf("OpenSSF Scorecard contained too many checks")
	}
	checks := make([]issue.OpenSSFCheck, 0, len(payload.Checks))
	for _, check := range payload.Checks {
		name := strings.TrimSpace(check.Name)
		if name == "" || len(name) > 64 || check.Score < -1 || check.Score > 10 {
			return issue.OpenSSFSnapshot{}, false, fmt.Errorf("OpenSSF Scorecard schema was unsupported")
		}
		var score *int
		if check.Score >= 0 {
			value := check.Score
			score = &value
		}
		checks = append(checks, issue.OpenSSFCheck{Name: name, Score: score})
	}
	analyzedAt := payload.Date.UTC()
	now := client.now().UTC()
	return issue.OpenSSFSnapshot{
		Available: true, AnalyzedAt: analyzedAt,
		UpstreamVersion: strings.TrimSpace(payload.Scorecard.Version),
		Stale:           analyzedAt.IsZero() || now.Sub(analyzedAt) > 30*24*time.Hour,
		Checks:          checks,
	}, false, nil
}

type scorecardResponse struct {
	Date      time.Time `json:"date"`
	Scorecard struct {
		Version string `json:"version"`
	} `json:"scorecard"`
	Checks []struct {
		Name  string `json:"name"`
		Score int    `json:"score"`
	} `json:"checks"`
}

func (client *Client) cached(key string) (issue.OpenSSFSnapshot, bool) {
	client.mu.Lock()
	defer client.mu.Unlock()
	entry, found := client.cache[key]
	if !found || !client.now().Before(entry.expires) {
		delete(client.cache, key)
		return issue.OpenSSFSnapshot{}, false
	}
	return entry.snapshot, true
}

func (client *Client) store(key string, snapshot issue.OpenSSFSnapshot) {
	client.mu.Lock()
	defer client.mu.Unlock()
	now := client.now()
	if len(client.cache) >= client.capacity {
		oldestKey := ""
		var oldest time.Time
		for candidateKey, entry := range client.cache {
			if oldestKey == "" || entry.created.Before(oldest) {
				oldestKey, oldest = candidateKey, entry.created
			}
		}
		delete(client.cache, oldestKey)
	}
	client.cache[key] = cacheEntry{snapshot: cloneSnapshot(snapshot), expires: now.Add(client.ttl), created: now}
}

func cloneSnapshot(snapshot issue.OpenSSFSnapshot) issue.OpenSSFSnapshot {
	snapshot.Checks = append([]issue.OpenSSFCheck(nil), snapshot.Checks...)
	for index := range snapshot.Checks {
		if snapshot.Checks[index].Score != nil {
			value := *snapshot.Checks[index].Score
			snapshot.Checks[index].Score = &value
		}
	}
	return snapshot
}
