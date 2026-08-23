package account

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestIssueClaimWorkflowValidatesReferencesAndSubmittedStates(t *testing.T) {
	t.Parallel()
	claim, err := NewIssueClaim(" OpenAI ", "OpenAI-Go", 42)
	if err != nil {
		t.Fatalf("NewIssueClaim() error = %v", err)
	}
	if claim.Issue.RepositoryOwner != "openai" ||
		claim.Issue.RepositoryName != "openai-go" ||
		claim.Status != IssueClaimNotStarted {
		t.Fatalf("claim = %+v", claim)
	}
	if _, updateErr := UpdateIssueClaim(
		claim,
		IssueClaimPRSubmitted,
		false,
		nil,
	); !errors.Is(updateErr, ErrInvalidFeatureInput) {
		t.Fatalf("submitted without PR error = %v", updateErr)
	}
	pullRequest, err := NewPullRequestReference("OpenAI", "OpenAI-Go", 91)
	if err != nil {
		t.Fatalf("NewPullRequestReference() error = %v", err)
	}
	updated, err := UpdateIssueClaim(
		claim,
		IssueClaimPRSubmitted,
		false,
		&pullRequest,
	)
	if err != nil || updated.PullRequest == nil ||
		updated.PullRequest.RepositoryOwner != "openai" {
		t.Fatalf("UpdateIssueClaim() = %+v, %v", updated, err)
	}
}

func TestIssueClaimWorkflowSupportsEveryPersonalStateAndArchive(t *testing.T) {
	t.Parallel()
	claim, err := NewIssueClaim("octocat", "hello-world", 1)
	if err != nil {
		t.Fatalf("NewIssueClaim() error = %v", err)
	}
	pullRequest, _ := NewPullRequestReference("octocat", "hello-world", 2)
	for _, status := range []IssueClaimStatus{
		IssueClaimNotStarted,
		IssueClaimResearching,
		IssueClaimImplementing,
		IssueClaimPRSubmitted,
		IssueClaimMerged,
	} {
		var linked *PullRequestReference
		if status == IssueClaimPRSubmitted || status == IssueClaimMerged {
			linked = &pullRequest
		}
		claim, err = UpdateIssueClaim(claim, status, true, linked)
		if err != nil || claim.Status != status || !claim.Archived {
			t.Fatalf("status %q = %+v, %v", status, claim, err)
		}
	}
}

