package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/apperror"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/coalesce"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

// RecommendIssueInput identifies one public issue and the contributor skills
// used for the explicit match denominator.
type RecommendIssueInput struct {
	Reference               issue.Reference
	DesiredSkills           []string
	ContributorProfile      issue.ContributorProfile
	IncludeRepositoryHealth bool
}

// RecommendIssueOutput contains the shared list/detail analysis plus
// operational metadata that remains outside the domain model.
type RecommendIssueOutput struct {
	Item             issue.RankedIssue
	RepositoryHealth issue.RepositoryHealthDashboard
	Dependencies     []string
	RateLimit        port.RateLimit
	Incomplete       bool
	CacheHit         bool
}

// IssueRecommender provides cached detail reads and a no-I/O fallback for
// candidates beyond the bounded enrichment window.
type IssueRecommender interface {
	// Execute returns cached or freshly loaded detail analysis, collapses
	// concurrent misses, and honors ctx cancellation.
	Execute(
		ctx context.Context,
		input RecommendIssueInput,
	) (RecommendIssueOutput, error)
	// EvaluateCandidate performs deterministic no-I/O fallback analysis with
	// unavailable repository evidence.
	EvaluateCandidate(
		candidate issue.Candidate,
		desiredSkills []string,
	) issue.RankedIssue
}

type recommendIssue struct {
	reader       port.GitHubIssueDetailReader
	healthReader port.RepositoryHealthReader
	cache        port.IssueDetailCache
	requests     coalesce.Group[string, issueDetailLoad]
	now          func() time.Time
}

// NewRecommendIssue composes detailed GitHub inspection with a concurrency-safe
// cache. Both dependencies are required.
func NewRecommendIssue(
	reader port.GitHubIssueDetailReader,
	cache port.IssueDetailCache,
	healthReaders ...port.RepositoryHealthReader,
) (IssueRecommender, error) {
	if reader == nil {
		return nil, fmt.Errorf(
			"compose issue recommendation: GitHub detail reader is required",
		)
	}
	if cache == nil {
		return nil, fmt.Errorf(
			"compose issue recommendation: detail cache is required",
		)
	}
	if len(healthReaders) > 1 {
		return nil, fmt.Errorf("compose issue recommendation: at most one health reader is supported")
	}
	var healthReader port.RepositoryHealthReader
	if len(healthReaders) == 1 {
		healthReader = healthReaders[0]
	}
	return &recommendIssue{
		reader: reader, healthReader: healthReader, cache: cache, now: time.Now,
	}, nil
}

func (usecase *recommendIssue) Execute(
	ctx context.Context,
	input RecommendIssueInput,
) (RecommendIssueOutput, error) {
	if err := ctx.Err(); err != nil {
		return RecommendIssueOutput{}, mapIssueDetailError(err)
	}
	key := input.Reference.CacheKey()
	if cached, found, err := usecase.cache.Get(ctx, key); err == nil && found {
		return usecase.output(
			cached,
			input.DesiredSkills,
			input.ContributorProfile,
			true,
			usecase.healthSnapshot(ctx, input),
		), nil
	} else if err != nil && ctx.Err() != nil {
		return RecommendIssueOutput{}, mapIssueDetailError(err)
	}

	load, err := usecase.requests.Do(ctx, key, func(
		sharedContext context.Context,
	) (issueDetailLoad, error) {
		if cached, found, err := usecase.cache.Get(
			sharedContext,
			key,
		); err == nil && found {
			return issueDetailLoad{detail: cached, cacheHit: true}, nil
		} else if err != nil && sharedContext.Err() != nil {
			return issueDetailLoad{}, err
		}

		detail, err := usecase.reader.GetIssueDetail(
			sharedContext,
			input.Reference.Owner(),
			input.Reference.RepositoryName(),
			input.Reference.Number(),
		)
		if err != nil {
			return issueDetailLoad{}, err
		}
		_ = usecase.cache.Set(sharedContext, key, detail)
		return issueDetailLoad{detail: detail}, nil
	})
	if err != nil {
		return RecommendIssueOutput{}, mapIssueDetailError(err)
	}
	return usecase.output(
		load.detail,
		input.DesiredSkills,
		input.ContributorProfile,
		load.cacheHit,
		usecase.healthSnapshot(ctx, input),
	), nil
}

func (usecase *recommendIssue) healthSnapshot(
	ctx context.Context,
	input RecommendIssueInput,
) issue.OpenSSFSnapshot {
	if !input.IncludeRepositoryHealth {
		return issue.OpenSSFSnapshot{Warning: "OpenSSF Scorecard was not requested for this view."}
	}
	return usecase.loadHealth(ctx, input.Reference)
}

type issueDetailLoad struct {
	detail   port.GitHubIssueDetailResult
	cacheHit bool
}

