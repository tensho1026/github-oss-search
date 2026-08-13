package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/profile"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/user"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/apperror"
	"github.com/tensho1026/github-issue-search/apps/api/internal/transport/response"
	"github.com/tensho1026/github-issue-search/apps/api/internal/usecase"
)

// GitHubProfileAnalysisHandler validates profile path input and presents
// public, explainable analysis through the shared response contract.
type GitHubProfileAnalysisHandler struct {
	analyze   usecase.AnalyzeGitHubProfile
	responder response.Responder
}

// NewGitHubProfileAnalysisHandler binds a profile-analysis use case to a
// responder. Dependencies must be fully composed by the router.
func NewGitHubProfileAnalysisHandler(
	analyze usecase.AnalyzeGitHubProfile,
	responder response.Responder,
) GitHubProfileAnalysisHandler {
	return GitHubProfileAnalysisHandler{
		analyze:   analyze,
		responder: responder,
	}
}

// Get handles one cancellable public profile-analysis request.
func (h GitHubProfileAnalysisHandler) Get(ctx *gin.Context) {
	username, err := user.ParseUsername(ctx.Param("username"))
	if err != nil {
		h.responder.Error(ctx, apperror.Wrap(
			apperror.CodeInvalidRequest,
			"GitHub username is invalid",
			http.StatusBadRequest,
			err,
		))
		return
	}

	output, err := h.analyze.Execute(ctx.Request.Context(), username)
	if err != nil {
		h.responder.Error(ctx, err)
		return
	}

	writeCacheStatus(ctx, output.CacheHit)
	var remaining *int
	if output.RateLimit.Known {
		remaining = &output.RateLimit.Remaining
	}
	h.responder.DataWithMeta(
		ctx,
		http.StatusOK,
		newGitHubProfileAnalysisResponse(output.Analysis),
		response.MetaOptions{RateLimitRemaining: remaining},
	)
}

type githubProfileAnalysisResponse struct {
	Username             string                          `json:"username"`
	Languages            []languageShareResponse         `json:"languages"`
	LanguageStatus       string                          `json:"languageStatus"`
	Frameworks           []string                        `json:"frameworks"`
	RecentTechnologies   []recentTechnologyResponse      `json:"recentTechnologies"`
	Contributions        contributionAnalysisResponse    `json:"contributions"`
	ContributionCalendar contributionCalendarResponse    `json:"contributionCalendar"`
	Portfolio            contributionPortfolioResponse   `json:"contributionPortfolio"`
	Journey              ossJourneyResponse              `json:"ossJourney"`
	OSSExperience        ossExperienceResponse           `json:"ossExperience"`
	RepositoryEvidence   repositoryEvidenceResponse      `json:"repositoryEvidence"`
	Proficiency          []technologyProficiencyResponse `json:"proficiency"`
	AnalysisWindow       analysisWindowResponse          `json:"analysisWindow"`
	RepositoriesAnalyzed int                             `json:"repositoriesAnalyzed"`
	Warnings             []profileWarningResponse        `json:"warnings"`
}

type languageShareResponse struct {
	Name       string `json:"name"`
	Percentage int    `json:"percentage"`
}

type profileWarningResponse struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Repository string `json:"repository,omitempty"`
}

type recentTechnologyResponse struct {
	Name              string   `json:"name"`
	Kind              string   `json:"kind"`
	LastUsedAt        string   `json:"lastUsedAt"`
	RepositoryCount   int      `json:"repositoryCount"`
	RepositorySources []string `json:"repositorySources"`
	Confidence        string   `json:"confidence"`
}

type countMetricResponse struct {
	Value  int    `json:"value"`
	Status string `json:"status"`
}

type contributionAnalysisResponse struct {
	WindowDays          int                 `json:"windowDays"`
	Commits             countMetricResponse `json:"commits"`
	IssuesOpened        countMetricResponse `json:"issuesOpened"`
	PullRequestsOpened  countMetricResponse `json:"pullRequestsOpened"`
	PullRequestReviews  countMetricResponse `json:"pullRequestReviews"`
	RepositoriesTouched countMetricResponse `json:"repositoriesTouched"`
}

type contributionCalendarResponse struct {
	Status string                     `json:"status"`
	Total  int                        `json:"total"`
	From   string                     `json:"from,omitempty"`
	To     string                     `json:"to,omitempty"`
	Weeks  []contributionWeekResponse `json:"weeks"`
}

