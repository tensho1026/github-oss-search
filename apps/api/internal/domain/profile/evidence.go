package profile

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/repository"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/user"
)

// Public-analysis bounds constrain time windows and response cardinality.
const (
	AnalysisWindowDays       = 365
	MaximumTechnologyResults = 20
	maximumLanguageResults   = 10
)

// EvidenceStatus distinguishes complete facts, bounded samples, and missing
// upstream evidence.
type EvidenceStatus string

// EvidenceStatus values preserve incomplete data instead of treating it as
// negative evidence.
const (
	EvidenceExact       EvidenceStatus = "exact"
	EvidenceSampled     EvidenceStatus = "sampled"
	EvidenceUnavailable EvidenceStatus = "unavailable"
)

// Confidence communicates how strongly the observed public sample supports a
// derived profile result.
type Confidence string

// Confidence values form the closed vocabulary exposed by profile analysis.
const (
	ConfidenceHigh        Confidence = "high"
	ConfidenceMedium      Confidence = "medium"
	ConfidenceLow         Confidence = "low"
	ConfidenceUnavailable Confidence = "unavailable"
)

// TechnologyKind separates programming languages from inferred frameworks.
type TechnologyKind string

// TechnologyKind values classify recent and proficient technologies.
const (
	TechnologyLanguage  TechnologyKind = "language"
	TechnologyFramework TechnologyKind = "framework"
)

// RepositorySource identifies how a repository relates to the analyzed user.
type RepositorySource string

// RepositorySource values describe mutually non-exclusive public collections.
const (
	RepositoryOwned       RepositorySource = "owned"
	RepositoryContributed RepositorySource = "contributed"
	RepositoryStarred     RepositorySource = "starred"
	RepositoryForked      RepositorySource = "forked"
)

// AnalysisWindow records the inclusive UTC evidence period and its
// public-data-only invariant.
type AnalysisWindow struct {
	From       time.Time
	To         time.Time
	Days       int
	PublicOnly bool
}

// CountEvidence is a raw count and the completeness metadata required to
// interpret zero safely.
type CountEvidence struct {
	Available bool
	Value     int
	Complete  bool
}

// CountMetric is a normalized count paired with a stable evidence status.
type CountMetric struct {
	Value  int
	Status EvidenceStatus
}

// RepositoryObservation contains bounded public evidence collected for one
// normalized repository.
type RepositoryObservation struct {
	Repository        repository.Summary
	Languages         map[string]int64
	LanguagesComplete bool
	Manifests         []Manifest
}

// RepositoryCollection is one bounded upstream collection and its sampling
// metadata.
type RepositoryCollection struct {
	Available    bool
	Repositories []RepositoryObservation
	Total        int
	TotalKnown   bool
	Limit        int
	HasMore      bool
}

// ProfileSnapshot is the complete transport-independent input to
// AnalyzeSnapshot. It contains public GitHub evidence only.
type ProfileSnapshot struct {
	Username      user.Username
	WindowFrom    time.Time
	WindowTo      time.Time
	Owned         RepositoryCollection
	Contributed   RepositoryCollection
	Starred       RepositoryCollection
	Forked        RepositoryCollection
	Contributions ContributionSnapshot
	Portfolio     PortfolioSnapshot
	Warnings      []Warning
}

// PortfolioContribution is one normalized public merged pull request.
type PortfolioContribution struct {
	RepositoryOwner string
	RepositoryName  string
	Number          int
	Title           string
	URL             string
	MergedAt        time.Time
	Language        string
}

// PortfolioSnapshot is the bounded upstream evidence for an OSS portfolio.
type PortfolioSnapshot struct {
	Available   bool
	TotalMerged int
	Complete    bool
	HasMore     bool
	Items       []PortfolioContribution
}

// PortfolioLanguageCount groups displayed merged pull requests by language.
type PortfolioLanguageCount struct {
	Name  string
	Count int
}

