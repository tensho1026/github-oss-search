package issue

import "time"

// Recommendation score limits define the fixed 100-point model invariant.
const (
	MaximumRecommendationScore = 100
	// ContributionMatchScoreVersion changes only when the public-profile
	// evidence model or percentage calculation changes incompatibly.
	ContributionMatchScoreVersion = "v1"

	SkillScoreMaximum        = 30
	IssueQualityScoreMaximum = 20
	RepositoryScoreMaximum   = 15
	ActivityScoreMaximum     = 15
	MaintainerScoreMaximum   = 10
	AvailabilityScoreMaximum = 10
)

// MatchStatus records whether a required technology is present in the
// contributor's explicitly supplied skill set. Unknown evidence is excluded
// from the percentage denominator.
type MatchStatus string

// MatchStatus values preserve unavailable evidence separately from mismatch.
const (
	MatchMatched   MatchStatus = "matched"
	MatchPartial   MatchStatus = "partial"
	MatchUnmatched MatchStatus = "unmatched"
	MatchUnknown   MatchStatus = "unknown"
)

// ContributionProfileStatus distinguishes usable public evidence from a
// bounded sample and from a profile that could not support a comparison.
type ContributionProfileStatus string

const (
	// ContributionProfileAvailable means the bounded profile supports a
	// complete comparison within its documented analysis window.
	ContributionProfileAvailable ContributionProfileStatus = "available"
	// ContributionProfilePartial means usable skills were observed while one
	// or more profile collections were sampled or incomplete.
	ContributionProfilePartial ContributionProfileStatus = "partial"
	// ContributionProfileUnavailable means no usable public technology
	// evidence was available; clients must not present a zero-skill claim.
	ContributionProfileUnavailable ContributionProfileStatus = "unavailable"
)

// ContributorSkill is one normalized public-profile technology observation.
// Strength is the deterministic profile proficiency level in the range 1-5.
type ContributorSkill struct {
	Name       string
	Strength   int
	Confidence Confidence
	Evidence   []Evidence
}

// ContributorProfile carries only bounded, normalized public evidence into
// issue recommendation. It deliberately excludes GitHub transport payloads.
type ContributorProfile struct {
	Status       ContributionProfileStatus
	Skills       []ContributorSkill
	Personalized bool
	Version      string
}

// SkillMatch is one explainable technology comparison.
type SkillMatch struct {
	Technology          string
	Status              MatchStatus
	Confidence          Confidence
	RequirementEvidence []Evidence
	ContributorEvidence []Evidence
}

// SkillMatchAssessment exposes the exact denominator used to calculate the
// percentage so unavailable evidence cannot appear as a mismatch.
type SkillMatchAssessment struct {
	Percentage   int
	Matched      int
	Partial      int
	Denominator  int
	Status       ContributionProfileStatus
	Personalized bool
	Version      string
	Skills       []SkillMatch
}

// RepositorySignalKey identifies one independently observed contribution
// readiness signal.
type RepositorySignalKey string

// RepositorySignalKey values enumerate contribution-readiness observations.
const (
	RepositoryREADME        RepositorySignalKey = "readme"
	RepositoryContributing  RepositorySignalKey = "contributing"
	RepositoryCI            RepositorySignalKey = "ci"
	RepositoryTests         RepositorySignalKey = "tests"
	RepositoryCodeOfConduct RepositorySignalKey = "code_of_conduct"
)

// RepositorySignal separates observed absence from incomplete inspection.
type RepositorySignal struct {
	Key      RepositorySignalKey
	State    SignalState
	Evidence []Evidence
}

// CIState is the normalized state of the default branch's latest check rollup.
type CIState string

// CIState values normalize the default branch's latest check rollup.
const (
	CIStateSuccess CIState = "success"
	CIStateFailure CIState = "failure"
	CIStatePending CIState = "pending"
	CIStateUnknown CIState = "unknown"
)

// AggregateStatus prevents missing and empty samples from becoming a
// misleading zero duration.
type AggregateStatus string

// AggregateStatus values distinguish usable samples from absent evidence.
const (
	AggregateAvailable   AggregateStatus = "available"
	AggregateUnavailable AggregateStatus = "unavailable"
)

// DurationAggregate describes a bounded public sample. Median and 90th
// percentile are robust summaries over completed, non-bot observations.
type DurationAggregate struct {
	Status       AggregateStatus
	Median       time.Duration
	Percentile90 time.Duration
	SampleSize   int
	WindowDays   int
	Truncated    bool
	Confidence   Confidence
}

// RatioAggregate describes a bounded numerator/denominator observation.
type RatioAggregate struct {
	Status      AggregateStatus
	Numerator   int
	Denominator int
	Percentage  int
	SampleSize  int
	WindowDays  int
	Truncated   bool
	Confidence  Confidence
}

// CountAggregate is a bounded count observation. Value is meaningful only
// when Status is available.
type CountAggregate struct {
	Status     AggregateStatus
	Value      int
	SampleSize int
	WindowDays int
	Truncated  bool
	Confidence Confidence
}

