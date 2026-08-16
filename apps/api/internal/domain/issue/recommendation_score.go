package issue

import (
	"cmp"
	"slices"
	"strings"
	"time"
)

const recommendationActivityWindowDays = 180

// Recommend calculates the documented 100-point score and stable reasons and
// warnings. It performs no I/O and does not mutate the supplied slices.
func Recommend(input RecommendationInput) Recommendation {
	now := input.Now.UTC()
	if now.IsZero() {
		now = input.Candidate.Issue.UpdatedAt.UTC()
	}

	contributorProfile := input.ContributorProfile
	if len(contributorProfile.Skills) == 0 && len(input.DesiredSkills) > 0 {
		contributorProfile = contributorProfileFromExplicitSkills(
			input.DesiredSkills,
		)
	}
	skillMatch := assessContributionMatch(
		input.Analysis.RequiredTechnologies,
		contributorProfile,
	)
	signals := normalizeRepositorySignals(input.RepositorySignals)
	maintainerResponse := AssessMaintainerResponse(input.Activity)
	stale := AssessStaleness(input)
	breakdown := ScoreBreakdown{
		SkillMatch:        scoreSkillMatch(skillMatch),
		IssueQuality:      scoreIssueQuality(input.Analysis.Quality),
		RepositoryQuality: scoreRepositoryQuality(signals),
		Activity:          scoreActivity(input.Activity, now),
		Maintainer:        scoreMaintainer(maintainerResponse),
		Availability: scoreAvailability(
			input.Candidate,
			input.Claim,
			now,
		),
	}

	components := scoreComponents(breakdown)
	total := 0
	reasons := make([]string, 0, len(components))
	for _, component := range components {
		total += component.Score
		reasons = append(reasons, component.Reasons...)
	}
	total = min(total, MaximumRecommendationScore)
	slices.Sort(reasons)
	reasons = slices.Compact(reasons)

	return Recommendation{
		Score:              total,
		Breakdown:          breakdown,
		SkillMatch:         skillMatch,
		RepositorySignals:  signals,
		Activity:           input.Activity,
		MaintainerResponse: maintainerResponse,
		Claim:              cloneClaimEvidence(input.Claim),
		Stale:              stale,
		Reasons:            reasons,
		Warnings:           recommendationWarnings(input, signals, stale, now),
	}
}

