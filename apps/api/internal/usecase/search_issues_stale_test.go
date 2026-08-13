package usecase

import (
	"testing"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
)

func TestFilterRankedIssuesByStaleRunsBeforePaginationInput(t *testing.T) {
	t.Parallel()
	criteria, err := issue.NewSearchCriteria(issue.SearchCriteriaOptions{
		Username: "octocat",
	})
	if err != nil {
		t.Fatalf("NewSearchCriteria() error = %v", err)
	}
	ranked := []issue.RankedIssue{
		staleRankedIssue(1, issue.StaleFresh),
		staleRankedIssue(2, issue.StaleStale),
		staleRankedIssue(3, issue.StaleUnknown),
	}

	got, excluded := filterRankedIssuesByStale(ranked, criteria)

	if excluded != 1 || len(got) != 2 ||
		got[0].Candidate.Issue.Number != 1 ||
		got[1].Candidate.Issue.Number != 3 {
		t.Fatalf("filterRankedIssuesByStale() = %+v, %d", got, excluded)
	}
}

func TestFilterRankedIssuesByStaleCanExplicitlyIncludeStale(t *testing.T) {
	t.Parallel()
	includeStale := true
	criteria, err := issue.NewSearchCriteria(issue.SearchCriteriaOptions{
		Username:     "octocat",
		IncludeStale: &includeStale,
	})
	if err != nil {
		t.Fatalf("NewSearchCriteria() error = %v", err)
	}
	ranked := []issue.RankedIssue{staleRankedIssue(1, issue.StaleStale)}

	got, excluded := filterRankedIssuesByStale(ranked, criteria)

	if excluded != 0 || len(got) != 1 {
		t.Fatalf("filterRankedIssuesByStale() = %+v, %d", got, excluded)
	}
}

func staleRankedIssue(number int, state issue.StaleState) issue.RankedIssue {
	return issue.RankedIssue{
		Candidate: issue.Candidate{Issue: issue.Summary{Number: number}},
		Recommendation: issue.Recommendation{
			Stale: issue.StaleAssessment{State: state},
		},
	}
}
