package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/apperror"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

type referenceReaderStub struct {
	input  ObserveGitHubReferenceInput
	result port.GitHubReferenceObservation
	err    error
}

func (reader *referenceReaderStub) ObserveReference(
	_ context.Context,
	kind port.GitHubReferenceKind,
	owner, repositoryName string,
	number int,
) (port.GitHubReferenceObservation, error) {
	reader.input = ObserveGitHubReferenceInput{Kind: kind, Owner: owner, RepositoryName: repositoryName, Number: number}
	return reader.result, reader.err
}

func TestObserveGitHubReferenceValidatesAndExecutes(t *testing.T) {
	reader := &referenceReaderStub{result: port.GitHubReferenceObservation{State: port.GitHubReferenceOpen}}
	service := NewObserveGitHubReference(reader)
	input := ObserveGitHubReferenceInput{Kind: port.GitHubReferenceIssue, Owner: "Acme", RepositoryName: "rocket", Number: 42}
	result, err := service.Execute(context.Background(), input)
	if err != nil || result.State != port.GitHubReferenceOpen || reader.input != input {
		t.Fatalf("Execute() = %+v, %v; input = %+v", result, err, reader.input)
	}
	for _, invalid := range []ObserveGitHubReferenceInput{
		{Kind: port.GitHubReferenceIssue, Owner: "acme", RepositoryName: "rocket", Number: 0},
		{Kind: port.GitHubReferenceKind("commit"), Owner: "acme", RepositoryName: "rocket", Number: 1},
		{Kind: port.GitHubReferenceRepository, Owner: "bad owner", RepositoryName: "rocket"},
	} {
		if _, executeErr := service.Execute(context.Background(), invalid); executeErr == nil {
			t.Fatalf("Execute(%+v) error = nil", invalid)
		}
	}
	if NewObserveGitHubReference(nil) != nil {
		t.Fatal("NewObserveGitHubReference(nil) != nil")
	}
}

func TestObserveGitHubReferenceMapsUpstreamErrors(t *testing.T) {
	cause := &port.GitHubError{Kind: port.GitHubErrorRateLimited}
	_, err := NewObserveGitHubReference(&referenceReaderStub{err: cause}).Execute(
		context.Background(),
		ObserveGitHubReferenceInput{Kind: port.GitHubReferenceRepository, Owner: "acme", RepositoryName: "rocket"},
	)
	var applicationError *apperror.Error
	if !errors.As(err, &applicationError) || applicationError.Code != apperror.CodeRateLimit {
		t.Fatalf("Execute() error = %v", err)
	}
}