// ActivityMetrics holds public repository facts and sampled maintenance
// aggregates. A zero value is not meaningful unless the corresponding status
// says it is available.
type ActivityMetrics struct {
	LastMeaningfulUpdate  time.Time
	CI                    CIState
	Contributors          CountAggregate
	PullRequestsOpened    CountAggregate
	StaleOpenPullRequests CountAggregate
	UnansweredIssues      CountAggregate
	PullRequestMerge      RatioAggregate
	IssueResponse         DurationAggregate
	PullRequestReview     DurationAggregate
	PullRequestMergeTime  DurationAggregate
}

// MaintainerResponseAssessment summarizes bounded historical maintainer
// interactions without promising a future response or merge.
type MaintainerResponseAssessment struct {
	Status             AggregateStatus
	Level              int
	Label              string
	Confidence         Confidence
	SampleSize         int
	WindowDays         int
	ResponseCoverage   RatioAggregate
	FirstIssueResponse DurationAggregate
	FirstPullReview    DurationAggregate
	PullRequestMerge   DurationAggregate
}

// ClaimEvidence is a conservative observation that another contributor has
// explicitly claimed the issue.
type ClaimEvidence struct {
	Claimed    bool
	Confidence Confidence
	Evidence   []Evidence
}

// StaleState distinguishes active opportunities from aging, stale, and
// insufficiently observed issues.
type StaleState string

// StaleState values are stable API classifications.
const (
	StaleFresh   StaleState = "fresh"
	StaleAging   StaleState = "aging"
	StaleStale   StaleState = "stale"
	StaleUnknown StaleState = "unknown"
)

// LinkedPullRequestObservation is one bounded closing-reference observation.
type LinkedPullRequestObservation struct {
	Number    int
	State     string
	IsDraft   bool
	UpdatedAt time.Time
	MergedAt  time.Time
}

// IssueHistory contains bounded issue-level events used only for stale
// classification. Truncation preserves uncertainty when absence is unknown.
type IssueHistory struct {
	Comments                    []CommentObservation
	CommentsTruncated           bool
	LinkedPullRequests          []LinkedPullRequestObservation
	LinkedPullRequestsTruncated bool
}

// StaleAssessment is an explainable result from the versioned stale policy.
type StaleAssessment struct {
	State                         StaleState
	PolicyVersion                 string
	Confidence                    Confidence
	AnalyzedAt                    time.Time
	FreshWithinDays               int
	StaleAfterDays                int
	IssueCreatedAt                time.Time
	IssueUpdatedAt                time.Time
	RepositoryActivityAt          time.Time
	LastMeaningfulIssueActivityAt time.Time
	LastMaintainerActivityAt      time.Time
	LastLinkedPullRequestAt       time.Time
	SampleSize                    int
	Truncated                     bool
	Evidence                      []Evidence
}

// Severity orders warning presentation and allows clients to style risk.
type Severity string

// Severity values order heuristic warning presentation.
const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Warning is an explainable heuristic risk. Evidence contains normalized
// facts and never arbitrary comment or issue text.
type Warning struct {
	Code     string
	Severity Severity
	Message  string
	Evidence []Evidence
}

// ScoreComponent is one fixed part of the documented 100-point model.
type ScoreComponent struct {
	Name    string
	Score   int
	Maximum int
	Reasons []string
}

// ScoreBreakdown makes the score invariant observable.
type ScoreBreakdown struct {
	SkillMatch        ScoreComponent
	IssueQuality      ScoreComponent
	RepositoryQuality ScoreComponent
	Activity          ScoreComponent
	Maintainer        ScoreComponent
	Availability      ScoreComponent
}

// RecommendationInput is the complete transport-independent scoring input.
type RecommendationInput struct {
	Candidate          Candidate
	Analysis           Analysis
	DesiredSkills      []string
	ContributorProfile ContributorProfile
	RepositorySignals  []RepositorySignal
	Activity           ActivityMetrics
	Claim              ClaimEvidence
	History            IssueHistory
	Now                time.Time
}

// Recommendation is the deterministic ranked result shared by list and
// detail transports.
type Recommendation struct {
	Score              int
	Breakdown          ScoreBreakdown
	SkillMatch         SkillMatchAssessment
	RepositorySignals  []RepositorySignal
	Activity           ActivityMetrics
	MaintainerResponse MaintainerResponseAssessment
	Claim              ClaimEvidence
	Stale              StaleAssessment
	Reasons            []string
	Warnings           []Warning
}

// RankedIssue is the shared evaluated shape used by both list and detail
// application flows.
type RankedIssue struct {
	Candidate        Candidate
	Analysis         Analysis
	Recommendation   Recommendation
	RepositoryHealth RepositoryHealthDashboard
}

// CommentObservation contains only the bounded public fields needed for
// conservative claim detection.
type CommentObservation struct {
	AuthorLogin       string
	AuthorType        string
	AuthorAssociation string
	Body              string
	CreatedAt         time.Time
}
