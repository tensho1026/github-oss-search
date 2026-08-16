package usecase

import (
	"testing"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/profile"
)

func TestContributionProfileFromAnalysisPreservesSamplingAndStrength(t *testing.T) {
	t.Parallel()
	analysis := profile.Analysis{
		LanguageStatus: profile.EvidenceSampled,
		Proficiency: []profile.TechnologyProficiency{
			{
				Name:       "Go",
				Kind:       profile.TechnologyLanguage,
				Level:      4,
				Confidence: profile.ConfidenceHigh,
			},
			{
				Name:       "React",
				Kind:       profile.TechnologyFramework,
				Level:      2,
				Confidence: profile.ConfidenceMedium,
			},
		},
	}

	got, meta := contributionProfileFromAnalysis(analysis, true)
	if got.Status != issue.ContributionProfilePartial || !got.Personalized ||
		got.Version != issue.ContributionMatchScoreVersion || len(got.Skills) != 2 ||
		got.Skills[0].Strength != 4 || !meta.incomplete || !meta.cacheHit {
		t.Fatalf("profile = %+v, meta = %+v", got, meta)
	}
}

func TestContributionProfileFromAnalysisKeepsEmptyEvidenceUnavailable(t *testing.T) {
	t.Parallel()
	got, meta := contributionProfileFromAnalysis(profile.Analysis{}, false)
	if got.Status != issue.ContributionProfileUnavailable ||
		!meta.incomplete || len(got.Skills) != 0 {
		t.Fatalf("profile = %+v, meta = %+v", got, meta)
	}
}
