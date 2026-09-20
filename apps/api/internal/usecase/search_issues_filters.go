package usecase

import (
	"time"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
)

func applyPostAnalysisFilters(
	ranked []issue.RankedIssue,
	criteria issue.SearchCriteria,
	partialMatches bool,
) ([]issue.RankedIssue, int, bool) {
	unfiltered := ranked
	filtered, staleExcluded := filterRankedIssuesByStale(ranked, criteria)
	effortFiltered := filterRankedIssuesByEffort(filtered, criteria)
	if len(effortFiltered) > 0 {
		return effortFiltered, staleExcluded, partialMatches
	}
	if len(filtered) > 0 {
		return filtered, staleExcluded, true
	}
	if len(unfiltered) > 0 {
		return unfiltered, 0, true
	}
	return effortFiltered, staleExcluded, partialMatches
}

func rankedPreferenceMatchCount(
	criteria issue.SearchCriteria,
	ranked issue.RankedIssue,
	now time.Time,
) int {
	score := issue.PreferenceMatchCount(criteria, ranked.Candidate, now)
	if criteria.IncludesStale() ||
		ranked.Recommendation.Stale.State != issue.StaleStale {
		score++
	}
	if maximum, configured := criteria.MaximumEffort(); !configured ||
		ranked.Analysis.Effort.Band.IsAtMost(maximum) {
		score++
	}
	return score
}

func filterRankedIssuesByStale(
	ranked []issue.RankedIssue,
	criteria issue.SearchCriteria,
) ([]issue.RankedIssue, int) {
	if criteria.IncludesStale() {
		return ranked, 0
	}
	filtered := make([]issue.RankedIssue, 0, len(ranked))
	excluded := 0
	for _, candidate := range ranked {
		if candidate.Recommendation.Stale.State == issue.StaleStale {
			excluded++
			continue
		}
		filtered = append(filtered, candidate)
	}
	return filtered, excluded
}

func filterRankedIssuesByEffort(
	ranked []issue.RankedIssue,
	criteria issue.SearchCriteria,
) []issue.RankedIssue {
	maximum, configured := criteria.MaximumEffort()
	if !configured {
		return ranked
	}

	filtered := make([]issue.RankedIssue, 0, len(ranked))
	for _, candidate := range ranked {
		if candidate.Analysis.Effort.Band.IsAtMost(maximum) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}