// ContributionPortfolio is the reproducible bounded portfolio analysis.
type ContributionPortfolio struct {
	Status          EvidenceStatus
	TotalMerged     int
	DisplayedMerged int
	RepositoryCount int
	HasMore         bool
	AnalyzedAt      time.Time
	Languages       []PortfolioLanguageCount
	Contributions   []PortfolioContribution
}

// JourneyMilestone is a dated, normalized OSS event backed by a canonical
// public GitHub URL. "First" milestones always mean first in the bounded
// observed sample, never a lifetime claim.
type JourneyMilestone struct {
	ID             string
	Kind           string
	OccurredAt     time.Time
	Title          string
	Description    string
	EvidenceURL    string
	RepositoryName string
	Technology     string
}

// OSSJourney is a deterministic timeline derived only from verified public
// contribution evidence.
type OSSJourney struct {
	Status     EvidenceStatus
	AnalyzedAt time.Time
	Milestones []JourneyMilestone
}

// StreakWeek groups canonical merged-PR evidence into one UTC Monday week.
type StreakWeek struct {
	StartedAt    time.Time
	EndedAt      time.Time
	EventCount   int
	EvidenceURLs []string
}

// ContributionStreak summarizes consecutive verified public activity weeks.
type ContributionStreak struct {
	Status          EvidenceStatus
	AnalyzedAt      time.Time
	Timezone        string
	WeekStartsOn    string
	CurrentWeeks    int
	LongestWeeks    int
	QualifyingWeeks int
	Weeks           []StreakWeek
}

// QuestProgress is one item in the versioned beginner OSS quest catalog.
type QuestProgress struct {
	ID          string
	Title       string
	Description string
	Status      string
	Current     int
	Target      int
	CompletedAt *time.Time
	EvidenceURL string
	NextAction  string
}

// OSSQuest is a deterministic, read-only quest evaluation. It never sends
// notifications or writes progress independently of canonical evidence.
type OSSQuest struct {
	CatalogVersion string
	EvaluatedAt    time.Time
	Completed      int
	Total          int
	NextQuestID    string
	Items          []QuestProgress
}

// ContributionSnapshot records public contribution counts before domain
// normalization.
type ContributionSnapshot struct {
	Available           bool
	Commits             CountEvidence
	IssuesOpened        CountEvidence
	PullRequestsOpened  CountEvidence
	PullRequestReviews  CountEvidence
	RepositoriesTouched CountEvidence
	Calendar            ContributionCalendar
}

// ContributionLevel is GitHub's normalized contribution-calendar intensity.
type ContributionLevel string

// ContributionLevel values are ordered from an empty day through the fourth
// public contribution quartile.
const (
	ContributionNone   ContributionLevel = "none"
	ContributionFirst  ContributionLevel = "first_quartile"
	ContributionSecond ContributionLevel = "second_quartile"
	ContributionThird  ContributionLevel = "third_quartile"
	ContributionFourth ContributionLevel = "fourth_quartile"
)

// ContributionDay is one public day at a server-authoritative grid position.
type ContributionDay struct {
	Date    time.Time
	Weekday int
	Count   int
	Level   ContributionLevel
}

// ContributionWeek is one ordered column in the contribution calendar.
type ContributionWeek struct {
	Index    int
	FirstDay time.Time
	Days     []ContributionDay
}

// ContributionCalendar is a bounded public one-year contribution grid.
type ContributionCalendar struct {
	Status EvidenceStatus
	Total  int
	From   time.Time
	To     time.Time
	Weeks  []ContributionWeek
}

// RepositorySample summarizes coverage and primary technologies for one
// repository source.
type RepositorySample struct {
	Status              EvidenceStatus
	Observed            int
	Total               *int
	Limit               int
	ActiveInWindow      int
	PrimaryTechnologies []LanguageShare
}

// RepositoryEvidence groups analysis coverage by relationship to the user.
type RepositoryEvidence struct {
	Owned       RepositorySample
	Contributed RepositorySample
	Starred     RepositorySample
	Forked      RepositorySample
}

