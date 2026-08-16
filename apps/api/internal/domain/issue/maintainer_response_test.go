package issue

import (
	"testing"
	"time"
)

func TestAssessMaintainerResponseUsesLatencyAndCoverage(t *testing.T) {
	t.Parallel()
	activity := ActivityMetrics{
		IssueResponse: SummarizeDurations(
			[]time.Duration{6 * time.Hour, 12 * time.Hour},
			180,
			false,
		),
		PullRequestReview: SummarizeDurations(
			[]time.Duration{48 * time.Hour},
			180,
			false,
		),
		PullRequestMergeTime: SummarizeDurations(
			[]time.Duration{60 * time.Hour},
			180,
			false,
		),
		UnansweredIssues: SummarizeCount(1, 4, 180, false),
	}

	got := AssessMaintainerResponse(activity)
	if got.Status != AggregateAvailable || got.Level != 4 ||
		got.Label != "Responsive" || got.SampleSize != 3 ||
		got.ResponseCoverage.Percentage != 75 ||
		got.PullRequestMerge.Median != 60*time.Hour {
		t.Fatalf("assessment = %+v", got)
	}
}

func TestAssessMaintainerResponsePenalizesLowResponseCoverage(t *testing.T) {
	t.Parallel()
	activity := ActivityMetrics{
		IssueResponse: SummarizeDurations(
			[]time.Duration{time.Hour},
			180,
			false,
		),
		UnansweredIssues: SummarizeCount(8, 10, 180, false),
	}
	got := AssessMaintainerResponse(activity)
	if got.Level != 1 || got.ResponseCoverage.Percentage != 20 {
		t.Fatalf("assessment = %+v", got)
	}
}

func TestAssessMaintainerResponseKeepsEmptySamplesUnavailable(t *testing.T) {
	t.Parallel()
	got := AssessMaintainerResponse(ActivityMetrics{
		PullRequestMergeTime: SummarizeDurations(
			[]time.Duration{48 * time.Hour},
			180,
			false,
		),
	})
	if got.Status != AggregateUnavailable || got.Level != 0 ||
		got.PullRequestMerge.Status != AggregateAvailable {
		t.Fatalf("assessment = %+v", got)
	}
}
