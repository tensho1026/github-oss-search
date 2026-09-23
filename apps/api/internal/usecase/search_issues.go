package usecase

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/apperror"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/coalesce"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

const maximumSearchIssuesPerRepository = 3

// SearchIssuesInput contains validated domain criteria and the requested
// application-level page.
type SearchIssuesInput struct {
	Criteria   issue.SearchCriteria
	Pagination issue.Pagination
}

// SearchIssuesPagination describes a page over at most the configured
// candidate window, after every eligibility rule has run.
type SearchIssuesPagination struct {
	Page       int
	PerPage    int
	Total      int
	TotalPages int
	HasNext    bool
}

// SearchIssuesOutput retains operational metadata without exposing GitHub
// payloads directly to the transport layer.
type SearchIssuesOutput struct {
	Items                         []issue.RankedIssue
	Pagination                    SearchIssuesPagination
	ExclusionCounts               map[issue.ExclusionReason]int
	CandidatesChecked             int
	UpstreamTotal                 int
	EnrichmentAttempted           int
	EnrichmentFailed              int
	GitHubIncomplete              bool
	EnrichmentIncomplete          bool
	PartialMatches                bool
	ContributionProfileStatus     issue.ContributionProfileStatus
	ContributionProfileIncomplete bool
	ContributionProfileCacheHit   bool
	RateLimit                     port.RateLimit
	CacheHit                      bool
}

// SearchIssues returns a post-filtered application page over a bounded
// GitHub candidate window. When exact preference filters would hide every
// safe issue, implementations keep the closest remaining matches and may
// perform one additional broadened GitHub search. Implementations must honor
// ctx and preserve incomplete-result metadata.
type SearchIssues interface {
	// Execute returns a post-filtered page, collapses concurrent misses, bounds
	// optional detail fan-out, and honors ctx.
	Execute(
		ctx context.Context,
		input SearchIssuesInput,
	) (SearchIssuesOutput, error)
}

type searchIssues struct {
	searcher        port.GitHubIssueSearcher
	cache           port.IssueSearchCache
	rankingCache    port.IssueSearchCache
	resultLimit     int
	recommender     IssueRecommender
	profileAnalyzer AnalyzeGitHubProfile
	analysisLimit   int
	maxConcurrency  int
	requests        coalesce.Group[string, port.IssueSearchCacheEntry]
	now             func() time.Time
}

// SearchIssuesOption configures optional bounded detail enrichment while
// preserving the candidate-only constructor used by isolated search tests.
type SearchIssuesOption func(*searchIssues) error

// WithContributionProfileAnalysis enables bounded public-profile evidence for
// Contribution Match. Profile failure remains non-fatal to issue discovery.
func WithContributionProfileAnalysis(
	analyzer AnalyzeGitHubProfile,
) SearchIssuesOption {
	return func(usecase *searchIssues) error {
		if analyzer == nil {
			return fmt.Errorf("profile analyzer is required")
		}
		usecase.profileAnalyzer = analyzer
		return nil
	}
}

// WithIssueSearchRankingCache enables a separately bounded cache for the
// expensive post-analysis ranking. Keeping it separate from the candidate
// cache prevents enriched responses from consuming the larger candidate
// cache's entire memory budget.
func WithIssueSearchRankingCache(
	cache port.IssueSearchCache,
) SearchIssuesOption {
	return func(usecase *searchIssues) error {
		if cache == nil {
			return fmt.Errorf("issue search ranking cache is required")
		}
		usecase.rankingCache = cache
		return nil
	}
}

// WithIssueRecommendationEnrichment enables detailed analysis for at most
// analysisLimit eligible candidates with bounded parallel GitHub requests.
func WithIssueRecommendationEnrichment(
	recommender IssueRecommender,
	analysisLimit int,
	maxConcurrency int,
) SearchIssuesOption {
	return func(usecase *searchIssues) error {
		if recommender == nil {
			return fmt.Errorf("issue recommender is required")
		}
		if analysisLimit < 1 || analysisLimit > usecase.resultLimit {
			return fmt.Errorf(
				"analysis limit must be between 1 and %d",
				usecase.resultLimit,
			)
		}
		if maxConcurrency < 1 || maxConcurrency > analysisLimit {
			return fmt.Errorf(
				"analysis concurrency must be between 1 and %d",
				analysisLimit,
			)
		}
		usecase.recommender = recommender
		usecase.analysisLimit = analysisLimit
		usecase.maxConcurrency = maxConcurrency
		return nil
	}
}