// ContributionAnalysis is the normalized one-year public contribution view.
type ContributionAnalysis struct {
	WindowDays          int
	Commits             CountMetric
	IssuesOpened        CountMetric
	PullRequestsOpened  CountMetric
	PullRequestReviews  CountMetric
	RepositoriesTouched CountMetric
}

// TechnologyEvidence is one bounded, explainable input to a technology score.
type TechnologyEvidence struct {
	Kind   string
	Value  int
	Status EvidenceStatus
}

// OSSExperience is a coarse public-only experience level and its supporting
// evidence.
type OSSExperience struct {
	Level      string
	Confidence Confidence
	PublicOnly bool
	Evidence   []TechnologyEvidence
}

// RecentTechnology records a technology observed in the configured analysis
// window.
type RecentTechnology struct {
	Name              string
	Kind              TechnologyKind
	LastUsedAt        time.Time
	RepositoryCount   int
	RepositorySources []RepositorySource
	Confidence        Confidence
}

// TechnologyProficiency is a deterministic five-level technology assessment.
type TechnologyProficiency struct {
	Name       string
	Kind       TechnologyKind
	Level      int
	Label      string
	Score      int
	Confidence Confidence
	Evidence   []TechnologyEvidence
}

// AnalyzeSnapshot derives bounded, explainable profile evidence without
// performing I/O. Private activity is intentionally absent from ProfileSnapshot.
func AnalyzeSnapshot(snapshot ProfileSnapshot) Analysis {
	window := normalizeWindow(snapshot.WindowFrom, snapshot.WindowTo)
	ownedFrameworks := inferCollectionFrameworks(snapshot.Owned)
	ownedLanguageBytes := collectionLanguageBytes(snapshot.Owned)
	ownedLanguages := topLanguageShares(
		AggregateLanguages(ownedLanguageBytes),
		maximumLanguageResults,
	)
	frameworkEvidence := collectionFrameworkEvidence(snapshot.Owned, ownedFrameworks)
	frameworks := sortedFrameworkNames(frameworkEvidence)
	recent := analyzeRecentTechnologies(snapshot, window.From, ownedFrameworks)
	contributions := analyzeContributions(snapshot.Contributions)
	portfolio := analyzePortfolio(snapshot.Portfolio, window.To)
	journey := analyzeJourney(portfolio)
	streak := analyzeContributionStreak(portfolio)
	quest := analyzeOSSQuest(contributions, portfolio, window.To)
	repositoryEvidence := RepositoryEvidence{
		Owned: analyzeRepositoryCollection(
			snapshot.Owned,
			window.From,
			ownedLanguages,
		),
		Contributed: analyzeRepositoryCollection(
			snapshot.Contributed,
			window.From,
			nil,
		),
		Starred: analyzeRepositoryCollection(
			snapshot.Starred,
			window.From,
			nil,
		),
		Forked: analyzeRepositoryCollection(
			snapshot.Forked,
			window.From,
			nil,
		),
	}

	return Analysis{
		Username:           snapshot.Username,
		Languages:          ownedLanguages,
		LanguageStatus:     languageEvidenceStatus(snapshot.Owned),
		Frameworks:         frameworks,
		RecentTechnologies: recent,
		Contributions:      contributions,
		ContributionCalendar: cloneContributionCalendar(
			snapshot.Contributions.Calendar,
		),
		Portfolio:          portfolio,
		Journey:            journey,
		Streak:             streak,
		Quest:              quest,
		OSSExperience:      analyzeOSSExperience(contributions),
		RepositoryEvidence: repositoryEvidence,
		Proficiency: buildTechnologyProficiency(
			snapshot,
			ownedLanguages,
			frameworkEvidence,
			recent,
		),
		Window:               window,
		RepositoriesAnalyzed: len(snapshot.Owned.Repositories),
		Warnings:             slices.Clone(snapshot.Warnings),
	}
}

func inferCollectionFrameworks(collection RepositoryCollection) [][]string {
	frameworks := make([][]string, len(collection.Repositories))
	for index, observation := range collection.Repositories {
		frameworks[index] = InferFrameworks(observation.Manifests)
	}
	return frameworks
}

