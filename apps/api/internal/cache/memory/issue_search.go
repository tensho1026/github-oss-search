package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

// IssueSearch stores bounded, eligible candidate windows by canonical
// criteria hash. It owns deep copies of every mutable collection.
type IssueSearch struct {
	store *lruCache[string, port.IssueSearchCacheEntry]
}

// NewIssueSearch constructs an LRU cache with positive capacity and TTL.
func NewIssueSearch(capacity int, ttl time.Duration) (*IssueSearch, error) {
	store, err := newLRUCache[string, port.IssueSearchCacheEntry](
		capacity,
		ttl,
		cloneIssueSearchEntry,
	)
	if err != nil {
		return nil, fmt.Errorf("create issue search cache: %w", err)
	}
	return &IssueSearch{store: store}, nil
}

// Get returns a deep copy, reports misses without error, and honors a
// pre-cancelled context before acquiring the cache lock.
func (cache *IssueSearch) Get(
	ctx context.Context,
	key string,
) (port.IssueSearchCacheEntry, bool, error) {
	return cache.store.get(ctx, key)
}

// Set stores a deep copy, refreshes TTL on replacement, and may evict the
// least-recently-used entry.
func (cache *IssueSearch) Set(
	ctx context.Context,
	key string,
	entry port.IssueSearchCacheEntry,
) error {
	return cache.store.set(ctx, key, entry)
}

func cloneIssueSearchEntry(
	entry port.IssueSearchCacheEntry,
) port.IssueSearchCacheEntry {
	cloned := entry
	cloned.Candidates = make([]issue.Candidate, len(entry.Candidates))
	for index, candidate := range entry.Candidates {
		cloned.Candidates[index] = candidate
		cloned.Candidates[index].Issue.Labels = append(
			[]string(nil),
			candidate.Issue.Labels...,
		)
		cloned.Candidates[index].Issue.Assignees = append(
			[]string(nil),
			candidate.Issue.Assignees...,
		)
	}
	cloned.ExclusionCounts = make(
		map[issue.ExclusionReason]int,
		len(entry.ExclusionCounts),
	)
	for reason, count := range entry.ExclusionCounts {
		cloned.ExclusionCounts[reason] = count
	}
	cloned.RankedCandidates = cloneRankedIssues(entry.RankedCandidates)
	return cloned
}

func cloneRankedIssues(source []issue.RankedIssue) []issue.RankedIssue {
	cloned := make([]issue.RankedIssue, len(source))
	for index, ranked := range source {
		cloned[index] = ranked
		cloned[index].Candidate.Issue.Labels = append(
			[]string(nil),
			ranked.Candidate.Issue.Labels...,
		)
		cloned[index].Candidate.Issue.Assignees = append(
			[]string(nil),
			ranked.Candidate.Issue.Assignees...,
		)
		cloned[index].Analysis = cloneAnalysis(ranked.Analysis)
		cloned[index].Recommendation = cloneRecommendation(ranked.Recommendation)
		cloned[index].RepositoryHealth = cloneRepositoryHealth(ranked.RepositoryHealth)
	}
	return cloned
}

func cloneAnalysis(source issue.Analysis) issue.Analysis {
	cloned := source
	cloned.Quality.Signals = make([]issue.QualitySignal, len(source.Quality.Signals))
	for index, signal := range source.Quality.Signals {
		cloned.Quality.Signals[index] = signal
		cloned.Quality.Signals[index].Evidence = cloneEvidence(signal.Evidence)
	}
	cloned.RequiredTechnologies = make([]issue.RequiredTechnology, len(source.RequiredTechnologies))
	for index, technology := range source.RequiredTechnologies {
		cloned.RequiredTechnologies[index] = technology
		cloned.RequiredTechnologies[index].Evidence = cloneEvidence(technology.Evidence)
	}
	cloned.Category.Matches = append([]issue.Category(nil), source.Category.Matches...)
	cloned.Category.Evidence = cloneEvidence(source.Category.Evidence)
	cloned.Scope.Areas = append([]issue.ChangeArea(nil), source.Scope.Areas...)
	cloned.Scope.Evidence = cloneEvidence(source.Scope.Evidence)
	cloned.Difficulty.Evidence = cloneEvidence(source.Difficulty.Evidence)
	cloned.Effort.Evidence = cloneEvidence(source.Effort.Evidence)
	return cloned
}

func cloneRecommendation(source issue.Recommendation) issue.Recommendation {
	cloned := source
	cloned.SkillMatch.Skills = make([]issue.SkillMatch, len(source.SkillMatch.Skills))
	for index, skill := range source.SkillMatch.Skills {
		cloned.SkillMatch.Skills[index] = skill
		cloned.SkillMatch.Skills[index].RequirementEvidence = cloneEvidence(skill.RequirementEvidence)
		cloned.SkillMatch.Skills[index].ContributorEvidence = cloneEvidence(skill.ContributorEvidence)
	}
	cloned.RepositorySignals = make([]issue.RepositorySignal, len(source.RepositorySignals))
	for index, signal := range source.RepositorySignals {
		cloned.RepositorySignals[index] = signal
		cloned.RepositorySignals[index].Evidence = cloneEvidence(signal.Evidence)
	}
	cloned.Claim.Evidence = cloneEvidence(source.Claim.Evidence)
	cloned.Stale.Evidence = cloneEvidence(source.Stale.Evidence)
	cloned.Breakdown.SkillMatch.Reasons = append([]string(nil), source.Breakdown.SkillMatch.Reasons...)
	cloned.Breakdown.IssueQuality.Reasons = append([]string(nil), source.Breakdown.IssueQuality.Reasons...)
	cloned.Breakdown.RepositoryQuality.Reasons = append([]string(nil), source.Breakdown.RepositoryQuality.Reasons...)
	cloned.Breakdown.Activity.Reasons = append([]string(nil), source.Breakdown.Activity.Reasons...)
	cloned.Breakdown.Maintainer.Reasons = append([]string(nil), source.Breakdown.Maintainer.Reasons...)
	cloned.Breakdown.Availability.Reasons = append([]string(nil), source.Breakdown.Availability.Reasons...)
	cloned.Reasons = append([]string(nil), source.Reasons...)
	cloned.Warnings = make([]issue.Warning, len(source.Warnings))
	for index, warning := range source.Warnings {
		cloned.Warnings[index] = warning
		cloned.Warnings[index].Evidence = cloneEvidence(warning.Evidence)
	}
	return cloned
}

func cloneRepositoryHealth(source issue.RepositoryHealthDashboard) issue.RepositoryHealthDashboard {
	cloned := source
	cloned.Categories = make([]issue.HealthCategory, len(source.Categories))
	for index, category := range source.Categories {
		cloned.Categories[index] = category
		cloned.Categories[index].Score = cloneIntPointer(category.Score)
		cloned.Categories[index].Components = make([]issue.HealthComponent, len(category.Components))
		for componentIndex, component := range category.Components {
			cloned.Categories[index].Components[componentIndex] = component
			cloned.Categories[index].Components[componentIndex].Score = cloneIntPointer(component.Score)
		}
		cloned.Categories[index].Warnings = append([]string(nil), category.Warnings...)
	}
	return cloned
}

func cloneEvidence(source []issue.Evidence) []issue.Evidence {
	return append([]issue.Evidence(nil), source...)
}

func cloneIntPointer(source *int) *int {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}

var _ port.IssueSearchCache = (*IssueSearch)(nil)