// NewSearchIssues validates search bounds and applies optional detail
// enrichment. Concurrent misses for an identical canonical key are collapsed.
func NewSearchIssues(
	searcher port.GitHubIssueSearcher,
	cache port.IssueSearchCache,
	resultLimit int,
	options ...SearchIssuesOption,
) (SearchIssues, error) {
	if searcher == nil {
		return nil, fmt.Errorf("compose issue search: GitHub searcher is required")
	}
	if cache == nil {
		return nil, fmt.Errorf("compose issue search: cache is required")
	}
	if resultLimit < 1 || resultLimit > issue.MaximumCandidateResults {
		return nil, fmt.Errorf(
			"compose issue search: result limit must be between 1 and %d",
			issue.MaximumCandidateResults,
		)
	}
	contract := &searchIssues{
		searcher:       searcher,
		cache:          cache,
		resultLimit:    resultLimit,
		maxConcurrency: 1,
		now:            time.Now,
	}
	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("compose issue search: option is required")
		}
		if err := option(contract); err != nil {
			return nil, fmt.Errorf("compose issue search: %w", err)
		}
	}
	return contract, nil
}

func (usecase *searchIssues) Execute(
	ctx context.Context,
	input SearchIssuesInput,
) (SearchIssuesOutput, error) {
	if err := ctx.Err(); err != nil {
		return SearchIssuesOutput{}, mapIssueSearchError(err)
	}

	key := input.Criteria.CacheKey()
	if cached, found, err := usecase.cache.Get(ctx, key); err == nil && found {
		return usecase.issueSearchOutput(
			ctx,
			cached,
			input,
			true,
		)
	} else if err != nil && ctx.Err() != nil {
		return SearchIssuesOutput{}, mapIssueSearchError(err)
	}

	entry, err := usecase.requests.Do(ctx, key, func(
		sharedContext context.Context,
	) (port.IssueSearchCacheEntry, error) {
		if cached, found, err := usecase.cache.Get(
			sharedContext,
			key,
		); err == nil && found {
			return cached, nil
		} else if err != nil && sharedContext.Err() != nil {
			return port.IssueSearchCacheEntry{}, err
		}

		entry, err := usecase.loadCandidateWindow(sharedContext, input.Criteria)
		if err != nil {
			return port.IssueSearchCacheEntry{}, err
		}

		_ = usecase.cache.Set(sharedContext, key, entry)
		return entry, nil
	})
	if err != nil {
		return SearchIssuesOutput{}, mapIssueSearchError(err)
	}
	return usecase.issueSearchOutput(
		ctx,
		entry,
		input,
		false,
	)
}