func cloneContributionCalendar(
	calendar ContributionCalendar,
) ContributionCalendar {
	weeks := make([]ContributionWeek, 0, len(calendar.Weeks))
	for _, week := range calendar.Weeks {
		week.Days = slices.Clone(week.Days)
		weeks = append(weeks, week)
	}
	calendar.Weeks = weeks
	if calendar.Status == "" {
		calendar.Status = EvidenceUnavailable
	}
	return calendar
}

func analyzeOSSQuest(
	contributions ContributionAnalysis,
	portfolio ContributionPortfolio,
	evaluatedAt time.Time,
) OSSQuest {
	firstPR := questFromCount(
		"first_pr", "Open your first pull request", "Create a public OSS pull request.",
		contributions.PullRequestsOpened, "Find a matching issue and open a focused PR.",
	)
	firstReview := questFromCount(
		"first_review", "Complete your first review", "Review a public OSS pull request.",
		contributions.PullRequestReviews, "Review an open PR and leave actionable feedback.",
	)
	firstMerged := QuestProgress{
		ID: "first_merge", Title: "Get your first PR merged",
		Description: "Have a public OSS pull request merged.", Target: 1,
		NextAction: "Respond to review feedback on an open pull request.",
	}
	threeRepositories := QuestProgress{
		ID: "three_repositories", Title: "Contribute to 3 repositories",
		Description: "Build verified merged contributions across three public projects.",
		Target:      3, NextAction: "Find a suitable issue in a new repository.",
	}
	if portfolio.Status == EvidenceUnavailable {
		firstMerged.Status = "unavailable"
		threeRepositories.Status = "unavailable"
	} else {
		firstMerged.Current = min(1, portfolio.DisplayedMerged)
		threeRepositories.Current = min(3, portfolio.RepositoryCount)
		firstMerged.Status = "in_progress"
		threeRepositories.Status = "locked"
		if len(portfolio.Contributions) > 0 {
			earliest := portfolio.Contributions[len(portfolio.Contributions)-1]
			completedAt := earliest.MergedAt
			firstMerged.Status = "completed"
			firstMerged.CompletedAt = &completedAt
			firstMerged.EvidenceURL = earliest.URL
		}
		if portfolio.RepositoryCount >= 3 {
			threeRepositories.Status = "completed"
			completedAt := portfolio.AnalyzedAt
			threeRepositories.CompletedAt = &completedAt
		} else if firstMerged.Status == "completed" {
			threeRepositories.Status = "in_progress"
		}
	}
	items := []QuestProgress{
		{
			ID: "first_issue_comment", Title: "Comment on your first issue",
			Description: "Join a public issue discussion.", Status: "unavailable",
			Target: 1, NextAction: "Comment on an issue after confirming you can help.",
		},
		firstPR, firstReview, firstMerged, threeRepositories,
	}
	if firstPR.Status != "completed" && firstReview.Status != "completed" {
		firstReview.Status = "locked"
	}
	completed := 0
	nextQuestID := ""
	for _, item := range items {
		if item.Status == "completed" {
			completed++
		} else if nextQuestID == "" && item.Status != "unavailable" && item.Status != "locked" {
			nextQuestID = item.ID
		}
	}
	return OSSQuest{
		CatalogVersion: "2026-08-01", EvaluatedAt: evaluatedAt,
		Completed: completed, Total: len(items), NextQuestID: nextQuestID, Items: items,
	}
}

func questFromCount(
	id, title, description string,
	metric CountMetric,
	nextAction string,
) QuestProgress {
	item := QuestProgress{
		ID: id, Title: title, Description: description,
		Current: min(1, metric.Value), Target: 1, NextAction: nextAction,
	}
	if metric.Status == EvidenceUnavailable {
		item.Status = "unavailable"
	} else if metric.Value > 0 {
		item.Status = "completed"
	} else {
		item.Status = "in_progress"
	}
	return item
}