func (usecase *recommendIssue) output(
	detail port.GitHubIssueDetailResult,
	desiredSkills []string,
	contributorProfile issue.ContributorProfile,
	cacheHit bool,
	openSSF issue.OpenSSFSnapshot,
) RecommendIssueOutput {
	history := issue.IssueHistory{
		Comments: append(
			[]issue.CommentObservation(nil),
			detail.Comments...,
		),
		CommentsTruncated: detail.CommentsTruncated || detail.Incomplete,
		LinkedPullRequests: append(
			[]issue.LinkedPullRequestObservation(nil),
			detail.LinkedPullRequests...,
		),
		LinkedPullRequestsTruncated: detail.LinkedPullRequestsTruncated ||
			detail.Incomplete,
	}
	item := evaluateIssueRecommendation(
		detail.Candidate, detail.Dependencies, detail.RepositorySignals,
		detail.Activity,
		issue.DetectClaim(detail.Comments, detail.CommentsTruncated),
		history, desiredSkills, contributorProfile, usecase.now(),
	)
	return RecommendIssueOutput{
		Item: item,
		RepositoryHealth: issue.AnalyzeRepositoryHealth(
			item.Recommendation.RepositorySignals,
			item.Recommendation.Activity,
			openSSF,
			usecase.now(),
		),
		Dependencies: append(
			make([]string, 0, len(detail.Dependencies)),
			detail.Dependencies...,
		),
		RateLimit:  detail.RateLimit,
		Incomplete: detail.Incomplete,
		CacheHit:   cacheHit,
	}
}

func (usecase *recommendIssue) loadHealth(
	ctx context.Context,
	reference issue.Reference,
) issue.OpenSSFSnapshot {
	if usecase.healthReader == nil {
		return issue.OpenSSFSnapshot{Warning: "OpenSSF Scorecard integration is not configured."}
	}
	snapshot, err := usecase.healthReader.GetOpenSSFScorecard(
		ctx, reference.Owner(), reference.RepositoryName(),
	)
	if err != nil {
		return issue.OpenSSFSnapshot{Warning: "OpenSSF Scorecard could not be retrieved; GitHub analysis remains available."}
	}
	return snapshot
}

func (usecase *recommendIssue) EvaluateCandidate(
	candidate issue.Candidate,
	desiredSkills []string,
) issue.RankedIssue {
	lastMeaningfulUpdate := candidate.Repository.UpdatedAt
	if candidate.Repository.PushedAt.After(lastMeaningfulUpdate) {
		lastMeaningfulUpdate = candidate.Repository.PushedAt
	}
	return evaluateIssueRecommendation(
		candidate,
		nil,
		nil,
		issue.ActivityMetrics{
			LastMeaningfulUpdate: lastMeaningfulUpdate,
			CI:                   issue.CIStateUnknown,
		},
		issue.DetectClaim(nil, true),
		issue.IssueHistory{
			CommentsTruncated:           true,
			LinkedPullRequestsTruncated: true,
		},
		desiredSkills,
		issue.ContributorProfile{},
		usecase.now(),
	)
}

func evaluateIssueRecommendation(
	candidate issue.Candidate,
	dependencies []string,
	repositorySignals []issue.RepositorySignal,
	activity issue.ActivityMetrics,
	claim issue.ClaimEvidence,
	history issue.IssueHistory,
	desiredSkills []string,
	contributorProfile issue.ContributorProfile,
	now time.Time,
) issue.RankedIssue {
	hasMaintainerGuidance := false
	for _, signal := range repositorySignals {
		if signal.Key == issue.RepositoryContributing &&
			signal.State == issue.SignalPresent {
			hasMaintainerGuidance = true
			break
		}
	}
	analysis := issue.AnalyzeIssue(issue.AnalysisInput{
		Candidate:             candidate,
		Dependencies:          dependencies,
		HasMaintainerGuidance: hasMaintainerGuidance,
	})
	recommendation := issue.Recommend(issue.RecommendationInput{
		Candidate:          candidate,
		Analysis:           analysis,
		DesiredSkills:      append([]string(nil), desiredSkills...),
		ContributorProfile: contributorProfile,
		RepositorySignals:  repositorySignals,
		Activity:           activity,
		Claim:              claim,
		History:            history,
		Now:                now,
	})
	return issue.RankedIssue{
		Candidate:      candidate,
		Analysis:       analysis,
		Recommendation: recommendation,
		RepositoryHealth: issue.AnalyzeRepositoryHealth(
			repositorySignals, activity,
			issue.OpenSSFSnapshot{Warning: "OpenSSF Scorecard is loaded only on repository detail."},
			now,
		),
	}
}

func mapIssueDetailError(err error) error {
	switch {
	case errors.Is(err, issue.ErrInvalidReference):
		return apperror.Wrap(
			apperror.CodeInvalidRequest,
			"Issue reference is invalid",
			http.StatusBadRequest,
			err,
		)
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return apperror.Wrap(
			apperror.CodeRequestTimeout,
			"The request was cancelled or timed out",
			http.StatusGatewayTimeout,
			err,
		)
	case port.IsGitHubError(err, port.GitHubErrorNotFound):
		return apperror.Wrap(
			apperror.CodeNotFound,
			"GitHub issue was not found",
			http.StatusNotFound,
			err,
		)
	case port.IsGitHubError(err, port.GitHubErrorRateLimited):
		return apperror.Wrap(
			apperror.CodeRateLimit,
			"GitHub API rate limit was exceeded",
			http.StatusTooManyRequests,
			err,
		)
	default:
		return apperror.Wrap(
			apperror.CodeGitHubAPI,
			"Unable to retrieve GitHub issue details",
			http.StatusBadGateway,
			err,
		)
	}
}

var _ IssueRecommender = (*recommendIssue)(nil)
