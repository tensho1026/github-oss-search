package issue

import (
	"math"
	"strings"
	"time"
)

// RepositoryHealthScoreVersion pins category weights and thresholds.
const RepositoryHealthScoreVersion = "2026-08-01"

// OpenSSFCheck is one normalized Scorecard observation. A nil score represents
// upstream -1, an absent check, or an unsupported schema.
type OpenSSFCheck struct {
	Name  string
	Score *int
}

// OpenSSFSnapshot contains only bounded facts accepted from Scorecard.
type OpenSSFSnapshot struct {
	Available       bool
	AnalyzedAt      time.Time
	UpstreamVersion string
	Stale           bool
	Checks          []OpenSSFCheck
	Warning         string
}

// HealthComponent exposes one versioned weighted input.
type HealthComponent struct {
	Key         string
	Weight      int
	Score       *int
	Status      string
	Source      string
	Description string
}

// HealthCategory is one independent explainable repository dimension.
type HealthCategory struct {
	Name          string
	Score         *int
	Status        string
	Confidence    Confidence
	AnalyzedAt    time.Time
	SourceVersion string
	Components    []HealthComponent
	Warnings      []string
}

// RepositoryHealthDashboard keeps security separate from GitHub activity and
// community evidence to avoid double counting.
type RepositoryHealthDashboard struct {
	ScoreVersion string
	AnalyzedAt   time.Time
	Categories   []HealthCategory
}

// AnalyzeRepositoryHealth deterministically scores normalized public inputs.
func AnalyzeRepositoryHealth(
	signals []RepositorySignal,
	activity ActivityMetrics,
	openSSF OpenSSFSnapshot,
	now time.Time,
) RepositoryHealthDashboard {
	now = now.UTC()
	return RepositoryHealthDashboard{
		ScoreVersion: RepositoryHealthScoreVersion,
		AnalyzedAt:   now,
		Categories: []HealthCategory{
			buildHealthCategory("activity", now, []HealthComponent{
				healthComponent("recency", 50, recencyScore(activity.LastMeaningfulUpdate, now), "github", "Days since the latest meaningful update."),
				healthComponent("ci", 25, ciHealthScore(activity.CI), "github", "Latest default-branch CI state."),
				healthComponent("pull_request_merge", 25, ratioScore(activity.PullRequestMerge), "github", "Merged pull requests in a bounded public sample."),
			}, nil),
			buildHealthCategory("community", now, []HealthComponent{
				healthComponent("contributors", 35, countScore(activity.Contributors, 10), "github", "Distinct contributors in the bounded sample."),
				healthComponent("issue_response", 25, responseScore(activity.IssueResponse), "github", "Median maintainer first response time."),
				healthComponent("pull_request_review", 25, responseScore(activity.PullRequestReview), "github", "Median maintainer review time."),
				healthComponent("contributing_guide", 15, signalScore(signals, RepositoryContributing), "github", "Contribution guidance inspection."),
			}, nil),
			buildHealthCategory("beginner_friendly", now, []HealthComponent{
				healthComponent("contributing_guide", 30, signalScore(signals, RepositoryContributing), "github", "Contribution guidance inspection."),
				healthComponent("readme", 20, signalScore(signals, RepositoryREADME), "github", "README inspection."),
				healthComponent("code_of_conduct", 15, signalScore(signals, RepositoryCodeOfConduct), "github", "Code of conduct inspection."),
				healthComponent("tests", 15, signalScore(signals, RepositoryTests), "github", "Test configuration inspection."),
				healthComponent("ci", 10, ciHealthScore(activity.CI), "github", "Latest default-branch CI state."),
				healthComponent("issue_response", 10, responseScore(activity.IssueResponse), "github", "Median maintainer first response time."),
			}, nil),
			securityHealthCategory(openSSF, now),
		},
	}
}