func analyzeContributionStreak(portfolio ContributionPortfolio) ContributionStreak {
	result := ContributionStreak{
		Status: portfolio.Status, AnalyzedAt: portfolio.AnalyzedAt,
		Timezone: "UTC", WeekStartsOn: "monday", Weeks: []StreakWeek{},
	}
	if portfolio.Status == EvidenceUnavailable {
		return result
	}
	type weekBucket struct {
		start time.Time
		urls  []string
	}
	buckets := make(map[string]*weekBucket)
	for _, contribution := range portfolio.Contributions {
		start := startOfUTCWeek(contribution.MergedAt)
		key := start.Format("2006-01-02")
		bucket, exists := buckets[key]
		if !exists {
			bucket = &weekBucket{start: start, urls: []string{}}
			buckets[key] = bucket
		}
		if !slices.Contains(bucket.urls, contribution.URL) {
			bucket.urls = append(bucket.urls, contribution.URL)
		}
	}
	starts := make([]time.Time, 0, len(buckets))
	for _, bucket := range buckets {
		slices.Sort(bucket.urls)
		starts = append(starts, bucket.start)
	}
	slices.SortFunc(starts, func(left, right time.Time) int { return left.Compare(right) })
	longest := 0
	run := 0
	for index, start := range starts {
		if index == 0 || start.Sub(starts[index-1]) == 7*24*time.Hour {
			run++
		} else {
			run = 1
		}
		longest = max(longest, run)
	}
	current := 0
	currentWeek := startOfUTCWeek(portfolio.AnalyzedAt)
	for index := len(starts) - 1; index >= 0; index-- {
		expected := currentWeek.AddDate(0, 0, -7*current)
		if !starts[index].Equal(expected) {
			break
		}
		current++
	}
	weeks := make([]StreakWeek, 0, len(starts))
	for index := len(starts) - 1; index >= 0; index-- {
		bucket := buckets[starts[index].Format("2006-01-02")]
		weeks = append(weeks, StreakWeek{
			StartedAt:  bucket.start,
			EndedAt:    bucket.start.AddDate(0, 0, 7).Add(-time.Nanosecond),
			EventCount: len(bucket.urls), EvidenceURLs: slices.Clone(bucket.urls),
		})
	}
	result.CurrentWeeks = current
	result.LongestWeeks = longest
	result.QualifyingWeeks = len(starts)
	result.Weeks = weeks
	return result
}

func startOfUTCWeek(value time.Time) time.Time {
	value = value.UTC()
	day := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
	daysSinceMonday := (int(day.Weekday()) + 6) % 7
	return day.AddDate(0, 0, -daysSinceMonday)
}

func analyzeJourney(portfolio ContributionPortfolio) OSSJourney {
	journey := OSSJourney{
		Status: portfolio.Status, AnalyzedAt: portfolio.AnalyzedAt,
		Milestones: []JourneyMilestone{},
	}
	if portfolio.Status == EvidenceUnavailable {
		return journey
	}
	items := slices.Clone(portfolio.Contributions)
	slices.SortFunc(items, func(left, right PortfolioContribution) int {
		if comparison := left.MergedAt.Compare(right.MergedAt); comparison != 0 {
			return comparison
		}
		leftKey := strings.ToLower(left.RepositoryOwner + "/" + left.RepositoryName)
		rightKey := strings.ToLower(right.RepositoryOwner + "/" + right.RepositoryName)
		if comparison := strings.Compare(leftKey, rightKey); comparison != 0 {
			return comparison
		}
		return left.Number - right.Number
	})
	seenRepositories := make(map[string]struct{})
	seenTechnologies := make(map[string]struct{})
	for _, item := range items {
		repositoryName := item.RepositoryOwner + "/" + item.RepositoryName
		repositoryKey := strings.ToLower(repositoryName)
		journey.Milestones = append(journey.Milestones, JourneyMilestone{
			ID:   "merged:" + repositoryKey + "#" + strconv.Itoa(item.Number),
			Kind: "merged_pull_request", OccurredAt: item.MergedAt,
			Title:       "Merged PR #" + strconv.Itoa(item.Number) + " in " + repositoryName,
			Description: "Observed public merge: " + item.Title,
			EvidenceURL: item.URL, RepositoryName: repositoryName,
		})
		if _, seen := seenRepositories[repositoryKey]; !seen {
			seenRepositories[repositoryKey] = struct{}{}
			journey.Milestones = append(journey.Milestones, JourneyMilestone{
				ID: "repository:" + repositoryKey, Kind: "repository_first",
				OccurredAt:  item.MergedAt,
				Title:       "First observed contribution to " + repositoryName,
				Description: "Earliest merged PR for this repository in the bounded sample.",
				EvidenceURL: item.URL, RepositoryName: repositoryName,
			})
		}
		technologyKey := strings.ToLower(item.Language)
		if technologyKey != "" {
			if _, seen := seenTechnologies[technologyKey]; !seen {
				seenTechnologies[technologyKey] = struct{}{}
				journey.Milestones = append(journey.Milestones, JourneyMilestone{
					ID: "technology:" + technologyKey, Kind: "technology_first",
					OccurredAt:  item.MergedAt,
					Title:       "First observed " + item.Language + " contribution",
					Description: "Earliest merged PR using this repository's primary language in the bounded sample.",
					EvidenceURL: item.URL, RepositoryName: repositoryName,
					Technology: item.Language,
				})
			}
		}
	}
	slices.SortFunc(journey.Milestones, func(left, right JourneyMilestone) int {
		if comparison := left.OccurredAt.Compare(right.OccurredAt); comparison != 0 {
			return comparison
		}
		return strings.Compare(left.ID, right.ID)
	})
	return journey
}