func (usecase *searchIssues) issueSearchOutput(
	ctx context.Context,
	entry port.IssueSearchCacheEntry,
	input SearchIssuesInput,
	cacheHit bool,
) (SearchIssuesOutput, error) {
	if !entry.RankedCandidatesReady && usecase.rankingCache != nil {
		if rankedEntry, found, err := usecase.rankingCache.Get(
			ctx,
			input.Criteria.CacheKey(),
		); err == nil && found && rankedEntry.RankedCandidatesReady {
			entry.RankedCandidates = rankedEntry.RankedCandidates
			entry.RankedCandidatesReady = true
			entry.RecommendationAttempted = rankedEntry.RecommendationAttempted
			entry.RecommendationFailed = rankedEntry.RecommendationFailed
			entry.RecommendationIncomplete = rankedEntry.RecommendationIncomplete
			entry.RecommendationEnrichedKeys = append(
				[]string(nil),
				rankedEntry.RecommendationEnrichedKeys...,
			)
			entry.ContributionProfileStatus = rankedEntry.ContributionProfileStatus
			entry.ContributionProfileIncomplete = rankedEntry.ContributionProfileIncomplete
			entry.ContributionProfileCacheHit = rankedEntry.ContributionProfileCacheHit
			entry.RateLimit = mergeRateLimits(entry.RateLimit, rankedEntry.RateLimit)
		} else if err != nil && ctx.Err() != nil {
			return SearchIssuesOutput{}, mapIssueSearchError(err)
		}
	}
	var (
		ranked             []issue.RankedIssue
		recommendationMeta issueRecommendationMeta
		profileMeta        contributionProfileMeta
		rateLimit          port.RateLimit
	)
	analysisLimit := usecase.analysisLimitFor(input, len(entry.Candidates))
	if entry.RankedCandidatesReady && usecase.hasEnrichmentFor(
		entry.Candidates,
		entry.RecommendationEnrichedKeys,
		entry.RecommendationAttempted,
		entry.RecommendationFailed,
		analysisLimit,
	) {
		// The expensive profile/detail analysis is valid for the same short
		// issue-search TTL. Keep only post-analysis ordering and request-local
		// filters below so page and sort changes do not repeat enrichment.
		ranked = entry.RankedCandidates
		recommendationMeta = issueRecommendationMeta{
			attempted:  entry.RecommendationAttempted,
			failed:     entry.RecommendationFailed,
			incomplete: entry.RecommendationIncomplete,
			enrichedKeys: append(
				[]string(nil),
				entry.RecommendationEnrichedKeys...,
			),
		}
		profileMeta = contributionProfileMeta{
			status:     entry.ContributionProfileStatus,
			incomplete: entry.ContributionProfileIncomplete,
			cacheHit:   entry.ContributionProfileCacheHit,
		}
		rateLimit = entry.RateLimit
	} else {
		contributorProfile, loadedProfileMeta, profileRateLimit :=
			usecase.loadContributionProfile(ctx, input.Criteria)
		var err error
		ranked, recommendationMeta, err = usecase.recommendCandidates(
			ctx,
			entry.Candidates,
			input.Criteria,
			contributorProfile,
			analysisLimit,
			entry.RankedCandidates,
			entry.RecommendationEnrichedKeys,
			issueRecommendationMeta{
				attempted:  entry.RecommendationAttempted,
				failed:     entry.RecommendationFailed,
				incomplete: entry.RecommendationIncomplete,
			},
		)
		if err != nil {
			return SearchIssuesOutput{}, mapIssueSearchError(err)
		}
		profileMeta = loadedProfileMeta
		rateLimit = mergeRateLimits(
			mergeRateLimits(entry.RateLimit, recommendationMeta.rateLimit),
			profileRateLimit,
		)

		entry.RankedCandidates = ranked
		entry.RankedCandidatesReady = true
		entry.RecommendationAttempted = recommendationMeta.attempted
		entry.RecommendationFailed = recommendationMeta.failed
		entry.RecommendationIncomplete = recommendationMeta.incomplete
		entry.RecommendationEnrichedKeys = append(
			[]string(nil),
			recommendationMeta.enrichedKeys...,
		)
		entry.ContributionProfileStatus = profileMeta.status
		entry.ContributionProfileIncomplete = profileMeta.incomplete
		entry.ContributionProfileCacheHit = profileMeta.cacheHit
		entry.RateLimit = rateLimit
		if usecase.rankingCache != nil {
			rankingEntry := port.IssueSearchCacheEntry{
				RankedCandidates:              ranked,
				RankedCandidatesReady:         true,
				RecommendationAttempted:       recommendationMeta.attempted,
				RecommendationFailed:          recommendationMeta.failed,
				RecommendationIncomplete:      recommendationMeta.incomplete,
				RecommendationEnrichedKeys:    recommendationMeta.enrichedKeys,
				ContributionProfileStatus:     profileMeta.status,
				ContributionProfileIncomplete: profileMeta.incomplete,
				ContributionProfileCacheHit:   profileMeta.cacheHit,
				RateLimit:                     rateLimit,
			}
			_ = usecase.rankingCache.Set(
				ctx,
				input.Criteria.CacheKey(),
				rankingEntry,
			)
		}
	}
	ranked, staleExcluded, partialMatches := applyPostAnalysisFilters(
		ranked,
		input.Criteria,
		entry.PartialMatches,
	)
	ranked = issue.SortRankedIssues(ranked, input.Criteria.SortBy())
	if partialMatches {
		now := usecase.now()
		slices.SortStableFunc(ranked, func(left, right issue.RankedIssue) int {
			return cmp.Compare(
				rankedPreferenceMatchCount(input.Criteria, right, now),
				rankedPreferenceMatchCount(input.Criteria, left, now),
			)
		})
	}
	ranked = limitIssuesPerRepository(ranked)
	total := len(ranked)
	totalPages := 0
	if total > 0 {
		totalPages = (total + input.Pagination.PerPage - 1) /
			input.Pagination.PerPage
	}

	items := make([]issue.RankedIssue, 0)
	pageIndex := input.Pagination.Page - 1
	if total > 0 && pageIndex <= total/input.Pagination.PerPage {
		start := pageIndex * input.Pagination.PerPage
		if start < total {
			end := min(start+input.Pagination.PerPage, total)
			items = append(items, ranked[start:end]...)
		}
	}

	exclusionCounts := cloneExclusionCounts(entry.ExclusionCounts)
	if staleExcluded > 0 {
		exclusionCounts[issue.ExclusionStale] += staleExcluded
	}
	return SearchIssuesOutput{
		Items: items,
		Pagination: SearchIssuesPagination{
			Page:       input.Pagination.Page,
			PerPage:    input.Pagination.PerPage,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    input.Pagination.Page < totalPages,
		},
		ExclusionCounts:               exclusionCounts,
		CandidatesChecked:             entry.CandidatesChecked,
		UpstreamTotal:                 entry.UpstreamTotal,
		EnrichmentAttempted:           recommendationMeta.attempted,
		EnrichmentFailed:              recommendationMeta.failed,
		GitHubIncomplete:              entry.IncompleteResults,
		EnrichmentIncomplete:          recommendationMeta.incomplete,
		PartialMatches:                partialMatches,
		ContributionProfileStatus:     profileMeta.status,
		ContributionProfileIncomplete: profileMeta.incomplete,
		ContributionProfileCacheHit:   profileMeta.cacheHit,
		RateLimit:                     rateLimit,
		CacheHit:                      cacheHit,
	}, nil
}