func assessContributionMatch(
	required []RequiredTechnology,
	profile ContributorProfile,
) SkillMatchAssessment {
	profile = normalizeContributorProfile(profile)
	contributorByName := make(map[string]ContributorSkill, len(profile.Skills))
	for _, skill := range profile.Skills {
		contributorByName[normalizeSkill(skill.Name)] = skill
	}

	requiredByName := make(map[string]RequiredTechnology, len(required))
	for _, technology := range required {
		name := strings.TrimSpace(technology.Name)
		if name == "" {
			continue
		}
		key := normalizeSkill(name)
		if current, exists := requiredByName[key]; exists {
			current.Evidence = append(
				append([]Evidence(nil), current.Evidence...),
				technology.Evidence...,
			)
			if confidenceRank(technology.Confidence) >
				confidenceRank(current.Confidence) {
				current.Confidence = technology.Confidence
			}
			requiredByName[key] = current
			continue
		}
		technology.Evidence = append([]Evidence(nil), technology.Evidence...)
		requiredByName[key] = technology
	}

	keys := make([]string, 0, len(requiredByName))
	for key := range requiredByName {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	matches := make([]SkillMatch, 0, len(keys))
	matched := 0
	partial := 0
	denominator := 0
	for _, key := range keys {
		technology := requiredByName[key]
		status := MatchUnknown
		confidence := ConfidenceLow
		var contributorEvidence []Evidence
		if technology.Confidence != ConfidenceLow &&
			profile.Status != ContributionProfileUnavailable {
			denominator++
			if contributor, exists := contributorByName[key]; exists {
				contributorEvidence = append(
					[]Evidence(nil),
					contributor.Evidence...,
				)
				confidence = minimumConfidence(
					technology.Confidence,
					contributor.Confidence,
				)
				if contributor.Strength >= 3 &&
					contributor.Confidence != ConfidenceLow {
					status = MatchMatched
					matched++
				} else {
					status = MatchPartial
					partial++
				}
			} else {
				status = MatchUnmatched
				confidence = technology.Confidence
			}
		}
		matches = append(matches, SkillMatch{
			Technology:          technology.Name,
			Status:              status,
			Confidence:          confidence,
			RequirementEvidence: append([]Evidence(nil), technology.Evidence...),
			ContributorEvidence: contributorEvidence,
		})
	}

	percentage := 0
	if denominator > 0 {
		percentage = roundedPercentage(matched*2+partial, denominator*2)
	}
	return SkillMatchAssessment{
		Percentage:   percentage,
		Matched:      matched,
		Partial:      partial,
		Denominator:  denominator,
		Status:       profile.Status,
		Personalized: profile.Personalized,
		Version:      profile.Version,
		Skills:       matches,
	}
}

func contributorProfileFromExplicitSkills(skills []string) ContributorProfile {
	result := ContributorProfile{
		Status:  ContributionProfileAvailable,
		Version: ContributionMatchScoreVersion,
		Skills:  make([]ContributorSkill, 0, len(skills)),
	}
	for _, value := range skills {
		name := strings.TrimSpace(value)
		if name == "" {
			continue
		}
		result.Skills = append(result.Skills, ContributorSkill{
			Name:       name,
			Strength:   5,
			Confidence: ConfidenceHigh,
			Evidence: []Evidence{{
				RuleID:      "contribution-match.explicit-skill",
				Source:      EvidenceDerived,
				Description: "technology was explicitly selected for this recommendation",
			}},
		})
	}
	if len(result.Skills) == 0 {
		result.Status = ContributionProfileUnavailable
	}
	return result
}

func normalizeContributorProfile(profile ContributorProfile) ContributorProfile {
	if profile.Version == "" {
		profile.Version = ContributionMatchScoreVersion
	}
	if profile.Status != ContributionProfileAvailable &&
		profile.Status != ContributionProfilePartial {
		profile.Status = ContributionProfileUnavailable
	}
	byName := make(map[string]ContributorSkill, len(profile.Skills))
	for _, skill := range profile.Skills {
		name := strings.TrimSpace(skill.Name)
		key := normalizeSkill(name)
		if key == "" || skill.Strength < 1 || skill.Strength > 5 {
			continue
		}
		skill.Name = name
		skill.Evidence = append([]Evidence(nil), skill.Evidence...)
		current, exists := byName[key]
		if !exists || skill.Strength > current.Strength ||
			(skill.Strength == current.Strength &&
				confidenceRank(skill.Confidence) > confidenceRank(current.Confidence)) {
			byName[key] = skill
		}
	}
	keys := make([]string, 0, len(byName))
	for key := range byName {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	profile.Skills = make([]ContributorSkill, 0, len(keys))
	for _, key := range keys {
		profile.Skills = append(profile.Skills, byName[key])
	}
	if len(profile.Skills) == 0 {
		profile.Status = ContributionProfileUnavailable
	}
	return profile
}

func minimumConfidence(left, right Confidence) Confidence {
	if confidenceRank(left) <= confidenceRank(right) {
		return left
	}
	return right
}

func normalizeSkill(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func scoreSkillMatch(match SkillMatchAssessment) ScoreComponent {
	score := scaleScore(match.Percentage, SkillScoreMaximum)
	reasons := []string{}
	if match.Denominator == 0 {
		reasons = append(reasons, "Contribution match is unavailable")
	} else if match.Personalized {
		reasons = append(
			reasons,
			"Contribution match uses bounded public GitHub profile evidence",
		)
	} else {
		reasons = append(
			reasons,
			"Contribution match uses explicitly selected technologies",
		)
	}
	return ScoreComponent{
		Name:    "skill_match",
		Score:   score,
		Maximum: SkillScoreMaximum,
		Reasons: reasons,
	}
}

func scoreIssueQuality(quality QualityAssessment) ScoreComponent {
	score := scaleScore(clamp(quality.Score, 0, 100), IssueQualityScoreMaximum)
	reasons := []string{"Issue quality uses independently observed description signals"}
	return ScoreComponent{
		Name:    "issue_quality",
		Score:   score,
		Maximum: IssueQualityScoreMaximum,
		Reasons: reasons,
	}
}

func normalizeRepositorySignals(
	input []RepositorySignal,
) []RepositorySignal {
	byKey := make(map[RepositorySignalKey]RepositorySignal, len(input))
	for _, signal := range input {
		if !validRepositorySignalKey(signal.Key) ||
			!validSignalState(signal.State) {
			continue
		}
		signal.Evidence = append([]Evidence(nil), signal.Evidence...)
		byKey[signal.Key] = signal
	}

	keys := []RepositorySignalKey{
		RepositoryREADME,
		RepositoryContributing,
		RepositoryCI,
		RepositoryTests,
		RepositoryCodeOfConduct,
	}
	signals := make([]RepositorySignal, 0, len(keys))
	for _, key := range keys {
		signal, exists := byKey[key]
		if !exists {
			signal = RepositorySignal{Key: key, State: SignalUnknown}
		}
		signals = append(signals, signal)
	}
	return signals
}

func validRepositorySignalKey(key RepositorySignalKey) bool {
	switch key {
	case RepositoryREADME,
		RepositoryContributing,
		RepositoryCI,
		RepositoryTests,
		RepositoryCodeOfConduct:
		return true
	default:
		return false
	}
}

func validSignalState(state SignalState) bool {
	switch state {
	case SignalPresent, SignalAbsent, SignalNotApplicable, SignalUnknown:
		return true
	default:
		return false
	}
}

func scoreRepositoryQuality(signals []RepositorySignal) ScoreComponent {
	weights := map[RepositorySignalKey]int{
		RepositoryREADME:        3,
		RepositoryContributing:  4,
		RepositoryCI:            3,
		RepositoryTests:         3,
		RepositoryCodeOfConduct: 2,
	}
	score := 0
	reasons := make([]string, 0, len(signals))
	for _, signal := range signals {
		if signal.State == SignalPresent {
			score += weights[signal.Key]
			reasons = append(
				reasons,
				"Repository provides "+string(signal.Key),
			)
		}
	}
	return ScoreComponent{
		Name:    "repository_quality",
		Score:   score,
		Maximum: RepositoryScoreMaximum,
		Reasons: reasons,
	}
}

func scoreActivity(
	activity ActivityMetrics,
	now time.Time,
) ScoreComponent {
	score := 0
	reasons := make([]string, 0, 4)
	if !activity.LastMeaningfulUpdate.IsZero() {
		age := now.Sub(activity.LastMeaningfulUpdate.UTC())
		switch {
		case age <= 30*24*time.Hour:
			score += 6
			reasons = append(reasons, "Repository activity was observed within 30 days")
		case age <= 90*24*time.Hour:
			score += 4
			reasons = append(reasons, "Repository activity was observed within 90 days")
		case age <= 180*24*time.Hour:
			score += 2
			reasons = append(reasons, "Repository activity was observed within 180 days")
		}
	}
	if activity.PullRequestMerge.Status == AggregateAvailable {
		switch {
		case activity.PullRequestMerge.Percentage >= 70:
			score += 4
		case activity.PullRequestMerge.Percentage >= 40:
			score += 2
		}
		reasons = append(
			reasons,
			"Pull request cadence uses a bounded 180-day sample",
		)
	}
	if activity.CI == CIStateSuccess {
		score += 3
		reasons = append(reasons, "Default branch checks are successful")
	}
	if activity.Contributors.Status == AggregateAvailable &&
		activity.Contributors.Value >= 2 {
		score += 2
		reasons = append(reasons, "Multiple public contributors were observed")
	}
	return ScoreComponent{
		Name:    "activity",
		Score:   score,
		Maximum: ActivityScoreMaximum,
		Reasons: reasons,
	}
}

func scoreMaintainer(assessment MaintainerResponseAssessment) ScoreComponent {
	scoreByLevel := [...]int{0, 1, 3, 6, 8, 10}
	score := 0
	if assessment.Status == AggregateAvailable && assessment.Level >= 1 &&
		assessment.Level <= 5 {
		score = scoreByLevel[assessment.Level]
	}
	reasons := make([]string, 0, 2)
	if assessment.Status == AggregateAvailable {
		reasons = append(reasons, "Maintainer response uses bounded maintainer-only samples")
	}
	if assessment.ResponseCoverage.Status == AggregateAvailable {
		reasons = append(reasons, "Issue response coverage includes unanswered sampled issues")
	}
	return ScoreComponent{
		Name:    "maintainer_responsiveness",
		Score:   score,
		Maximum: MaintainerScoreMaximum,
		Reasons: reasons,
	}
}

func scoreAvailability(
	candidate Candidate,
	claim ClaimEvidence,
	now time.Time,
) ScoreComponent {
	score := 0
	reasons := make([]string, 0, 3)
	if len(candidate.Issue.Assignees) == 0 {
		score += 4
		reasons = append(reasons, "Issue has no assignee")
	}
	if !claim.Claimed {
		score += 3
		reasons = append(reasons, "No explicit contributor claim was observed")
	}
	if !candidate.Issue.UpdatedAt.IsZero() &&
		now.Sub(candidate.Issue.UpdatedAt.UTC()) <= 90*24*time.Hour {
		score += 3
		reasons = append(reasons, "Issue activity was observed within 90 days")
	}
	return ScoreComponent{
		Name:    "availability",
		Score:   score,
		Maximum: AvailabilityScoreMaximum,
		Reasons: reasons,
	}
}

func recommendationWarnings(
	input RecommendationInput,
	signals []RepositorySignal,
	stale StaleAssessment,
	now time.Time,
) []Warning {
	warnings := make([]Warning, 0, 8)
	activityWindow := time.Duration(recommendationActivityWindowDays) *
		24 * time.Hour
	if input.Claim.Claimed {
		warnings = append(warnings, Warning{
			Code:     "likely_claimed",
			Severity: SeverityWarning,
			Message:  "A contributor explicitly indicated that they are working on this issue",
			Evidence: append([]Evidence(nil), input.Claim.Evidence...),
		})
	}
	switch stale.State {
	case StaleStale:
		warnings = append(warnings, Warning{
			Code:     "stale_issue",
			Severity: SeverityWarning,
			Message:  "This issue appears stale under the bounded stale-v1 policy",
			Evidence: append([]Evidence(nil), stale.Evidence...),
		})
	case StaleAging:
		warnings = append(warnings, Warning{
			Code:     "aging_issue",
			Severity: SeverityInfo,
			Message:  "This issue is aging; review recent activity before starting",
			Evidence: append([]Evidence(nil), stale.Evidence...),
		})
	case StaleUnknown:
		warnings = append(warnings, Warning{
			Code:     "stale_status_unknown",
			Severity: SeverityInfo,
			Message:  "Stale status is unknown because bounded history was incomplete",
			Evidence: append([]Evidence(nil), stale.Evidence...),
		})
	}
	if !input.Activity.LastMeaningfulUpdate.IsZero() &&
		now.Sub(input.Activity.LastMeaningfulUpdate.UTC()) > activityWindow {
		warnings = append(warnings, Warning{
			Code:     "stale_repository",
			Severity: SeverityCritical,
			Message:  "No meaningful repository activity was observed within 180 days",
			Evidence: []Evidence{{
				RuleID:      "activity.repository.stale",
				Source:      EvidenceDerived,
				Description: "last meaningful update is older than the activity window",
			}},
		})
	}
	if input.Activity.CI == CIStateFailure {
		warnings = append(warnings, Warning{
			Code:     "failing_ci",
			Severity: SeverityCritical,
			Message:  "The latest default-branch check rollup is failing",
			Evidence: []Evidence{{
				RuleID:      "activity.ci.failure",
				Source:      EvidenceDerived,
				Description: "default branch check rollup reports failure",
			}},
		})
	}
	if input.Activity.IssueResponse.Status == AggregateAvailable &&
		input.Activity.IssueResponse.Median > 14*24*time.Hour {
		warnings = append(warnings, Warning{
			Code:     "slow_issue_response",
			Severity: SeverityWarning,
			Message:  "Maintainer issue responses are slow in the bounded sample",
			Evidence: []Evidence{{
				RuleID:      "maintainer.issue_response.slow",
				Source:      EvidenceDerived,
				Description: "median maintainer response exceeds 14 days",
			}},
		})
	}
	if input.Activity.StaleOpenPullRequests.Status == AggregateAvailable &&
		input.Activity.StaleOpenPullRequests.Value > 0 {
		warnings = append(warnings, Warning{
			Code:     "abandoned_pull_request_risk",
			Severity: SeverityWarning,
			Message:  "Stale open pull requests were observed in the bounded sample",
			Evidence: []Evidence{{
				RuleID:      "maintainer.pull_request.long_lived",
				Source:      EvidenceDerived,
				Description: "an open pull request is older than 60 days and inactive for 30 days",
			}},
		})
	}
	if input.Activity.UnansweredIssues.Status == AggregateAvailable &&
		input.Activity.UnansweredIssues.Value > 0 {
		warnings = append(warnings, Warning{
			Code:     "unanswered_issue_risk",
			Severity: SeverityWarning,
			Message:  "Issues without a maintainer response were observed in the bounded sample",
			Evidence: []Evidence{{
				RuleID:      "maintainer.issue.unanswered",
				Source:      EvidenceDerived,
				Description: "an issue older than 14 days has no observed maintainer response",
			}},
		})
	}

	for _, signal := range signals {
		if signal.State == SignalUnknown {
			warnings = append(warnings, Warning{
				Code:     "repository_signal_unavailable",
				Severity: SeverityInfo,
				Message:  "Repository inspection is incomplete for " + string(signal.Key),
				Evidence: append([]Evidence(nil), signal.Evidence...),
			})
		}
	}
	slices.SortFunc(warnings, func(left, right Warning) int {
		if result := cmp.Compare(severityRank(right.Severity), severityRank(left.Severity)); result != 0 {
			return result
		}
		return cmp.Compare(left.Code, right.Code)
	})
	return warnings
}

func scoreComponents(breakdown ScoreBreakdown) []ScoreComponent {
	return []ScoreComponent{
		breakdown.SkillMatch,
		breakdown.IssueQuality,
		breakdown.RepositoryQuality,
		breakdown.Activity,
		breakdown.Maintainer,
		breakdown.Availability,
	}
}

func roundedPercentage(numerator, denominator int) int {
	if denominator <= 0 {
		return 0
	}
	return clamp((numerator*100+denominator/2)/denominator, 0, 100)
}

func scaleScore(percentage, maximum int) int {
	return (clamp(percentage, 0, 100)*maximum + 50) / 100
}

func clamp(value, minimum, maximum int) int {
	return min(max(value, minimum), maximum)
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityCritical:
		return 3
	case SeverityWarning:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

func cloneClaimEvidence(claim ClaimEvidence) ClaimEvidence {
	claim.Evidence = append([]Evidence(nil), claim.Evidence...)
	return claim
}
