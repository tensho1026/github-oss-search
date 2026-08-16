package usecase

import (
	"fmt"
	"strings"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/profile"
)

type contributionProfileMeta struct {
	status     issue.ContributionProfileStatus
	incomplete bool
	cacheHit   bool
}

func contributionProfileFromAnalysis(
	analysis profile.Analysis,
	cacheHit bool,
) (issue.ContributorProfile, contributionProfileMeta) {
	status := issue.ContributionProfileAvailable
	if analysis.LanguageStatus != profile.EvidenceExact || len(analysis.Warnings) > 0 {
		status = issue.ContributionProfilePartial
	}

	skills := make([]issue.ContributorSkill, 0, len(analysis.Proficiency))
	for _, technology := range analysis.Proficiency {
		if technology.Level < 1 || technology.Level > 5 ||
			technology.Confidence == profile.ConfidenceUnavailable {
			status = issue.ContributionProfilePartial
			continue
		}
		skills = append(skills, issue.ContributorSkill{
			Name:       technology.Name,
			Strength:   technology.Level,
			Confidence: contributionConfidence(technology.Confidence),
			Evidence: []issue.Evidence{{
				RuleID: "contribution-match.public-profile",
				Source: issue.EvidenceDerived,
				Description: fmt.Sprintf(
					"public %s evidence supports proficiency level %d/5",
					strings.TrimSpace(string(technology.Kind)),
					technology.Level,
				),
			}},
		})
	}
	if len(skills) == 0 {
		status = issue.ContributionProfileUnavailable
	}
	return issue.ContributorProfile{
			Status:       status,
			Skills:       skills,
			Personalized: true,
			Version:      issue.ContributionMatchScoreVersion,
		}, contributionProfileMeta{
			status:     status,
			incomplete: status != issue.ContributionProfileAvailable,
			cacheHit:   cacheHit,
		}
}

func contributionConfidence(value profile.Confidence) issue.Confidence {
	switch value {
	case profile.ConfidenceHigh:
		return issue.ConfidenceHigh
	case profile.ConfidenceMedium:
		return issue.ConfidenceMedium
	default:
		return issue.ConfidenceLow
	}
}

func explicitContributionProfile(skills []string) issue.ContributorProfile {
	result := issue.ContributorProfile{
		Status:  issue.ContributionProfileAvailable,
		Version: issue.ContributionMatchScoreVersion,
		Skills:  make([]issue.ContributorSkill, 0, len(skills)),
	}
	for _, value := range skills {
		name := strings.TrimSpace(value)
		if name == "" {
			continue
		}
		result.Skills = append(result.Skills, issue.ContributorSkill{
			Name:       name,
			Strength:   5,
			Confidence: issue.ConfidenceHigh,
			Evidence: []issue.Evidence{{
				RuleID:      "contribution-match.explicit-filter",
				Source:      issue.EvidenceDerived,
				Description: "technology was explicitly selected in search filters",
			}},
		})
	}
	if len(result.Skills) == 0 {
		result.Status = issue.ContributionProfileUnavailable
	}
	return result
}
