package issue

import (
	"testing"
	"time"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/repository"
)

func TestAssessStalenessKeepsOldIssueFreshAfterMaintainerActivity(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	input := staleTestInput(now)
	input.History.Comments = []CommentObservation{{
		AuthorLogin:       "maintainer",
		AuthorType:        AuthorHuman,
		AuthorAssociation: "MEMBER",
		CreatedAt:         now.AddDate(0, 0, -2),
	}}

	got := AssessStaleness(input)

	if got.State != StaleFresh || got.Confidence != ConfidenceHigh {
		t.Fatalf("AssessStaleness() = %+v", got)
	}
}

func TestAssessStalenessIgnoresBotOnlyActivity(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	input := staleTestInput(now)
	input.History.Comments = []CommentObservation{{
		AuthorLogin:       "stale[bot]",
		AuthorType:        AuthorBot,
		AuthorAssociation: "MEMBER",
		CreatedAt:         now.AddDate(0, 0, -1),
	}}

	got := AssessStaleness(input)

	if got.State != StaleStale || got.SampleSize != 1 {
		t.Fatalf("AssessStaleness() = %+v", got)
	}
}

func TestAssessStalenessReturnsUnknownForTruncatedOldHistory(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	input := staleTestInput(now)
	input.History.CommentsTruncated = true

	got := AssessStaleness(input)

	if got.State != StaleUnknown || got.Confidence != ConfidenceLow {
		t.Fatalf("AssessStaleness() = %+v", got)
	}
}

func TestAssessStalenessDetectsActiveAndMergedLinkedPullRequests(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	input := staleTestInput(now)
	input.History.LinkedPullRequests = []LinkedPullRequestObservation{{
		Number:    91,
		State:     StateOpen,
		UpdatedAt: now.AddDate(0, 0, -3),
	}}
	if got := AssessStaleness(input); got.State != StaleFresh {
		t.Fatalf("active linked PR = %+v", got)
	}

	mergedAt := now.AddDate(0, 0, -2)
	input.History.LinkedPullRequests[0].State = "closed"
	input.History.LinkedPullRequests[0].MergedAt = mergedAt
	if got := AssessStaleness(input); got.State != StaleStale {
		t.Fatalf("merged linked PR = %+v", got)
	}
}

func TestAssessStalenessFlagsArchivedRecentRepository(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	input := staleTestInput(now)
	input.Candidate.Issue.CreatedAt = now.AddDate(0, 0, -2)
	input.Candidate.Repository.IsArchived = true

	got := AssessStaleness(input)

	if got.State != StaleStale || got.Confidence != ConfidenceHigh {
		t.Fatalf("AssessStaleness() = %+v", got)
	}
}

func staleTestInput(now time.Time) RecommendationInput {
	return RecommendationInput{
		Candidate: Candidate{
			Repository: repository.Summary{
				UpdatedAt: now.AddDate(0, 0, -5),
				PushedAt:  now.AddDate(0, 0, -5),
			},
			Issue: Summary{
				CreatedAt: now.AddDate(0, 0, -240),
				UpdatedAt: now.AddDate(0, 0, -1),
			},
		},
		Activity: ActivityMetrics{
			LastMeaningfulUpdate: now.AddDate(0, 0, -5),
		},
		Now: now,
	}
}
