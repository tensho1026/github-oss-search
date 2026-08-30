package memory

import (
	"context"
	"testing"
	"time"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/repository"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

func TestIssueSearchCacheReturnsIndependentCopies(t *testing.T) {
	cache := newTestIssueSearchCache(t, 2, time.Hour)
	entry := testIssueSearchEntry()
	if err := cache.Set(context.Background(), "criteria-a", entry); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	first, found, err := cache.Get(context.Background(), "criteria-a")
	if err != nil || !found {
		t.Fatalf("Get() found = %t, error = %v", found, err)
	}
	first.Candidates[0].Issue.Labels[0] = "mutated"
	first.Candidates[0].Issue.Assignees[0] = "mutated"
	first.ExclusionCounts[issue.ExclusionStale] = 100

	second, found, err := cache.Get(context.Background(), "criteria-a")
	if err != nil || !found {
		t.Fatalf("Get() found = %t, error = %v", found, err)
	}
	if second.Candidates[0].Issue.Labels[0] != "good first issue" ||
		second.Candidates[0].Issue.Assignees[0] != "contributor" ||
		second.ExclusionCounts[issue.ExclusionStale] != 2 {
		t.Fatalf("cached entry was mutated = %+v", second)
	}
}

func TestIssueSearchCacheClonesRankedResults(t *testing.T) {
	cache := newTestIssueSearchCache(t, 1, time.Hour)
	evidence := []issue.Evidence{{
		RuleID: "test.rule",
		Source: issue.EvidenceDerived,
	}}
	score := 7
	entry := port.IssueSearchCacheEntry{
		RankedCandidates: []issue.RankedIssue{{
			Candidate: issue.Candidate{Issue: issue.Summary{
				Labels:    []string{"original"},
				Assignees: []string{"owner"},
			}},
			Analysis: issue.Analysis{
				Quality: issue.QualityAssessment{Signals: []issue.QualitySignal{{
					Evidence: evidence,
				}}},
				RequiredTechnologies: []issue.RequiredTechnology{{Evidence: evidence}},
				Category: issue.CategoryAssessment{
					Matches:  []issue.Category{issue.CategoryBackend},
					Evidence: evidence,
				},
				Scope: issue.ChangeScope{
					Areas:    []issue.ChangeArea{issue.ChangeBackend},
					Evidence: evidence,
				},
				Difficulty: issue.DifficultyAssessment{Evidence: evidence},
				Effort:     issue.EffortEstimate{Evidence: evidence},
			},
			Recommendation: issue.Recommendation{
				SkillMatch: issue.SkillMatchAssessment{Skills: []issue.SkillMatch{{
					RequirementEvidence: evidence,
					ContributorEvidence: evidence,
				}}},
				RepositorySignals: []issue.RepositorySignal{{Evidence: evidence}},
				Claim:             issue.ClaimEvidence{Evidence: evidence},
				Stale:             issue.StaleAssessment{Evidence: evidence},
				Breakdown: issue.ScoreBreakdown{
					SkillMatch:        issue.ScoreComponent{Reasons: []string{"skill"}},
					IssueQuality:      issue.ScoreComponent{Reasons: []string{"quality"}},
					RepositoryQuality: issue.ScoreComponent{Reasons: []string{"repository"}},
					Activity:          issue.ScoreComponent{Reasons: []string{"activity"}},
					Maintainer:        issue.ScoreComponent{Reasons: []string{"maintainer"}},
					Availability:      issue.ScoreComponent{Reasons: []string{"availability"}},
				},
				Reasons:  []string{"reason"},
				Warnings: []issue.Warning{{Evidence: evidence}},
			},
			RepositoryHealth: issue.RepositoryHealthDashboard{Categories: []issue.HealthCategory{{
				Score: &score,
				Components: []issue.HealthComponent{{
					Score: &score,
				}},
				Warnings: []string{"warning"},
			}}},
		}},
		RankedCandidatesReady: true,
	}
	if err := cache.Set(context.Background(), "ranked", entry); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	first, found, err := cache.Get(context.Background(), "ranked")
	if err != nil || !found {
		t.Fatalf("Get() found = %t, error = %v", found, err)
	}
	first.RankedCandidates[0].Analysis.Quality.Signals[0].Evidence[0].Description = "mutated"
	first.RankedCandidates[0].Recommendation.Reasons[0] = "mutated"
	*first.RankedCandidates[0].RepositoryHealth.Categories[0].Score = 99

	second, found, err := cache.Get(context.Background(), "ranked")
	if err != nil || !found {
		t.Fatalf("Get() found = %t, error = %v", found, err)
	}
	if second.RankedCandidates[0].Analysis.Quality.Signals[0].Evidence[0].Description == "mutated" ||
		second.RankedCandidates[0].Recommendation.Reasons[0] == "mutated" ||
		*second.RankedCandidates[0].RepositoryHealth.Categories[0].Score == 99 {
		t.Fatalf("ranked result was mutated = %+v", second)
	}
}

func TestIssueSearchCacheExpiresAndEvicts(t *testing.T) {
	cache := newTestIssueSearchCache(t, 1, time.Minute)
	now := time.Date(2026, time.July, 30, 0, 0, 0, 0, time.UTC)
	cache.store.now = func() time.Time { return now }
	ctx := context.Background()

	if err := cache.Set(ctx, "first", testIssueSearchEntry()); err != nil {
		t.Fatalf("Set(first) error = %v", err)
	}
	if err := cache.Set(ctx, "second", testIssueSearchEntry()); err != nil {
		t.Fatalf("Set(second) error = %v", err)
	}
	if _, found, _ := cache.Get(ctx, "first"); found {
		t.Fatal("least recently used entry was not evicted")
	}

	now = now.Add(time.Minute)
	if _, found, err := cache.Get(ctx, "second"); err != nil || found {
		t.Fatalf("expired Get() found = %t, error = %v", found, err)
	}
}

func TestIssueSearchCacheHonorsCancellation(t *testing.T) {
	cache := newTestIssueSearchCache(t, 1, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cache.Set(ctx, "criteria", testIssueSearchEntry()); err == nil {
		t.Fatal("Set() error = nil")
	}
	if _, found, err := cache.Get(ctx, "criteria"); err == nil || found {
		t.Fatalf("Get() found = %t, error = %v", found, err)
	}
}

func TestIssueSearchCacheRejectsInvalidConfiguration(t *testing.T) {
	for _, input := range []struct {
		capacity int
		ttl      time.Duration
	}{
		{capacity: 0, ttl: time.Minute},
		{capacity: 1, ttl: 0},
	} {
		if _, err := NewIssueSearch(input.capacity, input.ttl); err == nil {
			t.Fatalf("NewIssueSearch(%d, %s) error = nil", input.capacity, input.ttl)
		}
	}
}

func newTestIssueSearchCache(
	t *testing.T,
	capacity int,
	ttl time.Duration,
) *IssueSearch {
	t.Helper()
	cache, err := NewIssueSearch(capacity, ttl)
	if err != nil {
		t.Fatalf("NewIssueSearch() error = %v", err)
	}
	return cache
}

func testIssueSearchEntry() port.IssueSearchCacheEntry {
	return port.IssueSearchCacheEntry{
		Candidates: []issue.Candidate{{
			Repository: repository.Summary{FullName: "example/repo"},
			Issue: issue.Summary{
				Number:    1,
				Labels:    []string{"good first issue"},
				Assignees: []string{"contributor"},
			},
		}},
		ExclusionCounts: map[issue.ExclusionReason]int{
			issue.ExclusionStale: 2,
		},
		UpstreamTotal:     100,
		IncompleteResults: true,
		RateLimit:         port.RateLimit{Known: true, Remaining: 29},
	}
}
