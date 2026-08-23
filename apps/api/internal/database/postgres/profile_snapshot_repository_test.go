package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/account"
)

func TestProfileSnapshotRepositoryListsAndUpsertsOwnedMonthlyData(t *testing.T) {
	t.Parallel()
	accountID := mustAccountID(t)
	month := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	values := snapshotRowValues(month)
	executor := &scriptedAccountExecutor{
		queryRows: &valueRows{rows: []valueRow{{values: values}}},
		rowQueue:  []pgx.Row{valueRow{values: values}},
	}
	repository := AccountRepository{executor: executor, queryTimeout: time.Second}

	items, err := repository.ListProfileSnapshots(context.Background(), accountID)
	if err != nil || len(items) != 1 || items[0].Languages[0] != "Go" ||
		items[0].Proficiency[0].Level != 3 {
		t.Fatalf("ListProfileSnapshots() = %+v, %v", items, err)
	}
	written, err := repository.UpsertProfileSnapshot(context.Background(), items[0])
	if err != nil || written.Month != month || written.OSSActivity != 10 {
		t.Fatalf("UpsertProfileSnapshot() = %+v, %v", written, err)
	}
	if len(executor.calls) != 2 || executor.calls[0].arguments[0] != accountID.String() ||
		!strings.Contains(executor.calls[1].query, "count(*)") ||
		executor.calls[1].arguments[10] != account.MaximumProfileSnapshots {
		t.Fatalf("calls = %+v", executor.calls)
	}
}

func TestProfileSnapshotRepositoryMapsQueryAndMissingErrors(t *testing.T) {
	t.Parallel()
	accountID := mustAccountID(t)
	queryFailure := AccountRepository{
		executor:     &scriptedAccountExecutor{queryError: errors.New("secret driver error")},
		queryTimeout: time.Second,
	}
	if _, err := queryFailure.ListProfileSnapshots(context.Background(), accountID); !errors.Is(err, ErrQueryFailed) {
		t.Fatalf("ListProfileSnapshots() error = %v", err)
	}
	missing := AccountRepository{
		executor:     &scriptedAccountExecutor{rowQueue: []pgx.Row{valueRow{err: pgx.ErrNoRows}}},
		queryTimeout: time.Second,
	}
	snapshot, _ := account.NewProfileSnapshot(accountID, nil, nil, 0, 0, nil, 0, 0, 0, time.Now())
	if _, err := missing.UpsertProfileSnapshot(context.Background(), snapshot); !errors.Is(err, account.ErrNotFound) {
		t.Fatalf("UpsertProfileSnapshot() error = %v", err)
	}
}

func snapshotRowValues(month time.Time) []any {
	return []any{
		month,
		[]byte(`["Go"]`),
		[]byte(`["React"]`),
		10,
		2,
		[]byte(`[{"name":"Go","level":3}]`),
		1,
		2,
		4,
		month,
		month,
	}
}
