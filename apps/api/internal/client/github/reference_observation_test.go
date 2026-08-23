package github

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

func TestObserveReferenceNormalizesPublicStates(t *testing.T) {
	tests := []struct {
		name     string
		kind     port.GitHubReferenceKind
		number   int
		status   int
		body     string
		want     port.GitHubReferenceState
		wantPath string
	}{
		{name: "repository", kind: port.GitHubReferenceRepository, status: 200, body: `{}`, want: port.GitHubReferenceAvailable, wantPath: "/repos/acme/rocket"},
		{name: "open issue", kind: port.GitHubReferenceIssue, number: 42, status: 200, body: `{"state":"open"}`, want: port.GitHubReferenceOpen, wantPath: "/repos/acme/rocket/issues/42"},
		{name: "closed issue", kind: port.GitHubReferenceIssue, number: 42, status: 200, body: `{"state":"closed"}`, want: port.GitHubReferenceClosed, wantPath: "/repos/acme/rocket/issues/42"},
		{name: "merged pull request", kind: port.GitHubReferencePullRequest, number: 7, status: 200, body: `{"state":"closed","merged_at":"2026-08-01T00:00:00Z"}`, want: port.GitHubReferenceMerged, wantPath: "/repos/acme/rocket/pulls/7"},
		{name: "inaccessible", kind: port.GitHubReferenceIssue, number: 42, status: 404, body: `{}`, want: port.GitHubReferenceInaccessible, wantPath: "/repos/acme/rocket/issues/42"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != test.wantPath {
					t.Errorf("path = %q", request.URL.Path)
				}
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(test.status)
				_, _ = io.WriteString(writer, test.body)
			}))
			defer server.Close()
			result, err := newTestClient(t, server.URL, "token").ObserveReference(
				context.Background(), test.kind, "acme", "rocket", test.number,
			)
			if err != nil || result.State != test.want {
				t.Fatalf("ObserveReference() = %+v, %v", result, err)
			}
		})
	}
}

func TestObserveReferenceRejectsInvalidInputAndPayload(t *testing.T) {
	client := newTestClient(t, "https://example.test", "")
	if _, err := client.ObserveReference(context.Background(), port.GitHubReferenceKind("commit"), "acme", "rocket", 1); err == nil {
		t.Fatal("unsupported kind error = nil")
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(writer, `{`)
	}))
	defer server.Close()
	if _, err := newTestClient(t, server.URL, "").ObserveReference(
		context.Background(), port.GitHubReferenceIssue, "acme", "rocket", 1,
	); err == nil {
		t.Fatal("invalid payload error = nil")
	}
}