func analyzePortfolio(
	snapshot PortfolioSnapshot,
	analyzedAt time.Time,
) ContributionPortfolio {
	if !snapshot.Available {
		return ContributionPortfolio{
			Status:        EvidenceUnavailable,
			Languages:     []PortfolioLanguageCount{},
			Contributions: []PortfolioContribution{},
			AnalyzedAt:    analyzedAt,
		}
	}
	items := slices.Clone(snapshot.Items)
	slices.SortFunc(items, func(left, right PortfolioContribution) int {
		if comparison := right.MergedAt.Compare(left.MergedAt); comparison != 0 {
			return comparison
		}
		leftKey := strings.ToLower(left.RepositoryOwner + "/" + left.RepositoryName)
		rightKey := strings.ToLower(right.RepositoryOwner + "/" + right.RepositoryName)
		if comparison := strings.Compare(leftKey, rightKey); comparison != 0 {
			return comparison
		}
		return right.Number - left.Number
	})
	repositories := make(map[string]struct{})
	languages := make(map[string]int)
	for _, item := range items {
		repositories[strings.ToLower(item.RepositoryOwner+"/"+item.RepositoryName)] = struct{}{}
		if item.Language != "" {
			languages[item.Language]++
		}
	}
	languageCounts := make([]PortfolioLanguageCount, 0, len(languages))
	for name, count := range languages {
		languageCounts = append(languageCounts, PortfolioLanguageCount{Name: name, Count: count})
	}
	slices.SortFunc(languageCounts, func(left, right PortfolioLanguageCount) int {
		if left.Count != right.Count {
			return right.Count - left.Count
		}
		return strings.Compare(left.Name, right.Name)
	})
	status := EvidenceSampled
	if snapshot.Complete && !snapshot.HasMore {
		status = EvidenceExact
	}
	return ContributionPortfolio{
		Status:          status,
		TotalMerged:     max(snapshot.TotalMerged, len(items)),
		DisplayedMerged: len(items),
		RepositoryCount: len(repositories),
		HasMore:         snapshot.HasMore,
		AnalyzedAt:      analyzedAt,
		Languages:       languageCounts,
		Contributions:   items,
	}
}

func normalizeWindow(from, to time.Time) AnalysisWindow {
	to = to.UTC()
	if to.IsZero() {
		to = time.Now().UTC()
	}
	from = from.UTC()
	if from.IsZero() || !from.Before(to) {
		from = to.AddDate(0, 0, -AnalysisWindowDays)
	}
	return AnalysisWindow{
		From:       from,
		To:         to,
		Days:       int(math.Ceil(to.Sub(from).Hours() / 24)),
		PublicOnly: true,
	}
}

