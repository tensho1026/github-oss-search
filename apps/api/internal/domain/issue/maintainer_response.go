package issue

import "time"

// AssessMaintainerResponse derives one five-level historical response score
// from bounded maintainer-only issue and pull-request observations.
func AssessMaintainerResponse(
	activity ActivityMetrics,
) MaintainerResponseAssessment {
	issueResponse := normalizedDurationAggregate(activity.IssueResponse)
	pullReview := normalizedDurationAggregate(activity.PullRequestReview)
	mergeTime := normalizedDurationAggregate(activity.PullRequestMergeTime)
	coverage := responseCoverage(activity.UnansweredIssues)

	weightedLevel := 0
	sampleSize := 0
	windowDays := 0
	truncated := false
	for _, metric := range []DurationAggregate{issueResponse, pullReview} {
		if metric.Status != AggregateAvailable || metric.SampleSize < 1 {
			continue
		}
		weightedLevel += responseLevel(metric.Median) * metric.SampleSize
		sampleSize += metric.SampleSize
		windowDays = max(windowDays, metric.WindowDays)
		truncated = truncated || metric.Truncated
	}
	if sampleSize == 0 {
		return MaintainerResponseAssessment{
			Status:             AggregateUnavailable,
			Confidence:         ConfidenceLow,
			WindowDays:         max(windowDays, mergeTime.WindowDays),
			ResponseCoverage:   coverage,
			FirstIssueResponse: issueResponse,
			FirstPullReview:    pullReview,
			PullRequestMerge:   mergeTime,
		}
	}

	level := clamp(weightedLevel/sampleSize, 1, 5)
	if coverage.Status == AggregateAvailable {
		switch {
		case coverage.Percentage < 25:
			level = min(level, 1)
		case coverage.Percentage < 50:
			level = min(level, 2)
		case coverage.Percentage < 75:
			level = min(level, 3)
		}
	}
	return MaintainerResponseAssessment{
		Status:             AggregateAvailable,
		Level:              level,
		Label:              maintainerResponseLabel(level),
		Confidence:         sampleConfidence(sampleSize, truncated),
		SampleSize:         sampleSize,
		WindowDays:         windowDays,
		ResponseCoverage:   coverage,
		FirstIssueResponse: issueResponse,
		FirstPullReview:    pullReview,
		PullRequestMerge:   mergeTime,
	}
}

func normalizedDurationAggregate(
	metric DurationAggregate,
) DurationAggregate {
	if metric.Status != AggregateAvailable || metric.SampleSize < 1 ||
		metric.Median <= 0 {
		return DurationAggregate{
			Status:     AggregateUnavailable,
			WindowDays: metric.WindowDays,
			Truncated:  metric.Truncated,
			Confidence: ConfidenceLow,
		}
	}
	return metric
}

func responseCoverage(unanswered CountAggregate) RatioAggregate {
	if unanswered.Status != AggregateAvailable || unanswered.SampleSize < 1 ||
		unanswered.Value < 0 || unanswered.Value > unanswered.SampleSize {
		return RatioAggregate{
			Status:     AggregateUnavailable,
			WindowDays: unanswered.WindowDays,
			Truncated:  unanswered.Truncated,
			Confidence: ConfidenceLow,
		}
	}
	return SummarizeRatio(
		unanswered.SampleSize-unanswered.Value,
		unanswered.SampleSize,
		unanswered.WindowDays,
		unanswered.Truncated,
	)
}

func responseLevel(duration time.Duration) int {
	switch {
	case duration <= 24*time.Hour:
		return 5
	case duration <= 72*time.Hour:
		return 4
	case duration <= 7*24*time.Hour:
		return 3
	case duration <= 14*24*time.Hour:
		return 2
	default:
		return 1
	}
}

func maintainerResponseLabel(level int) string {
	switch level {
	case 5:
		return "Very responsive"
	case 4:
		return "Responsive"
	case 3:
		return "Moderate"
	case 2:
		return "Slow"
	default:
		return "Very slow"
	}
}
