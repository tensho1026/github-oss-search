package issue

import (
	"strings"
	"time"
)

// Versioned stale-policy constants keep classification deterministic and
// observable across API clients.
const (
	StalePolicyVersion   = "stale-v1"
	StaleFreshWithinDays = 30
	StaleAfterDays       = 180
)

// AssessStaleness classifies one issue from normalized, bounded human,
// maintainer, linked-PR, and repository evidence.
func AssessStaleness(input RecommendationInput) StaleAssessment {
	now := input.Now.UTC()
	if now.IsZero() {
		now = input.Candidate.Issue.UpdatedAt.UTC()
	}
	assessment := StaleAssessment{
		State:                StaleUnknown,
		PolicyVersion:        StalePolicyVersion,
		Confidence:           ConfidenceLow,
		AnalyzedAt:           now,
		FreshWithinDays:      StaleFreshWithinDays,
		StaleAfterDays:       StaleAfterDays,
		IssueCreatedAt:       input.Candidate.Issue.CreatedAt.UTC(),
		IssueUpdatedAt:       input.Candidate.Issue.UpdatedAt.UTC(),
		RepositoryActivityAt: input.Activity.LastMeaningfulUpdate.UTC(),
		Truncated: input.History.CommentsTruncated ||
			input.History.LinkedPullRequestsTruncated,
		Evidence: []Evidence{{
			RuleID:      "stale.policy.version",
			Source:      EvidenceDerived,
			Description: "stale-v1 uses bounded human, maintainer, linked pull request, and repository activity",
		}},
	}
	if assessment.RepositoryActivityAt.IsZero() {
		assessment.RepositoryActivityAt = latestTime(
			input.Candidate.Repository.PushedAt,
			input.Candidate.Repository.UpdatedAt,
		).UTC()
	}
	if now.IsZero() || assessment.IssueCreatedAt.IsZero() ||
		assessment.IssueUpdatedAt.IsZero() ||
		assessment.IssueCreatedAt.After(now) {
		assessment.Evidence = append(assessment.Evidence, Evidence{
			RuleID:      "stale.issue.dates.unavailable",
			Source:      EvidenceIssueMetadata,
			Description: "required issue dates were unavailable or invalid",
		})
		return assessment
	}

	lastHuman := assessment.IssueCreatedAt
	comments := input.History.Comments
	if len(comments) > MaximumClaimComments {
		comments = comments[:MaximumClaimComments]
		assessment.Truncated = true
	}
	for _, comment := range comments {
		if comment.CreatedAt.IsZero() || comment.CreatedAt.After(now) {
			continue
		}
		assessment.SampleSize++
		if IsBotIdentity(comment.AuthorLogin, comment.AuthorType) {
			continue
		}
		lastHuman = latestTime(lastHuman, comment.CreatedAt)
		if isMaintainerAssociation(comment.AuthorAssociation) {
			assessment.LastMaintainerActivityAt = latestTime(
				assessment.LastMaintainerActivityAt,
				comment.CreatedAt,
			)
		}
	}

	linked := input.History.LinkedPullRequests
	if len(linked) > MaximumMetricSamples {
		linked = linked[:MaximumMetricSamples]
		assessment.Truncated = true
	}
	linkedOpen := false
	linkedMerged := false
	for _, pullRequest := range linked {
		if pullRequest.Number < 1 ||
			pullRequest.UpdatedAt.IsZero() || pullRequest.UpdatedAt.After(now) {
			continue
		}
		assessment.SampleSize++
		if pullRequest.IsDraft {
			continue
		}
		assessment.LastLinkedPullRequestAt = latestTime(
			assessment.LastLinkedPullRequestAt,
			pullRequest.UpdatedAt,
		)
		lastHuman = latestTime(lastHuman, pullRequest.UpdatedAt)
		linkedOpen = linkedOpen || strings.EqualFold(pullRequest.State, StateOpen)
		linkedMerged = linkedMerged || !pullRequest.MergedAt.IsZero()
	}
	assessment.LastMeaningfulIssueActivityAt = lastHuman.UTC()

	if input.Candidate.Repository.IsArchived {
		return staleResult(assessment, StaleStale, ConfidenceHigh,
			"stale.repository.archived",
			"the repository is archived and no longer accepts normal contributions")
	}
	if linkedMerged {
		return staleResult(assessment, StaleStale, ConfidenceHigh,
			"stale.pull_request.merged",
			"a bounded closing pull request reference is already merged")
	}
	if recentWithin(now, assessment.LastMaintainerActivityAt, StaleFreshWithinDays) {
		return staleResult(assessment, StaleFresh, ConfidenceHigh,
			"stale.maintainer.recent",
			"recent maintainer activity was observed on the issue")
	}
	if linkedOpen && recentWithin(
		now,
		assessment.LastLinkedPullRequestAt,
		StaleFreshWithinDays,
	) {
		return staleResult(assessment, StaleFresh, ConfidenceHigh,
			"stale.pull_request.active",
			"a recently updated linked open pull request was observed")
	}
	if recentWithin(now, assessment.IssueCreatedAt, StaleFreshWithinDays) {
		return staleResult(assessment, StaleFresh, ConfidenceMedium,
			"stale.issue.recently_created",
			"the issue was created within the fresh window")
	}
	if input.Claim.Claimed || len(input.Candidate.Issue.Assignees) > 0 || linkedOpen {
		return staleResult(assessment, StaleAging, ConfidenceMedium,
			"stale.issue.work_observed",
			"assignment, claim, or linked open pull request evidence indicates work in progress")
	}
	if assessment.Truncated {
		return staleResult(assessment, StaleUnknown, ConfidenceLow,
			"stale.history.truncated",
			"bounded issue history was incomplete, so absence cannot prove staleness")
	}
	if olderThan(now, assessment.LastMeaningfulIssueActivityAt, StaleAfterDays) {
		return staleResult(assessment, StaleStale, ConfidenceHigh,
			"stale.issue.inactive",
			"no meaningful human, maintainer, or linked pull request activity was observed within 180 days")
	}
	return staleResult(assessment, StaleAging, ConfidenceMedium,
		"stale.issue.aging",
		"the issue is outside the fresh window but has not met the stale threshold")
}

func staleResult(
	assessment StaleAssessment,
	state StaleState,
	confidence Confidence,
	ruleID string,
	description string,
) StaleAssessment {
	assessment.State = state
	assessment.Confidence = confidence
	assessment.Evidence = append(assessment.Evidence, Evidence{
		RuleID:      ruleID,
		Source:      EvidenceDerived,
		Description: description,
	})
	return assessment
}

func recentWithin(now, observed time.Time, days int) bool {
	return !observed.IsZero() && !observed.After(now) &&
		now.Sub(observed) <= time.Duration(days)*24*time.Hour
}

func olderThan(now, observed time.Time, days int) bool {
	return !observed.IsZero() && !observed.After(now) &&
		now.Sub(observed) > time.Duration(days)*24*time.Hour
}

func latestTime(values ...time.Time) time.Time {
	var latest time.Time
	for _, value := range values {
		if value.After(latest) {
			latest = value
		}
	}
	return latest
}

func isMaintainerAssociation(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "OWNER", "MEMBER", "COLLABORATOR":
		return true
	default:
		return false
	}
}