type contributionWeekResponse struct {
	Index    int                       `json:"index"`
	FirstDay string                    `json:"firstDay"`
	Days     []contributionDayResponse `json:"days"`
}

type contributionDayResponse struct {
	Date    string `json:"date"`
	Weekday int    `json:"weekday"`
	Count   int    `json:"count"`
	Level   string `json:"level"`
}

type contributionPortfolioResponse struct {
	Status          string                          `json:"status"`
	TotalMerged     int                             `json:"totalMerged"`
	DisplayedMerged int                             `json:"displayedMerged"`
	RepositoryCount int                             `json:"repositoryCount"`
	HasMore         bool                            `json:"hasMore"`
	AnalyzedAt      string                          `json:"analyzedAt"`
	Languages       []portfolioLanguageResponse     `json:"languages"`
	Contributions   []portfolioContributionResponse `json:"contributions"`
}

type portfolioLanguageResponse struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type portfolioContributionResponse struct {
	RepositoryOwner string `json:"repositoryOwner"`
	RepositoryName  string `json:"repositoryName"`
	Number          int    `json:"number"`
	Title           string `json:"title"`
	URL             string `json:"url"`
	MergedAt        string `json:"mergedAt"`
	Language        string `json:"language,omitempty"`
	Summary         string `json:"summary"`
}

type ossJourneyResponse struct {
	Status     string                     `json:"status"`
	AnalyzedAt string                     `json:"analyzedAt"`
	Milestones []journeyMilestoneResponse `json:"milestones"`
}