func (usecase *searchIssues) loadContributionProfile(
	ctx context.Context,
	criteria issue.SearchCriteria,
) (issue.ContributorProfile, contributionProfileMeta, port.RateLimit) {
	explicit := explicitContributionProfile(desiredIssueSkills(criteria))
	if usecase.profileAnalyzer == nil {
		return explicit, contributionProfileMeta{
			status:     explicit.Status,
			incomplete: explicit.Status != issue.ContributionProfileAvailable,
		}, port.RateLimit{}
	}
	output, err := usecase.profileAnalyzer.Execute(ctx, criteria.Username())
	if err != nil {
		return explicit, contributionProfileMeta{
			status:     issue.ContributionProfileUnavailable,
			incomplete: true,
		}, port.RateLimit{}
	}
	profile, meta := contributionProfileFromAnalysis(output.Analysis, output.CacheHit)
	return profile, meta, output.RateLimit
}

type issueRecommendationMeta struct {
	attempted    int
	failed       int
	incomplete   bool
	rateLimit    port.RateLimit
	enrichedKeys []string
}

func (usecase *searchIssues) analysisLimitFor(
	input SearchIssuesInput,
	candidateCount int,
) int {
	if usecase.recommender == nil || candidateCount == 0 {
		return 0
	}
	limit := usecase.analysisLimit
	if input.Pagination.Page <= limit/input.Pagination.PerPage {
		limit = min(limit, input.Pagination.Page*input.Pagination.PerPage)
	}
	return min(limit, candidateCount)
}

func (usecase *searchIssues) hasEnrichmentFor(
	candidates []issue.Candidate,
	enrichedKeys []string,
	attempted int,
	failed int,
	limit int,
) bool {
	if usecase.recommender == nil || limit == 0 {
		return true
	}
	enriched := make(map[string]struct{}, len(enrichedKeys))
	for _, key := range enrichedKeys {
		enriched[key] = struct{}{}
	}
	required := make(map[string]struct{}, limit)
	for index := 0; index < limit; index++ {
		required[repositoryRecommendationKey(candidates[index])] = struct{}{}
	}
	if len(enriched) == 0 && failed == 0 && attempted >= len(required) {
		// Accept entries written before RecommendationEnrichedKeys existed.
		return true
	}
	for key := range required {
		if _, ok := enriched[key]; !ok {
			return false
		}
	}
	return true
}

