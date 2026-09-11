package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/profile"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/user"
)

func TestGitHubProfileSnapshotExecutesDependenciesConcurrently(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	getUser := profileSnapshotUserStub{started: started, release: release}
	analyze := profileSnapshotAnalysisStub{started: started, release: release}
	contract := NewGitHubProfileSnapshot(&getUser, &analyze)

	resultCh := make(chan struct {
		output GitHubProfileSnapshotOutput
		err    error
	}, 1)
	go func() {
		output, err := contract.Execute(
			context.Background(),
			user.Username("octocat"),
		)
		resultCh <- struct {
			output GitHubProfileSnapshotOutput
			err    error
		}{output, err}
	}()

	seen := map[string]bool{}
	for range 2 {
		seen[<-started] = true
	}
	close(release)
	result := <-resultCh
	if result.err != nil || !seen["user"] || !seen["analysis"] {
		t.Fatalf("result = %+v, seen = %+v", result, seen)
	}
	if result.output.User.Profile.Login.String() != "octocat" ||
		result.output.Analysis.Analysis.Username.String() != "octocat" {
		t.Fatalf("output = %+v", result.output)
	}
}

func TestGitHubProfileSnapshotPrefersUserError(t *testing.T) {
	want := errors.New("profile unavailable")
	contract := NewGitHubProfileSnapshot(
		&profileSnapshotUserStub{err: want},
		&profileSnapshotAnalysisStub{},
	)

	_, err := contract.Execute(context.Background(), user.Username("octocat"))
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

type profileSnapshotUserStub struct {
	started chan<- string
	release <-chan struct{}
	err     error
}

func (stub *profileSnapshotUserStub) Execute(
	_ context.Context,
	username user.Username,
) (GetGitHubUserOutput, error) {
	if stub.started != nil {
		stub.started <- "user"
		<-stub.release
	}
	if stub.err != nil {
		return GetGitHubUserOutput{}, stub.err
	}
	return GetGitHubUserOutput{
		Profile: user.Profile{Login: username},
	}, nil
}

type profileSnapshotAnalysisStub struct {
	started chan<- string
	release <-chan struct{}
}

func (stub *profileSnapshotAnalysisStub) Execute(
	_ context.Context,
	username user.Username,
) (AnalyzeGitHubProfileOutput, error) {
	if stub.started != nil {
		stub.started <- "analysis"
		<-stub.release
	}
	return AnalyzeGitHubProfileOutput{
		Analysis: profile.Analysis{Username: username},
	}, nil
}
