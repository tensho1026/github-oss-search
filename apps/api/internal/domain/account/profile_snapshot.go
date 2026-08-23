package account

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	// MaximumProfileSnapshots bounds one account's retained monthly history.
	MaximumProfileSnapshots = 24
	// MaximumSnapshotTechnologies bounds each technology evidence collection.
	MaximumSnapshotTechnologies = 20
	// MaximumSnapshotMetric bounds every retained non-negative counter.
	MaximumSnapshotMetric = 1_000_000
)

// SnapshotProficiency is one bounded technology level retained for trends.
type SnapshotProficiency struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

// ProfileSnapshot is one private calendar-month aggregate. It stores no
// repository payload, issue body, credential, or private GitHub evidence.
type ProfileSnapshot struct {
	AccountID          ID
	Month              time.Time
	Languages          []string
	Frameworks         []string
	OSSActivity        int
	MergedPullRequests int
	Proficiency        []SnapshotProficiency
	CompletedQuests    int
	CurrentStreak      int
	LongestStreak      int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewProfileSnapshot validates a bounded private trend aggregate and assigns
// it to the UTC month containing now.
func NewProfileSnapshot(
	accountID ID,
	languages, frameworks []string,
	ossActivity, mergedPullRequests int,
	proficiency []SnapshotProficiency,
	completedQuests, currentStreak, longestStreak int,
	now time.Time,
) (ProfileSnapshot, error) {
	languages, err := normalizeSnapshotNames(languages)
	if err != nil {
		return ProfileSnapshot{}, err
	}
	frameworks, err = normalizeSnapshotNames(frameworks)
	if err != nil {
		return ProfileSnapshot{}, err
	}
	if len(proficiency) > MaximumSnapshotTechnologies {
		return ProfileSnapshot{}, fmt.Errorf("%w: too many proficiency values", ErrInvalidFeatureInput)
	}
	seen := map[string]struct{}{}
	levels := make([]SnapshotProficiency, 0, len(proficiency))
	for _, value := range proficiency {
		names, parseErr := normalizeSnapshotNames([]string{value.Name})
		if parseErr != nil || value.Level < 1 || value.Level > 5 {
			return ProfileSnapshot{}, fmt.Errorf("%w: invalid proficiency", ErrInvalidFeatureInput)
		}
		key := strings.ToLower(names[0])
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		levels = append(levels, SnapshotProficiency{Name: names[0], Level: value.Level})
	}
	slices.SortFunc(levels, func(a, b SnapshotProficiency) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	for _, metric := range []int{ossActivity, mergedPullRequests, completedQuests, currentStreak, longestStreak} {
		if metric < 0 || metric > MaximumSnapshotMetric {
			return ProfileSnapshot{}, fmt.Errorf("%w: snapshot metric is out of range", ErrInvalidFeatureInput)
		}
	}
	month := now.UTC()
	month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	return ProfileSnapshot{AccountID: accountID, Month: month, Languages: languages, Frameworks: frameworks, OSSActivity: ossActivity, MergedPullRequests: mergedPullRequests, Proficiency: levels, CompletedQuests: completedQuests, CurrentStreak: currentStreak, LongestStreak: longestStreak}, nil
}

func normalizeSnapshotNames(values []string) ([]string, error) {
	if len(values) > MaximumSnapshotTechnologies {
		return nil, fmt.Errorf("%w: too many technologies", ErrInvalidFeatureInput)
	}
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" || utf8.RuneCountInString(value) > 64 || len(value) > 128 {
			return nil, fmt.Errorf("%w: invalid technology", ErrInvalidFeatureInput)
		}
		for _, character := range value {
			if unicode.IsControl(character) {
				return nil, fmt.Errorf("%w: invalid technology", ErrInvalidFeatureInput)
			}
		}
		key := strings.ToLower(value)
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	slices.SortFunc(result, func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) })
	return result, nil
}
