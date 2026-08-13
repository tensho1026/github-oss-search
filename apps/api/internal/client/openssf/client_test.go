package openssf

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientNormalizesCachesAndMarksUnknownChecks(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		if request.URL.Path != "/projects/github.com/acme/rocket" {
			t.Errorf("path = %q", request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
          "date":"2026-08-10T00:00:00Z",
          "scorecard":{"version":"v5.2.1"},
          "checks":[{"name":"Maintained","score":9},{"name":"Code-Review","score":-1}]
        }`))
	}))
	defer server.Close()
	baseURL, _ := url.Parse(server.URL)
	client, err := NewClient(baseURL, time.Second, time.Hour, 10)
	if err != nil {
		t.Fatal(err)
	}
	client.now = func() time.Time { return time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC) }
	first, err := client.GetOpenSSFScorecard(context.Background(), "acme", "rocket")
	if err != nil {
		t.Fatal(err)
	}
	second, err := client.GetOpenSSFScorecard(context.Background(), "acme", "rocket")
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 || !first.Available || first.Stale ||
		first.UpstreamVersion != "v5.2.1" || len(first.Checks) != 2 ||
		first.Checks[0].Score == nil || *first.Checks[0].Score != 9 ||
		first.Checks[1].Score != nil || len(second.Checks) != 2 {
		t.Fatalf("first = %+v, second = %+v, calls = %d", first, second, calls.Load())
	}
}

func TestClientHandlesAbsentRetryAndMalformedResults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		statuses      []int
		body          string
		wantAvailable bool
		wantCalls     int32
		wantError     bool
	}{
		{name: "absent", statuses: []int{http.StatusNotFound}, wantCalls: 1},
		{name: "retry", statuses: []int{http.StatusServiceUnavailable, http.StatusOK}, body: `{"date":"2026-08-13T00:00:00Z","checks":[]}`, wantAvailable: true, wantCalls: 2},
		{name: "malformed", statuses: []int{http.StatusOK}, body: `{"checks":[{"name":"Maintained","score":11}]}`, wantCalls: 1, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				call := calls.Add(1)
				status := test.statuses[min(int(call)-1, len(test.statuses)-1)]
				writer.WriteHeader(status)
				if status == http.StatusOK {
					_, _ = writer.Write([]byte(test.body))
				}
			}))
			defer server.Close()
			baseURL, _ := url.Parse(server.URL)
			client, err := NewClient(baseURL, time.Second, time.Hour, 10)
			if err != nil {
				t.Fatal(err)
			}
			snapshot, err := client.GetOpenSSFScorecard(context.Background(), "acme", "rocket")
			if (err != nil) != test.wantError || snapshot.Available != test.wantAvailable || calls.Load() != test.wantCalls {
				t.Fatalf("snapshot = %+v, err = %v, calls = %d", snapshot, err, calls.Load())
			}
		})
	}
}
