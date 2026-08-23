package account

import (
	"errors"
	"testing"
	"time"
)

func TestNewProfileSnapshotNormalizesBoundedMonthlyAggregate(t *testing.T) {
	t.Parallel()
	accountID, err := ParseID("123e4567-e89b-42d3-a456-426614174000")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := NewProfileSnapshot(
		accountID,
		[]string{" TypeScript ", "go", "Go"},
		[]string{"React"},
		42,
		3,
		[]SnapshotProficiency{{Name: "Go", Level: 3}, {Name: "typescript", Level: 4}},
		2,
		5,
		8,
		time.Date(2026, time.August, 23, 18, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
	)
	if err != nil {
		t.Fatalf("NewProfileSnapshot() error = %v", err)
	}
	if snapshot.Month != time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("month = %s", snapshot.Month)
	}
	if len(snapshot.Languages) != 2 || snapshot.Languages[0] != "go" || snapshot.Languages[1] != "TypeScript" {
		t.Fatalf("languages = %#v", snapshot.Languages)
	}
	if len(snapshot.Proficiency) != 2 || snapshot.Proficiency[0].Name != "Go" {
		t.Fatalf("proficiency = %#v", snapshot.Proficiency)
	}
}

func TestNewProfileSnapshotRejectsUnboundedOrUnsafeValues(t *testing.T) {
	t.Parallel()
	accountID, _ := ParseID("123e4567-e89b-42d3-a456-426614174000")
	tests := []struct {
		name        string
		languages   []string
		activity    int
		proficiency []SnapshotProficiency
	}{
		{name: "negative metric", activity: -1},
		{name: "control character", languages: []string{"Go\nRust"}},
		{name: "invalid proficiency", proficiency: []SnapshotProficiency{{Name: "Go", Level: 6}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewProfileSnapshot(accountID, test.languages, nil, test.activity, 0, test.proficiency, 0, 0, 0, time.Now())
			if !errors.Is(err, ErrInvalidFeatureInput) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
