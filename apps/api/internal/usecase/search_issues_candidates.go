package usecase

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

func (usecase *searchIssues) loadCandidateWindow(
	ctx context.Context,
	criteria issue.SearchCriteria,
) (port.IssueSearchCacheEntry, error) {
	result, err := usecase.searcher.SearchIssues(
		ctx,
		criteria,
		usecase.resultLimit,
	)
	if err != nil {
		return port.IssueSearchCacheEntry{}, err
	}
	entry := filterIssueCandidates(criteria, result, usecase.now())
	if len(entry.Candidates) > 0 {
		return entry, nil
	}

	relaxed := criteria.RelaxedDiscovery()
	if relaxed.CacheKey() == criteria.CacheKey() {
		return entry, nil
	}
	relaxedResult, err := usecase.searcher.SearchIssues(
		ctx,
		relaxed,
		usecase.resultLimit,
	)
	if err != nil {
		return port.IssueSearchCacheEntry{}, err
	}
	relaxedEntry := filterIssueCandidates(criteria, relaxedResult, usecase.now())
	relaxedEntry.RateLimit = mergeRateLimits(entry.RateLimit, relaxedEntry.RateLimit)
	relaxedEntry.IncompleteResults = entry.IncompleteResults ||
		relaxedEntry.IncompleteResults
	relaxedEntry.ExclusionCounts = mergeExclusionCounts(
		entry.ExclusionCounts,
		relaxedEntry.ExclusionCounts,
	)
	return relaxedEntry, nil
}

func filterIssueCandidates(
	criteria issue.SearchCriteria,
	result port.GitHubIssueSearchResult,
	now time.Time,
) port.IssueSearchCacheEntry {
	exact := make([]issue.Candidate, 0, len(result.Candidates))
	partial := make([]issue.Candidate, 0, len(result.Candidates))
	exclusionCounts := make(map[issue.ExclusionReason]int)
	for _, candidate := range result.Candidates {
		reasons := issue.ExclusionReasons(criteria, candidate, now)
		if len(reasons) == 0 {
			exact = append(exact, candidate)
			continue
		}
		for _, reason := range reasons {
			exclusionCounts[reason]++
		}
		if issue.HasOnlyPreferenceExclusions(reasons) {
			partial = append(partial, candidate)
		}
	}

	candidates := exact
	partialMatches := false
	if len(exact) == 0 && len(partial) > 0 {
		candidates = partial
		partialMatches = true
		slices.SortStableFunc(candidates, func(left, right issue.Candidate) int {
			return cmp.Compare(
				issue.PreferenceMatchCount(criteria, right, now),
				issue.PreferenceMatchCount(criteria, left, now),
			)
		})
	}

	return port.IssueSearchCacheEntry{
		Candidates:        candidates,
		ExclusionCounts:   exclusionCounts,
		CandidatesChecked: len(result.Candidates),
		UpstreamTotal:     result.TotalCount,
		IncompleteResults: result.IncompleteResults,
		PartialMatches:    partialMatches,
		RateLimit:         result.RateLimit,
	}
}
