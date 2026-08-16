package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/account"
)

func TestAccountRepositoryListsIssueClaimsWithSummary(t *testing.T) {
	accountID := mustAccountID(t)
	now := time.Date(2026, 8, 16, 1, 2, 3, 0, time.UTC)
	executor := &scriptedAccountExecutor{
		queryRows: &valueRows{rows: []valueRow{{values: []any{
			featureResourceID,
			"OpenAI",
			"openai-go",
			42,
			"implementing",
			false,
			nil,
			nil,
			nil,
			"open",
			"unverified",
			int64(3),
			now,
			now,
			7,
			1,
			2,
			3,
			0,
			1,
			1,
		}}}},
	}
	repository := AccountRepository{executor: executor, queryTimeout: time.Second}
	page, _ := account.NewPage(2, 10)

	result, err := repository.ListIssueClaims(context.Background(), accountID, page)
	if err != nil {
		t.Fatalf("ListIssueClaims() error = %v", err)
	}
	if len(result.Items) != 1 || result.Total != 7 ||
		result.Summary.Implementing != 3 || result.Summary.Archived != 1 {
		t.Fatalf("ListIssueClaims() result = %+v", result)
	}
	claim := result.Items[0]
	if claim.AccountID != accountID || claim.Issue.RepositoryOwner != "openai" ||
		claim.Status != account.IssueClaimImplementing || claim.Version != 3 {
		t.Fatalf("ListIssueClaims() claim = %+v", claim)
	}
	call := executor.calls[0]
	if !strings.Contains(call.query, "WHERE account_id = $1") ||
		call.arguments[0] != accountID.String() || call.arguments[2] != 10 {
		t.Fatalf("ListIssueClaims() query = %+v", call)
	}
}

func TestAccountRepositoryWritesIssueClaims(t *testing.T) {
	accountID := mustAccountID(t)
	claimID := mustResourceID(t)
	now := time.Date(2026, 8, 16, 1, 2, 3, 0, time.UTC)
	claim, _ := account.NewIssueClaim("OpenAI", "openai-go", 42)
	pullRequest, _ := account.NewPullRequestReference("OpenAI", "openai-go", 84)
	claim, _ = account.UpdateIssueClaim(
		claim,
		account.IssueClaimPRSubmitted,
		false,
		&pullRequest,
	)
	claim.ID = claimID
	claim.AccountID = accountID

	successRow := valueRow{values: []any{
		featureResourceID,
		"openai",
		"openai-go",
		42,
		"pr_submitted",
		false,
		"openai",
		"openai-go",
		84,
		"open",
		"open",
		int64(1),
		now,
		now,
	}}
	executor := &scriptedAccountExecutor{rowQueue: []pgx.Row{successRow}}
	repository := AccountRepository{executor: executor, queryTimeout: time.Second}

	result, err := repository.UpsertIssueClaim(context.Background(), claim)
	if err != nil || result.PullRequest == nil || result.PullRequest.Number != 84 {
		t.Fatalf("UpsertIssueClaim() = %+v, %v", result, err)
	}
	call := executor.calls[0]
	if !strings.Contains(call.query, "pg_advisory_xact_lock") ||
		call.arguments[7] == nil || call.arguments[9] == nil ||
		call.arguments[10] != int64(account.MaximumIssueClaims) {
		t.Fatalf("UpsertIssueClaim() query = %+v", call)
	}

	quotaRepository := AccountRepository{
		executor: &scriptedAccountExecutor{rowQueue: []pgx.Row{
			valueRow{err: pgx.ErrNoRows},
		}},
		queryTimeout: time.Second,
	}
	if _, err := quotaRepository.UpsertIssueClaim(
		context.Background(),
		claim,
	); !errors.Is(err, account.ErrQuotaExceeded) {
		t.Fatalf("quota UpsertIssueClaim() error = %v", err)
	}

	staleClaim := claim
	staleClaim.Version = 2
	staleRepository := AccountRepository{
		executor: &scriptedAccountExecutor{rowQueue: []pgx.Row{
			valueRow{err: pgx.ErrNoRows},
			valueRow{values: []any{int64(3)}},
		}},
		queryTimeout: time.Second,
	}
	if _, err := staleRepository.UpdateIssueClaim(
		context.Background(),
		staleClaim,
	); !errors.Is(err, account.ErrVersionConflict) {
		t.Fatalf("stale UpdateIssueClaim() error = %v", err)
	}
}

func TestAccountRepositoryDeletesIssueClaims(t *testing.T) {
	accountID := mustAccountID(t)
	claimID := mustResourceID(t)

	successExecutor := &scriptedAccountExecutor{commandTags: []pgconn.CommandTag{
		pgconn.NewCommandTag("DELETE 1"),
	}}
	successRepository := AccountRepository{
		executor: successExecutor, queryTimeout: time.Second,
	}
	if err := successRepository.DeleteIssueClaim(
		context.Background(), accountID, claimID, 3,
	); err != nil {
		t.Fatalf("DeleteIssueClaim() error = %v", err)
	}
	if !strings.Contains(successExecutor.calls[0].query, "account_id = $1") {
		t.Fatalf("DeleteIssueClaim() query = %q", successExecutor.calls[0].query)
	}

	missingRepository := AccountRepository{
		executor: &scriptedAccountExecutor{
			commandTags: []pgconn.CommandTag{pgconn.NewCommandTag("DELETE 0")},
			rowQueue:    []pgx.Row{valueRow{err: pgx.ErrNoRows}},
		},
		queryTimeout: time.Second,
	}
	if err := missingRepository.DeleteIssueClaim(
		context.Background(), accountID, claimID, 3,
	); !errors.Is(err, account.ErrNotFound) {
		t.Fatalf("missing DeleteIssueClaim() error = %v", err)
	}

	failureRepository := AccountRepository{
		executor:     &scriptedAccountExecutor{execError: errors.New("database unavailable")},
		queryTimeout: time.Second,
	}
	if err := failureRepository.DeleteIssueClaim(
		context.Background(), accountID, claimID, 3,
	); !errors.Is(err, ErrQueryFailed) {
		t.Fatalf("failed DeleteIssueClaim() error = %v", err)
	}
}

func TestAccountRepositoryListIssueClaimFailures(t *testing.T) {
	accountID := mustAccountID(t)
	page, _ := account.NewPage(1, 10)
	tests := []struct {
		name     string
		executor *scriptedAccountExecutor
	}{
		{
			name: "query",
			executor: &scriptedAccountExecutor{
				queryError: errors.New("database unavailable"),
			},
		},
		{
			name: "scan",
			executor: &scriptedAccountExecutor{
				queryRows: &valueRows{rows: []valueRow{{err: errors.New("bad row")}}},
			},
		},
		{
			name: "iteration",
			executor: &scriptedAccountExecutor{
				queryRows: &valueRows{err: errors.New("connection lost")},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := AccountRepository{
				executor: test.executor, queryTimeout: time.Second,
			}
			if _, err := repository.ListIssueClaims(
				context.Background(), accountID, page,
			); !errors.Is(err, ErrQueryFailed) {
				t.Fatalf("ListIssueClaims() error = %v", err)
			}
		})
	}
}