func TestBookmarkReferenceNormalizesAndValidatesTarget(t *testing.T) {
	number := 42
	reference, err := NewBookmarkReference(
		BookmarkTargetIssue,
		"OpenAI",
		"OpenAI-Go",
		&number,
	)
	if err != nil {
		t.Fatalf("NewBookmarkReference() error = %v", err)
	}
	if reference.RepositoryOwner != "openai" ||
		reference.RepositoryName != "openai-go" ||
		reference.IssueNumber == nil ||
		*reference.IssueNumber != 42 {
		t.Fatalf("reference = %+v", reference)
	}

	for _, test := range []struct {
		name       string
		targetType BookmarkTarget
		owner      string
		repository string
		number     *int
	}{
		{
			name:       "issue number missing",
			targetType: BookmarkTargetIssue,
			owner:      "openai",
			repository: "openai-go",
		},
		{
			name:       "repository includes issue number",
			targetType: BookmarkTargetRepository,
			owner:      "openai",
			repository: "openai-go",
			number:     &number,
		},
		{
			name:       "invalid owner",
			targetType: BookmarkTargetRepository,
			owner:      "../private",
			repository: "repo",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewBookmarkReference(
				test.targetType,
				test.owner,
				test.repository,
				test.number,
			)
			if !errors.Is(err, ErrInvalidFeatureInput) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestNormalizeSavedSearchFiltersUsesAnonymousDomainContracts(t *testing.T) {
	issueFilters, err := NormalizeSavedSearchFilters(
		SearchTypeIssue,
		json.RawMessage(`{"username":"octocat","languages":[" Go "]}`),
	)
	if err != nil {
		t.Fatalf("NormalizeSavedSearchFilters(issue) error = %v", err)
	}
	for _, fragment := range []string{
		`"username":"octocat"`,
		`"languages":["Go"]`,
		`"minimumStars":10`,
		`"labels":["good first issue","help wanted"]`,
	} {
		if !strings.Contains(string(issueFilters), fragment) {
			t.Errorf("issue filters missing %s: %s", fragment, issueFilters)
		}
	}

	repositoryFilters, err := NormalizeSavedSearchFilters(
		SearchTypeRepository,
		json.RawMessage(`{"licenses":["mit"],"forkPolicy":"include"}`),
	)
	if err != nil {
		t.Fatalf(
			"NormalizeSavedSearchFilters(repository) error = %v",
			err,
		)
	}
	for _, fragment := range []string{
		`"licenses":["MIT"]`,
		`"minimumStars":10`,
		`"forkPolicy":"include"`,
	} {
		if !strings.Contains(string(repositoryFilters), fragment) {
			t.Errorf(
				"repository filters missing %s: %s",
				fragment,
				repositoryFilters,
			)
		}
	}
}

func TestNormalizeSavedSearchFiltersRejectsInvalidDocuments(t *testing.T) {
	tests := []struct {
		name       string
		searchType SearchType
		raw        string
	}{
		{
			name:       "unknown field",
			searchType: SearchTypeIssue,
			raw:        `{"username":"octocat","private":true}`,
		},
		{
			name:       "invalid issue criteria",
			searchType: SearchTypeIssue,
			raw:        `{"username":""}`,
		},
		{
			name:       "invalid repository criteria",
			searchType: SearchTypeRepository,
			raw:        `{"maximumDifficulty":99}`,
		},
		{
			name:       "unknown search type",
			searchType: SearchType("profile"),
			raw:        `{}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NormalizeSavedSearchFilters(
				test.searchType,
				json.RawMessage(test.raw),
			)
			if !errors.Is(err, ErrInvalidFeatureInput) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestAccountFeatureValueObjectsRejectInvalidValues(t *testing.T) {
	if name, err := NormalizeSavedSearchName("  Go work  "); err != nil ||
		name != "Go work" {
		t.Fatalf("NormalizeSavedSearchName() = %q, %v", name, err)
	}
	if _, err := NormalizeSavedSearchName("\n"); !errors.Is(
		err,
		ErrInvalidFeatureInput,
	) {
		t.Fatalf("NormalizeSavedSearchName() error = %v", err)
	}
	if _, err := NewPreferences(
		ThemeDark,
		ReducedMotionReduce,
		50,
	); err != nil {
		t.Fatalf("NewPreferences() error = %v", err)
	}
	if _, err := NewPreferences(
		Theme("neon"),
		ReducedMotionSystem,
		20,
	); !errors.Is(err, ErrInvalidFeatureInput) {
		t.Fatalf("NewPreferences() error = %v", err)
	}
	if _, err := NewPage(1, 50); err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}
	if _, err := NewPage(0, 50); !errors.Is(err, ErrInvalidFeatureInput) {
		t.Fatalf("NewPage() error = %v", err)
	}
}

func TestResourceIDRoundTripAndRejectsAccountBoundaryMistakes(t *testing.T) {
	id, err := ParseResourceID("8bbfd7ed-a424-4ec3-a1b8-647006da1816")
	if err != nil {
		t.Fatalf("ParseResourceID() error = %v", err)
	}
	if id.String() != "8bbfd7ed-a424-4ec3-a1b8-647006da1816" {
		t.Fatalf("ResourceID.String() = %q", id.String())
	}
	if _, parseErr := ParseResourceID("not-a-resource"); !errors.Is(
		parseErr,
		ErrInvalidFeatureInput,
	) {
		t.Fatalf("ParseResourceID() error = %v", parseErr)
	}
	generated, err := NewResourceID()
	if err != nil || generated == (ResourceID{}) {
		t.Fatalf("NewResourceID() = %v, %v", generated, err)
	}
	page, err := NewPage(3, 20)
	if err != nil || page.Offset() != 40 {
		t.Fatalf("Page.Offset() = %d, %v", page.Offset(), err)
	}
}

func TestBookmarkMetadataAndSavedSearchKeysAreBounded(t *testing.T) {
	note, collection, tags, err := NormalizeBookmarkMetadata(
		"  Read CONTRIBUTING  ", " This week ", []string{"Go", "go", " beginner "},
	)
	if err != nil || note != "Read CONTRIBUTING" || collection != "This week" ||
		!slices.Equal(tags, []string{"Go", "beginner"}) {
		t.Fatalf("NormalizeBookmarkMetadata() = %q, %q, %v, %v", note, collection, tags, err)
	}
	for _, invalidTags := range [][]string{
		make([]string, MaximumBookmarkTags+1),
		{""},
		{strings.Repeat("x", MaximumBookmarkTagRunes+1)},
	} {
		if _, _, _, invalidErr := NormalizeBookmarkMetadata("", "", invalidTags); !errors.Is(invalidErr, ErrInvalidFeatureInput) {
			t.Fatalf("NormalizeBookmarkMetadata(%v) error = %v", invalidTags, invalidErr)
		}
	}
	keys, err := NormalizeSavedSearchResultKeys([]string{"Acme/Rocket#2", "acme/rocket#2", "acme/rocket#1"})
	if err != nil || !slices.Equal(keys, []string{"acme/rocket#1", "acme/rocket#2"}) {
		t.Fatalf("NormalizeSavedSearchResultKeys() = %v, %v", keys, err)
	}
	if _, err := NormalizeSavedSearchResultKeys(make([]string, MaximumSavedSearchResultKeys+1)); !errors.Is(err, ErrInvalidFeatureInput) {
		t.Fatalf("NormalizeSavedSearchResultKeys() error = %v", err)
	}
}
