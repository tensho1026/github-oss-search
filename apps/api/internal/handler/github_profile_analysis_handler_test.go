package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/profile"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/user"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/apperror"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
	"github.com/tensho1026/github-issue-search/apps/api/internal/transport/response"
	"github.com/tensho1026/github-issue-search/apps/api/internal/usecase"
)

func TestGitHubProfileAnalysisHandlerGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	analyze := &analyzeGitHubProfileStub{
		output: usecase.AnalyzeGitHubProfileOutput{
			Analysis: profile.Analysis{
				Username: "octocat",
				Languages: []profile.LanguageShare{
					{Name: "Go", Percentage: 60},
					{Name: "TypeScript", Percentage: 40},
				},
				LanguageStatus: profile.EvidenceSampled,
				Frameworks:     []string{"Gin", "React"},
				RecentTechnologies: []profile.RecentTechnology{{
					Name:            "Go",
					Kind:            profile.TechnologyLanguage,
					LastUsedAt:      now.Add(-time.Hour),
					RepositoryCount: 2,
					RepositorySources: []profile.RepositorySource{
						profile.RepositoryContributed,
						profile.RepositoryOwned,
					},
					Confidence: profile.ConfidenceMedium,
				}},
				Contributions: profile.ContributionAnalysis{
					WindowDays: profile.AnalysisWindowDays,
					Commits: profile.CountMetric{
						Value:  10,
						Status: profile.EvidenceSampled,
					},
					IssuesOpened: profile.CountMetric{
						Value:  2,
						Status: profile.EvidenceExact,
					},
					PullRequestsOpened: profile.CountMetric{
						Value:  5,
						Status: profile.EvidenceExact,
					},
					PullRequestReviews: profile.CountMetric{
						Value:  3,
						Status: profile.EvidenceSampled,
					},
					RepositoriesTouched: profile.CountMetric{
						Value:  2,
						Status: profile.EvidenceSampled,
					},
				},
				ContributionCalendar: profile.ContributionCalendar{
					Status: profile.EvidenceExact,
					Total:  2,
					From:   time.Date(2026, time.July, 26, 0, 0, 0, 0, time.UTC),
					To:     time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
					Weeks: []profile.ContributionWeek{{
						Index:    0,
						FirstDay: time.Date(2026, time.July, 26, 0, 0, 0, 0, time.UTC),
						Days: []profile.ContributionDay{
							{Date: time.Date(2026, time.July, 26, 0, 0, 0, 0, time.UTC), Weekday: 0, Level: profile.ContributionNone},
							{Date: time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC), Weekday: 1, Count: 2, Level: profile.ContributionSecond},
						},
					}},
				},
				Portfolio: profile.ContributionPortfolio{
					Status:          profile.EvidenceSampled,
					TotalMerged:     8,
					DisplayedMerged: 1,
					RepositoryCount: 1,
					HasMore:         true,
					AnalyzedAt:      now,
					Languages: []profile.PortfolioLanguageCount{{
						Name: "Go", Count: 1,
					}},
					Contributions: []profile.PortfolioContribution{{
						RepositoryOwner: "community",
						RepositoryName:  "project",
						Number:          42,
						Title:           "Add bounded retries",
						URL:             "https://github.com/community/project/pull/42",
						MergedAt:        now.Add(-time.Hour),
						Language:        "Go",
					}},
				},
				Journey: profile.OSSJourney{
					Status: profile.EvidenceSampled, AnalyzedAt: now,
					Milestones: []profile.JourneyMilestone{{
						ID:   "merged:community/project#42",
						Kind: "merged_pull_request", OccurredAt: now.Add(-time.Hour),
						Title:          "Merged PR #42 in community/project",
						Description:    "Observed public merge: Add bounded retries",
						EvidenceURL:    "https://github.com/community/project/pull/42",
						RepositoryName: "community/project",
					}},
				},
				Streak: profile.ContributionStreak{
					Status: profile.EvidenceSampled, AnalyzedAt: now,
					Timezone: "UTC", WeekStartsOn: "monday",
					CurrentWeeks: 1, LongestWeeks: 1, QualifyingWeeks: 1,
					Weeks: []profile.StreakWeek{{
						StartedAt:    time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
						EndedAt:      time.Date(2026, time.August, 2, 23, 59, 59, 0, time.UTC),
						EventCount:   1,
						EvidenceURLs: []string{"https://github.com/community/project/pull/42"},
					}},
				},
				Quest: profile.OSSQuest{
					CatalogVersion: "2026-08-01", EvaluatedAt: now,
					Completed: 1, Total: 5, NextQuestID: "first_review",
					Items: []profile.QuestProgress{{
						ID: "first_pr", Title: "Open your first pull request",
						Description: "Create a public OSS pull request.",
						Status:      "completed", Current: 1, Target: 1,
						NextAction: "Find a matching issue and open a focused PR.",
					}},
				},
				OSSExperience: profile.OSSExperience{
					Level:      "active",
					Confidence: profile.ConfidenceHigh,
					PublicOnly: true,
					Evidence: []profile.TechnologyEvidence{{
						Kind:   "authored_pull_requests",
						Value:  5,
						Status: profile.EvidenceExact,
					}},
				},
				RepositoryEvidence: profile.RepositoryEvidence{
					Owned: profile.RepositorySample{
						Status:         profile.EvidenceSampled,
						Observed:       2,
						Total:          profileTotal(5),
						Limit:          20,
						ActiveInWindow: 2,
						PrimaryTechnologies: []profile.LanguageShare{{
							Name:       "Go",
							Percentage: 60,
						}},
					},
					Contributed: profile.RepositorySample{
						Status:         profile.EvidenceExact,
						Observed:       2,
						Total:          profileTotal(2),
						Limit:          20,
						ActiveInWindow: 2,
					},
					Starred: profile.RepositorySample{
						Status:   profile.EvidenceSampled,
						Observed: 1,
						Limit:    20,
					},
					Forked: profile.RepositorySample{
						Status:   profile.EvidenceExact,
						Observed: 0,
						Total:    profileTotal(0),
						Limit:    20,
					},
				},
				Proficiency: []profile.TechnologyProficiency{{
					Name:       "Go",
					Kind:       profile.TechnologyLanguage,
					Level:      4,
					Label:      "advanced",
					Score:      65,
					Confidence: profile.ConfidenceMedium,
					Evidence: []profile.TechnologyEvidence{{
						Kind:   "owned_repositories",
						Value:  2,
						Status: profile.EvidenceSampled,
					}},
				}},
				Window: profile.AnalysisWindow{
					From:       now.AddDate(0, 0, -profile.AnalysisWindowDays),
					To:         now,
					Days:       profile.AnalysisWindowDays,
					PublicOnly: true,
				},
				RepositoriesAnalyzed: 2,
				Warnings: []profile.Warning{{
					Code:       "manifest_data_unavailable",
					Message:    "A framework manifest could not be retrieved",
					Repository: "octocat/private",
				}},
			},
			RateLimit: port.RateLimit{Known: true, Remaining: 37},
			CacheHit:  true,
		},
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "username", Value: "octocat"}}
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/api/github/users/octocat/profile-analysis",
		nil,
	)

	NewGitHubProfileAnalysisHandler(
		analyze,
		response.NewResponder(),
	).Get(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	for _, value := range []string{
		`"username":"octocat"`,
		`"name":"Go","percentage":60`,
		`"languageStatus":"sampled"`,
		`"frameworks":["Gin","React"]`,
		`"recentTechnologies":[{"name":"Go","kind":"language"`,
		`"repositorySources":["contributed","owned"]`,
		`"commits":{"value":10,"status":"sampled"}`,
		`"pullRequestsOpened":{"value":5,"status":"exact"}`,
		`"contributionCalendar":{"status":"exact","total":2,"from":"2026-07-26","to":"2026-07-27","weeks":[{"index":0,"firstDay":"2026-07-26"`,
		`"date":"2026-07-27","weekday":1,"count":2,"level":"second_quartile"`,
		`"contributionPortfolio":{"status":"sampled","totalMerged":8,"displayedMerged":1,"repositoryCount":1,"hasMore":true`,
		`"name":"Go","count":1`,
		`"number":42,"title":"Add bounded retries","url":"https://github.com/community/project/pull/42"`,
		`"summary":"Merged public pull request in community/project: Add bounded retries"`,
		`"ossJourney":{"status":"sampled","analyzedAt":"2026-07-30T12:00:00Z"`,
		`"id":"merged:community/project#42","kind":"merged_pull_request"`,
		`"evidenceUrl":"https://github.com/community/project/pull/42"`,
		`"contributionStreak":{"status":"sampled","analyzedAt":"2026-07-30T12:00:00Z","timezone":"UTC","weekStartsOn":"monday","currentWeeks":1,"longestWeeks":1,"qualifyingWeeks":1`,
		`"eventCount":1,"evidenceUrls":["https://github.com/community/project/pull/42"]`,
		`"ossQuest":{"catalogVersion":"2026-08-01","evaluatedAt":"2026-07-30T12:00:00Z","completed":1,"total":5,"nextQuestId":"first_review"`,
		`"id":"first_pr","title":"Open your first pull request","description":"Create a public OSS pull request.","status":"completed","current":1,"target":1,"completedAt":null`,
		`"ossExperience":{"level":"active","confidence":"high","publicOnly":true`,
		`"owned":{"status":"sampled","observed":2,"total":5`,
		`"starred":{"status":"sampled","observed":1,"total":null`,
		`"proficiency":[{"name":"Go","kind":"language","level":4,"label":"advanced"`,
		`"analysisWindow":{"from":"2025-07-30T12:00:00Z","to":"2026-07-30T12:00:00Z","days":365,"publicOnly":true}`,
		`"repositoriesAnalyzed":2`,
		`"code":"manifest_data_unavailable"`,
		`"repository":"octocat/private"`,
		`"rateLimitRemaining":37`,
	} {
		if !strings.Contains(recorder.Body.String(), value) {
			t.Errorf("body missing %s: %s", value, recorder.Body.String())
		}
	}
	if analyze.username != "octocat" {
		t.Fatalf("usecase username = %q", analyze.username)
	}
	if recorder.Header().Get(issueSearchCacheHeader) != issueSearchCacheHit {
		t.Fatalf(
			"%s = %q",
			issueSearchCacheHeader,
			recorder.Header().Get(issueSearchCacheHeader),
		)
	}
}