func securityHealthCategory(snapshot OpenSSFSnapshot, now time.Time) HealthCategory {
	weights := []struct {
		key, check string
		weight     int
	}{
		{"maintained", "Maintained", 25},
		{"code_review", "Code-Review", 25},
		{"contributors", "Contributors", 15},
		{"ci_tests", "CI-Tests", 20},
		{"vulnerabilities", "Vulnerabilities", 15},
	}
	components := make([]HealthComponent, 0, len(weights))
	for _, definition := range weights {
		components = append(components, healthComponent(
			definition.key, definition.weight,
			openSSFScore(snapshot, definition.check), "openssf_scorecard",
			"Heuristic OpenSSF Scorecard check: "+definition.check+".",
		))
	}
	warnings := []string{"OpenSSF Scorecard observations are heuristics, not a security certification."}
	if !snapshot.Available {
		warnings = append(warnings, "OpenSSF Scorecard evidence is unavailable.")
	}
	if snapshot.Stale {
		warnings = append(warnings, "The published OpenSSF analysis is older than 30 days.")
	}
	if snapshot.Warning != "" {
		warnings = append(warnings, snapshot.Warning)
	}
	analyzedAt := snapshot.AnalyzedAt.UTC()
	if analyzedAt.IsZero() {
		analyzedAt = now
	}
	category := buildHealthCategory("security", analyzedAt, components, warnings)
	category.SourceVersion = snapshot.UpstreamVersion
	if !snapshot.Available {
		category.Score = nil
		category.Status = "unavailable"
		category.Confidence = ConfidenceLow
	} else if category.Score == nil {
		category.Warnings = append(category.Warnings, "Supported OpenSSF checks were unavailable or unknown.")
	} else if snapshot.Stale {
		category.Status = "partial"
		category.Confidence = ConfidenceMedium
	}
	return category
}

func buildHealthCategory(
	name string,
	analyzedAt time.Time,
	components []HealthComponent,
	warnings []string,
) HealthCategory {
	weighted, knownWeight := 0, 0
	for _, component := range components {
		if component.Score != nil {
			weighted += *component.Score * component.Weight
			knownWeight += component.Weight
		}
	}
	status := "unavailable"
	confidence := ConfidenceLow
	var score *int
	if knownWeight > 0 {
		value := int(math.Round(float64(weighted) / float64(knownWeight)))
		score = &value
		status = "partial"
		confidence = ConfidenceMedium
		if knownWeight == 100 {
			status = "available"
			confidence = ConfidenceHigh
		}
	}
	return HealthCategory{
		Name: name, Score: score, Status: status, Confidence: confidence,
		AnalyzedAt: analyzedAt, Components: components,
		Warnings: append([]string{}, warnings...),
	}
}

func healthComponent(
	key string, weight int, score *int, source, description string,
) HealthComponent {
	status := "unavailable"
	if score != nil {
		status = "available"
	}
	return HealthComponent{Key: key, Weight: weight, Score: score, Status: status, Source: source, Description: description}
}

func integerPointer(value int) *int { return &value }

func recencyScore(updatedAt, now time.Time) *int {
	if updatedAt.IsZero() {
		return nil
	}
	days := int(now.Sub(updatedAt.UTC()).Hours() / 24)
	switch {
	case days <= 30:
		return integerPointer(100)
	case days <= 90:
		return integerPointer(75)
	case days <= 180:
		return integerPointer(40)
	default:
		return integerPointer(0)
	}
}

func ciHealthScore(state CIState) *int {
	switch state {
	case CIStateSuccess:
		return integerPointer(100)
	case CIStatePending:
		return integerPointer(50)
	case CIStateFailure:
		return integerPointer(0)
	default:
		return nil
	}
}

func ratioScore(value RatioAggregate) *int {
	if value.Status != AggregateAvailable {
		return nil
	}
	return integerPointer(max(0, min(100, value.Percentage)))
}

func countScore(value CountAggregate, target int) *int {
	if value.Status != AggregateAvailable {
		return nil
	}
	return integerPointer(min(100, value.Value*100/target))
}

func responseScore(value DurationAggregate) *int {
	if value.Status != AggregateAvailable {
		return nil
	}
	switch {
	case value.Median <= 24*time.Hour:
		return integerPointer(100)
	case value.Median <= 3*24*time.Hour:
		return integerPointer(80)
	case value.Median <= 7*24*time.Hour:
		return integerPointer(60)
	case value.Median <= 14*24*time.Hour:
		return integerPointer(40)
	default:
		return integerPointer(10)
	}
}

func signalScore(signals []RepositorySignal, key RepositorySignalKey) *int {
	for _, signal := range signals {
		if signal.Key != key {
			continue
		}
		switch signal.State {
		case SignalPresent:
			return integerPointer(100)
		case SignalAbsent:
			return integerPointer(0)
		default:
			return nil
		}
	}
	return nil
}

func openSSFScore(snapshot OpenSSFSnapshot, name string) *int {
	if !snapshot.Available {
		return nil
	}
	for _, check := range snapshot.Checks {
		if strings.EqualFold(check.Name, name) && check.Score != nil {
			return integerPointer(max(0, min(100, *check.Score*10)))
		}
	}
	return nil
}
