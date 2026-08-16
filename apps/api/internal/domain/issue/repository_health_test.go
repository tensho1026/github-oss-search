package issue

import (
	"testing"
	"time"
)

func TestAnalyzeRepositoryHealthIsVersionedAndExplainable(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC)
	dashboard := AnalyzeRepositoryHealth(
		[]RepositorySignal{
			{Key: RepositoryREADME, State: SignalPresent},
			{Key: RepositoryContributing, State: SignalPresent},
			{Key: RepositoryCI, State: SignalPresent},
			{Key: RepositoryTests, State: SignalPresent},
			{Key: RepositoryCodeOfConduct, State: SignalAbsent},
		},
		ActivityMetrics{
			LastMeaningfulUpdate: now.AddDate(0, 0, -10), CI: CIStateSuccess,
			Contributors:      CountAggregate{Status: AggregateAvailable, Value: 8},
			PullRequestMerge:  RatioAggregate{Status: AggregateAvailable, Percentage: 80},
			IssueResponse:     DurationAggregate{Status: AggregateAvailable, Median: 6 * time.Hour},
			PullRequestReview: DurationAggregate{Status: AggregateAvailable, Median: 2 * 24 * time.Hour},
		},
		OpenSSFSnapshot{
			Available: true, AnalyzedAt: now.Add(-time.Hour), UpstreamVersion: "v5.2.1",
			Checks: []OpenSSFCheck{
				{Name: "Maintained", Score: integerPointer(10)},
				{Name: "Code-Review", Score: integerPointer(8)},
				{Name: "Contributors", Score: integerPointer(7)},
				{Name: "CI-Tests", Score: integerPointer(9)},
				{Name: "Vulnerabilities", Score: integerPointer(10)},
			},
		},
		now,
	)
	if dashboard.ScoreVersion != RepositoryHealthScoreVersion ||
		len(dashboard.Categories) != 4 {
		t.Fatalf("dashboard = %+v", dashboard)
	}
	for _, category := range dashboard.Categories {
		if category.Score == nil || category.Status != "available" ||
			category.Confidence != ConfidenceHigh {
			t.Fatalf("category = %+v", category)
		}
	}
	security := dashboard.Categories[3]
	if *security.Score != 89 || len(security.Components) != 5 ||
		len(security.Warnings) == 0 {
		t.Fatalf("security = %+v", security)
	}
}

func TestAnalyzeRepositoryHealthPreservesMissingAndUnknownEvidence(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC)
	unknown := -1
	dashboard := AnalyzeRepositoryHealth(nil, ActivityMetrics{}, OpenSSFSnapshot{
		Available: true, Stale: true,
		Checks: []OpenSSFCheck{{Name: "Maintained", Score: nil}, {Name: "ignored", Score: &unknown}},
	}, now)
	for index, category := range dashboard.Categories {
		if category.Score != nil || category.Status != "unavailable" {
			t.Fatalf("category %d = %+v", index, category)
		}
	}
	if len(dashboard.Categories[3].Warnings) < 2 {
		t.Fatalf("security warnings = %+v", dashboard.Categories[3].Warnings)
	}
}