type journeyMilestoneResponse struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	OccurredAt     string `json:"occurredAt"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	EvidenceURL    string `json:"evidenceUrl"`
	RepositoryName string `json:"repositoryName"`
	Technology     string `json:"technology,omitempty"`
}

type technologyEvidenceResponse struct {
	Kind   string `json:"kind"`
	Value  int    `json:"value"`
	Status string `json:"status"`
}

type ossExperienceResponse struct {
	Level      string                       `json:"level"`
	Confidence string                       `json:"confidence"`
	PublicOnly bool                         `json:"publicOnly"`
	Evidence   []technologyEvidenceResponse `json:"evidence"`
}

type repositorySampleResponse struct {
	Status              string                  `json:"status"`
	Observed            int                     `json:"observed"`
	Total               *int                    `json:"total"`
	Limit               int                     `json:"limit"`
	ActiveInWindow      int                     `json:"activeInWindow"`
	PrimaryTechnologies []languageShareResponse `json:"primaryTechnologies"`
}

type repositoryEvidenceResponse struct {
	Owned       repositorySampleResponse `json:"owned"`
	Contributed repositorySampleResponse `json:"contributed"`
	Starred     repositorySampleResponse `json:"starred"`
	Forked      repositorySampleResponse `json:"forked"`
}

type technologyProficiencyResponse struct {
	Name       string                       `json:"name"`
	Kind       string                       `json:"kind"`
	Level      int                          `json:"level"`
	Label      string                       `json:"label"`
	Score      int                          `json:"score"`
	Confidence string                       `json:"confidence"`
	Evidence   []technologyEvidenceResponse `json:"evidence"`
}

type analysisWindowResponse struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Days       int    `json:"days"`
	PublicOnly bool   `json:"publicOnly"`
}

func newGitHubProfileAnalysisResponse(
	analysis profile.Analysis,
) githubProfileAnalysisResponse {
	warnings := make([]profileWarningResponse, 0, len(analysis.Warnings))
	for _, warning := range analysis.Warnings {
		warnings = append(warnings, profileWarningResponse{
			Code:       warning.Code,
			Message:    warning.Message,
			Repository: warning.Repository,
		})
	}
	frameworks := make([]string, len(analysis.Frameworks))
	copy(frameworks, analysis.Frameworks)

	return githubProfileAnalysisResponse{
		Username:       analysis.Username.String(),
		Languages:      newLanguageShareResponses(analysis.Languages),
		LanguageStatus: string(analysis.LanguageStatus),
		Frameworks:     frameworks,
		RecentTechnologies: newRecentTechnologyResponses(
			analysis.RecentTechnologies,
		),
		Contributions: newContributionAnalysisResponse(
			analysis.Contributions,
		),
		ContributionCalendar: newContributionCalendarResponse(
			analysis.ContributionCalendar,
		),
		Portfolio:     newContributionPortfolioResponse(analysis.Portfolio),
		Journey:       newOSSJourneyResponse(analysis.Journey),
		OSSExperience: newOSSExperienceResponse(analysis.OSSExperience),
		RepositoryEvidence: newRepositoryEvidenceResponse(
			analysis.RepositoryEvidence,
		),
		Proficiency: newTechnologyProficiencyResponses(
			analysis.Proficiency,
		),
		AnalysisWindow: analysisWindowResponse{
			From:       analysis.Window.From.Format(time.RFC3339),
			To:         analysis.Window.To.Format(time.RFC3339),
			Days:       analysis.Window.Days,
			PublicOnly: analysis.Window.PublicOnly,
		},
		RepositoriesAnalyzed: analysis.RepositoriesAnalyzed,
		Warnings:             warnings,
	}
}

func newContributionCalendarResponse(
	calendar profile.ContributionCalendar,
) contributionCalendarResponse {
	weeks := make([]contributionWeekResponse, 0, len(calendar.Weeks))
	for _, week := range calendar.Weeks {
		days := make([]contributionDayResponse, 0, len(week.Days))
		for _, day := range week.Days {
			days = append(days, contributionDayResponse{
				Date:    day.Date.Format(time.DateOnly),
				Weekday: day.Weekday,
				Count:   day.Count,
				Level:   string(day.Level),
			})
		}
		weeks = append(weeks, contributionWeekResponse{
			Index:    week.Index,
			FirstDay: week.FirstDay.Format(time.DateOnly),
			Days:     days,
		})
	}
	response := contributionCalendarResponse{
		Status: string(calendar.Status),
		Total:  calendar.Total,
		Weeks:  weeks,
	}
	if !calendar.From.IsZero() {
		response.From = calendar.From.Format(time.DateOnly)
	}
	if !calendar.To.IsZero() {
		response.To = calendar.To.Format(time.DateOnly)
	}
	return response
}

func newOSSJourneyResponse(journey profile.OSSJourney) ossJourneyResponse {
	milestones := make([]journeyMilestoneResponse, 0, len(journey.Milestones))
	for _, milestone := range journey.Milestones {
		milestones = append(milestones, journeyMilestoneResponse{
			ID: milestone.ID, Kind: milestone.Kind,
			OccurredAt: milestone.OccurredAt.Format(time.RFC3339),
			Title:      milestone.Title, Description: milestone.Description,
			EvidenceURL:    milestone.EvidenceURL,
			RepositoryName: milestone.RepositoryName,
			Technology:     milestone.Technology,
		})
	}
	return ossJourneyResponse{
		Status:     string(journey.Status),
		AnalyzedAt: journey.AnalyzedAt.Format(time.RFC3339),
		Milestones: milestones,
	}
}

func newContributionPortfolioResponse(
	portfolio profile.ContributionPortfolio,
) contributionPortfolioResponse {
	languages := make([]portfolioLanguageResponse, 0, len(portfolio.Languages))
	for _, language := range portfolio.Languages {
		languages = append(languages, portfolioLanguageResponse{
			Name: language.Name, Count: language.Count,
		})
	}
	items := make(
		[]portfolioContributionResponse,
		0,
		len(portfolio.Contributions),
	)
	for _, item := range portfolio.Contributions {
		items = append(items, portfolioContributionResponse{
			RepositoryOwner: item.RepositoryOwner,
			RepositoryName:  item.RepositoryName,
			Number:          item.Number,
			Title:           item.Title,
			URL:             item.URL,
			MergedAt:        item.MergedAt.Format(time.RFC3339),
			Language:        item.Language,
			Summary: "Merged public pull request in " + item.RepositoryOwner +
				"/" + item.RepositoryName + ": " + item.Title,
		})
	}
	return contributionPortfolioResponse{
		Status:          string(portfolio.Status),
		TotalMerged:     portfolio.TotalMerged,
		DisplayedMerged: portfolio.DisplayedMerged,
		RepositoryCount: portfolio.RepositoryCount,
		HasMore:         portfolio.HasMore,
		AnalyzedAt:      portfolio.AnalyzedAt.Format(time.RFC3339),
		Languages:       languages,
		Contributions:   items,
	}
}

func newLanguageShareResponses(
	languages []profile.LanguageShare,
) []languageShareResponse {
	result := make([]languageShareResponse, 0, len(languages))
	for _, language := range languages {
		result = append(result, languageShareResponse{
			Name:       language.Name,
			Percentage: language.Percentage,
		})
	}
	return result
}

func newRecentTechnologyResponses(
	technologies []profile.RecentTechnology,
) []recentTechnologyResponse {
	result := make([]recentTechnologyResponse, 0, len(technologies))
	for _, technology := range technologies {
		sources := make(
			[]string,
			0,
			len(technology.RepositorySources),
		)
		for _, source := range technology.RepositorySources {
			sources = append(sources, string(source))
		}
		result = append(result, recentTechnologyResponse{
			Name:              technology.Name,
			Kind:              string(technology.Kind),
			LastUsedAt:        technology.LastUsedAt.Format(time.RFC3339),
			RepositoryCount:   technology.RepositoryCount,
			RepositorySources: sources,
			Confidence:        string(technology.Confidence),
		})
	}
	return result
}

func newCountMetricResponse(metric profile.CountMetric) countMetricResponse {
	return countMetricResponse{
		Value:  metric.Value,
		Status: string(metric.Status),
	}
}

func newContributionAnalysisResponse(
	analysis profile.ContributionAnalysis,
) contributionAnalysisResponse {
	return contributionAnalysisResponse{
		WindowDays:   analysis.WindowDays,
		Commits:      newCountMetricResponse(analysis.Commits),
		IssuesOpened: newCountMetricResponse(analysis.IssuesOpened),
		PullRequestsOpened: newCountMetricResponse(
			analysis.PullRequestsOpened,
		),
		PullRequestReviews: newCountMetricResponse(
			analysis.PullRequestReviews,
		),
		RepositoriesTouched: newCountMetricResponse(
			analysis.RepositoriesTouched,
		),
	}
}

func newTechnologyEvidenceResponses(
	evidence []profile.TechnologyEvidence,
) []technologyEvidenceResponse {
	result := make([]technologyEvidenceResponse, 0, len(evidence))
	for _, item := range evidence {
		result = append(result, technologyEvidenceResponse{
			Kind:   item.Kind,
			Value:  item.Value,
			Status: string(item.Status),
		})
	}
	return result
}

func newOSSExperienceResponse(
	experience profile.OSSExperience,
) ossExperienceResponse {
	return ossExperienceResponse{
		Level:      experience.Level,
		Confidence: string(experience.Confidence),
		PublicOnly: experience.PublicOnly,
		Evidence:   newTechnologyEvidenceResponses(experience.Evidence),
	}
}

func newRepositoryEvidenceResponse(
	evidence profile.RepositoryEvidence,
) repositoryEvidenceResponse {
	return repositoryEvidenceResponse{
		Owned:       newRepositorySampleResponse(evidence.Owned),
		Contributed: newRepositorySampleResponse(evidence.Contributed),
		Starred:     newRepositorySampleResponse(evidence.Starred),
		Forked:      newRepositorySampleResponse(evidence.Forked),
	}
}

func newRepositorySampleResponse(
	sample profile.RepositorySample,
) repositorySampleResponse {
	return repositorySampleResponse{
		Status:         string(sample.Status),
		Observed:       sample.Observed,
		Total:          sample.Total,
		Limit:          sample.Limit,
		ActiveInWindow: sample.ActiveInWindow,
		PrimaryTechnologies: newLanguageShareResponses(
			sample.PrimaryTechnologies,
		),
	}
}

func newTechnologyProficiencyResponses(
	proficiency []profile.TechnologyProficiency,
) []technologyProficiencyResponse {
	result := make(
		[]technologyProficiencyResponse,
		0,
		len(proficiency),
	)
	for _, technology := range proficiency {
		result = append(result, technologyProficiencyResponse{
			Name:       technology.Name,
			Kind:       string(technology.Kind),
			Level:      technology.Level,
			Label:      technology.Label,
			Score:      technology.Score,
			Confidence: string(technology.Confidence),
			Evidence:   newTechnologyEvidenceResponses(technology.Evidence),
		})
	}
	return result
}