func (usecase *searchIssues) recommendCandidates(
	ctx context.Context,
	candidates []issue.Candidate,
	criteria issue.SearchCriteria,
	contributorProfile issue.ContributorProfile,
	limit int,
	existingRanked []issue.RankedIssue,
	existingEnrichedKeys []string,
	previousMeta issueRecommendationMeta,
) ([]issue.RankedIssue, issueRecommendationMeta, error) {
	desiredSkills := desiredIssueSkills(criteria)
	ranked := make([]issue.RankedIssue, len(candidates))
	meta := previousMeta
	meta.enrichedKeys = append([]string(nil), existingEnrichedKeys...)
	existingByCandidate := make(map[string]issue.RankedIssue, len(existingRanked))
	for _, item := range existingRanked {
		existingByCandidate[issueRecommendationKey(item.Candidate)] = item
	}
	enrichedRepositories := make(map[string]struct{}, len(existingEnrichedKeys))
	for _, key := range existingEnrichedKeys {
		enrichedRepositories[key] = struct{}{}
	}
	detailOutputs := make([]RecommendIssueOutput, limit)
	detailErrors := make([]error, limit)
	leaderFor := make([]int, limit)
	for index := range leaderFor {
		leaderFor[index] = -1
	}
	leaders := make([]int, 0, limit)
	leaderByRepository := make(map[string]int, limit)
	for index := range limit {
		candidate := candidates[index]
		key := repositoryRecommendationKey(candidate)
		if _, alreadyEnriched := enrichedRepositories[key]; alreadyEnriched {
			continue
		}
		if leader, exists := leaderByRepository[key]; exists {
			leaderFor[index] = leader
			continue
		}
		leaderByRepository[key] = index
		leaderFor[index] = index
		leaders = append(leaders, index)
	}
	meta.attempted += len(leaders)

	group, groupContext := errgroup.WithContext(ctx)
	group.SetLimit(usecase.maxConcurrency)
	for _, index := range leaders {
		index := index
		group.Go(func() error {
			candidate := candidates[index]
			reference, err := issue.NewReference(
				candidate.Repository.Owner,
				candidate.Repository.Name,
				candidate.Issue.Number,
			)
			if err != nil {
				detailErrors[index] = err
				return nil
			}
			output, err := usecase.recommender.Execute(
				groupContext,
				RecommendIssueInput{
					Reference:          reference,
					DesiredSkills:      desiredSkills,
					ContributorProfile: contributorProfile,
				},
			)
			if err != nil {
				if groupContext.Err() != nil {
					return groupContext.Err()
				}
				detailErrors[index] = err
				return nil
			}
			detailOutputs[index] = output
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, issueRecommendationMeta{}, err
	}

	for _, leader := range leaders {
		if detailErrors[leader] == nil {
			enrichedRepositories[repositoryRecommendationKey(candidates[leader])] = struct{}{}
		}
	}
	meta.enrichedKeys = meta.enrichedKeys[:0]
	for key := range enrichedRepositories {
		meta.enrichedKeys = append(meta.enrichedKeys, key)
	}
	slices.Sort(meta.enrichedKeys)

	for index, candidate := range candidates {
		if existing, ok := existingByCandidate[issueRecommendationKey(candidate)]; ok {
			ranked[index] = existing
		}
		if index < limit {
			leader := leaderFor[index]
			if leader < 0 {
				if ranked[index].Candidate.Issue.Number != 0 {
					continue
				}
				ranked[index] = fallbackRecommendation(
					usecase,
					candidate,
					desiredSkills,
					contributorProfile,
					false,
				)
				continue
			}
			if detailErrors[leader] == nil {
				output := detailOutputs[leader]
				if index == leader {
					ranked[index] = output.Item
					meta.rateLimit = mergeRateLimits(
						meta.rateLimit,
						output.RateLimit,
					)
				} else {
					ranked[index] = sharedRepositoryRecommendation(
						candidate,
						output.Item.Recommendation,
						output.Dependencies,
						desiredSkills,
						contributorProfile,
						usecase.now(),
					)
				}
				meta.incomplete = meta.incomplete || output.Incomplete
				continue
			}
			if index == leader {
				meta.failed++
				meta.incomplete = true
			}
		}
		if ranked[index].Candidate.Issue.Number != 0 {
			continue
		}
		ranked[index] = fallbackRecommendation(
			usecase,
			candidate,
			desiredSkills,
			contributorProfile,
			index < limit,
		)
	}
	return issue.RankIssues(ranked), meta, nil
}

func repositoryRecommendationKey(candidate issue.Candidate) string {
	return strings.ToLower(
		candidate.Repository.Owner + "/" + candidate.Repository.Name,
	)
}

func issueRecommendationKey(candidate issue.Candidate) string {
	return repositoryRecommendationKey(candidate) + "#" +
		fmt.Sprint(candidate.Issue.Number)
}

func sharedRepositoryRecommendation(
	candidate issue.Candidate,
	repositoryRecommendation issue.Recommendation,
	dependencies []string,
	desiredSkills []string,
	contributorProfile issue.ContributorProfile,
	now time.Time,
) issue.RankedIssue {
	ranked := evaluateIssueRecommendation(
		candidate,
		dependencies,
		repositoryRecommendation.RepositorySignals,
		repositoryRecommendation.Activity,
		issue.DetectClaim(nil, true),
		issue.IssueHistory{
			CommentsTruncated:           true,
			LinkedPullRequestsTruncated: true,
		},
		desiredSkills,
		contributorProfile,
		now,
	)
	ranked.Recommendation.Warnings = append(
		ranked.Recommendation.Warnings,
		issue.Warning{
			Code:     "claim_evidence_unavailable",
			Severity: issue.SeverityInfo,
			Message: "Repository evidence was reused, but this issue's " +
				"comment window was not inspected",
			Evidence: []issue.Evidence{{
				RuleID:      "recommendation.claim.unavailable",
				Source:      issue.EvidenceDerived,
				Description: "claim detection was not run for this list candidate",
			}},
		},
	)
	return ranked
}

func fallbackRecommendation(
	recommender *searchIssues,
	candidate issue.Candidate,
	desiredSkills []string,
	contributorProfile issue.ContributorProfile,
	enrichmentFailed bool,
) issue.RankedIssue {
	var ranked issue.RankedIssue
	if recommender.recommender != nil && !contributorProfile.Personalized {
		ranked = recommender.recommender.EvaluateCandidate(
			candidate,
			desiredSkills,
		)
	} else {
		ranked = evaluateIssueRecommendation(
			candidate,
			nil,
			nil,
			issue.ActivityMetrics{
				LastMeaningfulUpdate: candidate.Repository.UpdatedAt,
				CI:                   issue.CIStateUnknown,
			},
			issue.DetectClaim(nil, true),
			issue.IssueHistory{
				CommentsTruncated:           true,
				LinkedPullRequestsTruncated: true,
			},
			desiredSkills,
			contributorProfile,
			recommender.now(),
		)
	}
	if enrichmentFailed {
		ranked.Recommendation.Warnings = append(
			ranked.Recommendation.Warnings,
			issue.Warning{
				Code:     "detail_enrichment_unavailable",
				Severity: issue.SeverityInfo,
				Message: "Detailed repository inspection was unavailable; " +
					"the score uses bounded candidate metadata",
				Evidence: []issue.Evidence{{
					RuleID:      "recommendation.detail.unavailable",
					Source:      issue.EvidenceDerived,
					Description: "bounded detail enrichment did not complete",
				}},
			},
		)
	}
	return ranked
}

func desiredIssueSkills(criteria issue.SearchCriteria) []string {
	languages := criteria.Languages()
	frameworks := criteria.Frameworks()
	skills := make([]string, 0, len(languages)+len(frameworks))
	for _, language := range languages {
		skills = append(skills, language.String())
	}
	for _, framework := range frameworks {
		skills = append(skills, framework.String())
	}
	return skills
}

func cloneExclusionCounts(
	counts map[issue.ExclusionReason]int,
) map[issue.ExclusionReason]int {
	cloned := make(map[issue.ExclusionReason]int, len(counts))
	for reason, count := range counts {
		cloned[reason] = count
	}
	return cloned
}

func mergeExclusionCounts(
	left map[issue.ExclusionReason]int,
	right map[issue.ExclusionReason]int,
) map[issue.ExclusionReason]int {
	merged := cloneExclusionCounts(left)
	for reason, count := range right {
		merged[reason] += count
	}
	return merged
}

func mapIssueSearchError(err error) error {
	switch {
	case errors.Is(err, issue.ErrInvalidSearchCriteria):
		return apperror.Wrap(
			apperror.CodeInvalidRequest,
			"Issue search criteria are invalid",
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
			"Unable to search GitHub issues",
			http.StatusBadGateway,
			err,
		)
	}
}

var _ SearchIssues = (*searchIssues)(nil)
