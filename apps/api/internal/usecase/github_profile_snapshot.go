package usecase

import (
	"context"
	"sync"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/user"
)

// GitHubProfileSnapshotOutput combines the profile card and its derived
// public analysis so a profile page can be hydrated by one server request.
type GitHubProfileSnapshotOutput struct {
	User     GetGitHubUserOutput
	Analysis AnalyzeGitHubProfileOutput
}

// GetGitHubProfileSnapshot loads the two profile views concurrently while
// preserving the existing independently cached analysis path.
type GetGitHubProfileSnapshot interface {
	Execute(
		ctx context.Context,
		username user.Username,
	) (GitHubProfileSnapshotOutput, error)
}

type githubProfileSnapshot struct {
	getUser GetGitHubUser
	analyze AnalyzeGitHubProfile
}

// NewGitHubProfileSnapshot composes the profile snapshot orchestration.
func NewGitHubProfileSnapshot(
	getUser GetGitHubUser,
	analyze AnalyzeGitHubProfile,
) GetGitHubProfileSnapshot {
	return &githubProfileSnapshot{
		getUser: getUser,
		analyze: analyze,
	}
}

func (s *githubProfileSnapshot) Execute(
	ctx context.Context,
	username user.Username,
) (GitHubProfileSnapshotOutput, error) {
	var (
		output      GitHubProfileSnapshotOutput
		userErr     error
		analysisErr error
	)
	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		output.User, userErr = s.getUser.Execute(ctx, username)
	}()
	go func() {
		defer group.Done()
		output.Analysis, analysisErr = s.analyze.Execute(ctx, username)
	}()
	group.Wait()
	if userErr != nil {
		return GitHubProfileSnapshotOutput{}, userErr
	}
	if analysisErr != nil {
		return GitHubProfileSnapshotOutput{}, analysisErr
	}
	return output, nil
}

var _ GetGitHubProfileSnapshot = (*githubProfileSnapshot)(nil)