func TestGitHubProfileAnalysisHandlerRejectsInvalidUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	analyze := &analyzeGitHubProfileStub{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "username", Value: "invalid--username"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	NewGitHubProfileAnalysisHandler(
		analyze,
		response.NewResponder(),
	).Get(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	if analyze.called {
		t.Fatal("usecase was called for invalid input")
	}
	if !strings.Contains(recorder.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGitHubProfileAnalysisHandlerWritesUsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	analyze := &analyzeGitHubProfileStub{
		err: apperror.New(
			apperror.CodeRateLimit,
			"GitHub API rate limit was exceeded",
			http.StatusTooManyRequests,
		),
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "username", Value: "octocat"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	NewGitHubProfileAnalysisHandler(
		analyze,
		response.NewResponder(),
	).Get(ctx)

	if recorder.Code != http.StatusTooManyRequests ||
		!strings.Contains(recorder.Body.String(), "GITHUB_RATE_LIMIT_EXCEEDED") {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}

type analyzeGitHubProfileStub struct {
	output   usecase.AnalyzeGitHubProfileOutput
	err      error
	username user.Username
	called   bool
}

func (stub *analyzeGitHubProfileStub) Execute(
	_ context.Context,
	username user.Username,
) (usecase.AnalyzeGitHubProfileOutput, error) {
	stub.called = true
	stub.username = username
	return stub.output, stub.err
}

var _ usecase.AnalyzeGitHubProfile = (*analyzeGitHubProfileStub)(nil)

func profileTotal(value int) *int {
	return &value
}
