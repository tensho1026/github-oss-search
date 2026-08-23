package usecase

import (
	"context"
	"fmt"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

// ObserveGitHubReferenceInput is one bounded public reference check.
type ObserveGitHubReferenceInput struct {
	Kind           port.GitHubReferenceKind
	Owner          string
	RepositoryName string
	Number         int
}

// ObserveGitHubReference retrieves public state without copying object data.
type ObserveGitHubReference interface {
	// Execute validates and performs one public reference observation.
	Execute(context.Context, ObserveGitHubReferenceInput) (port.GitHubReferenceObservation, error)
}

type observeGitHubReference struct {
	reader port.GitHubReferenceReader
}

// NewObserveGitHubReference composes the read-only public observer.
func NewObserveGitHubReference(reader port.GitHubReferenceReader) ObserveGitHubReference {
	if reader == nil {
		return nil
	}
	return &observeGitHubReference{reader: reader}
}

func (usecase *observeGitHubReference) Execute(
	ctx context.Context,
	input ObserveGitHubReferenceInput,
) (port.GitHubReferenceObservation, error) {
	validationNumber := input.Number
	if input.Kind == port.GitHubReferenceRepository {
		validationNumber = 1
	} else if input.Number < 1 {
		return port.GitHubReferenceObservation{}, fmt.Errorf("reference number must be positive")
	}
	if input.Kind != port.GitHubReferenceRepository &&
		input.Kind != port.GitHubReferenceIssue &&
		input.Kind != port.GitHubReferencePullRequest {
		return port.GitHubReferenceObservation{}, fmt.Errorf("unsupported reference kind")
	}
	if _, err := issue.NewReference(input.Owner, input.RepositoryName, validationNumber); err != nil {
		return port.GitHubReferenceObservation{}, err
	}
	result, err := usecase.reader.ObserveReference(
		ctx, input.Kind, input.Owner, input.RepositoryName, input.Number,
	)
	if err != nil {
		return port.GitHubReferenceObservation{}, mapGitHubUserError(err)
	}
	return result, nil
}

var _ ObserveGitHubReference = (*observeGitHubReference)(nil)