func analyzeContributions(snapshot ContributionSnapshot) ContributionAnalysis {
	if !snapshot.Available {
		unavailable := CountMetric{Status: EvidenceUnavailable}
		return ContributionAnalysis{
			WindowDays:          AnalysisWindowDays,
			Commits:             unavailable,
			IssuesOpened:        unavailable,
			PullRequestsOpened:  unavailable,
			PullRequestReviews:  unavailable,
			RepositoriesTouched: unavailable,
		}
	}
	return ContributionAnalysis{
		WindowDays:          AnalysisWindowDays,
		Commits:             countMetric(snapshot.Commits),
		IssuesOpened:        countMetric(snapshot.IssuesOpened),
		PullRequestsOpened:  countMetric(snapshot.PullRequestsOpened),
		PullRequestReviews:  countMetric(snapshot.PullRequestReviews),
		RepositoriesTouched: countMetric(snapshot.RepositoriesTouched),
	}
}

func countMetric(evidence CountEvidence) CountMetric {
	if !evidence.Available {
		return CountMetric{Status: EvidenceUnavailable}
	}
	status := EvidenceSampled
	if evidence.Complete {
		status = EvidenceExact
	}
	return CountMetric{Value: max(0, evidence.Value), Status: status}
}

func analyzeRepositoryCollection(
	collection RepositoryCollection,
	windowStart time.Time,
	precomputedLanguages []LanguageShare,
) RepositorySample {
	if !collection.Available {
		return RepositorySample{
			Status:              EvidenceUnavailable,
			PrimaryTechnologies: []LanguageShare{},
		}
	}
	status := EvidenceExact
	if collection.HasMore ||
		!collection.TotalKnown ||
		(collection.TotalKnown &&
			collection.Total > len(collection.Repositories)) {
		status = EvidenceSampled
	}
	var total *int
	if collection.TotalKnown {
		value := max(0, collection.Total)
		total = &value
	}
	active := 0
	for _, observation := range collection.Repositories {
		if !observation.Repository.IsArchived &&
			!repositoryLastUsed(observation.Repository).Before(windowStart) {
			active++
		}
	}
	return RepositorySample{
		Status:         status,
		Observed:       len(collection.Repositories),
		Total:          total,
		Limit:          collection.Limit,
		ActiveInWindow: active,
		PrimaryTechnologies: repositoryCollectionLanguages(
			collection,
			precomputedLanguages,
		),
	}
}

func repositoryCollectionLanguages(
	collection RepositoryCollection,
	precomputed []LanguageShare,
) []LanguageShare {
	if precomputed != nil {
		return slices.Clone(precomputed)
	}
	return topLanguageShares(
		AggregateLanguages(collectionLanguageBytes(collection)),
		maximumLanguageResults,
	)
}

func collectionLanguageBytes(
	collection RepositoryCollection,
) []map[string]int64 {
	result := make([]map[string]int64, 0, len(collection.Repositories))
	for _, observation := range collection.Repositories {
		if len(observation.Languages) > 0 {
			result = append(result, observation.Languages)
			continue
		}
		if language := strings.TrimSpace(
			observation.Repository.MainLanguage,
		); language != "" {
			result = append(result, map[string]int64{language: 1})
		}
	}
	return result
}

func topLanguageShares(
	languages []LanguageShare,
	limit int,
) []LanguageShare {
	if limit <= 0 || len(languages) == 0 {
		return []LanguageShare{}
	}
	if len(languages) <= limit {
		return slices.Clone(languages)
	}
	kept := slices.Clone(languages[:limit-1])
	other := 0
	for _, language := range languages[limit-1:] {
		other += language.Percentage
	}
	return append(kept, LanguageShare{Name: "Other", Percentage: other})
}

func repositoryLastUsed(item repository.Summary) time.Time {
	return maxTime(item.PushedAt, item.UpdatedAt).UTC()
}
