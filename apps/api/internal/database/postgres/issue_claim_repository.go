package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/account"
)

// ListIssueClaims returns only account-owned tasks ordered by active state and
// recent user updates. Window counts provide one progress summary.
func (repository *AccountRepository) ListIssueClaims(
	ctx context.Context,
	accountID account.ID,
	page account.Page,
) (account.IssueClaimPage, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	rows, err := repository.executor.Query(
		queryContext,
		listIssueClaimsSQL,
		accountID.String(),
		page.PerPage,
		page.Offset(),
	)
	if err != nil {
		return account.IssueClaimPage{}, ErrQueryFailed
	}
	defer rows.Close()
	result := account.IssueClaimPage{PageResult: account.PageResult[account.IssueClaim]{
		Items: make([]account.IssueClaim, 0, page.PerPage),
		Page:  page,
	}}
	for rows.Next() {
		claim, summary, scanErr := scanIssueClaim(rows, accountID)
		if scanErr != nil {
			return account.IssueClaimPage{}, ErrQueryFailed
		}
		result.Items = append(result.Items, claim)
		result.Total = summary.Total
		result.Summary = summary
	}
	if rows.Err() != nil {
		return account.IssueClaimPage{}, ErrQueryFailed
	}
	return result, nil
}

// UpsertIssueClaim inserts one task or returns the existing account-owned
// canonical reference without incrementing its version.
func (repository *AccountRepository) UpsertIssueClaim(
	ctx context.Context,
	claim account.IssueClaim,
) (account.IssueClaim, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	result, err := repository.writeIssueClaim(
		queryContext,
		upsertIssueClaimSQL,
		claim,
		account.MaximumIssueClaims,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return account.IssueClaim{}, account.ErrQuotaExceeded
	}
	return result, err
}

// UpdateIssueClaim changes only an owned row whose optimistic version matches.
func (repository *AccountRepository) UpdateIssueClaim(
	ctx context.Context,
	claim account.IssueClaim,
) (account.IssueClaim, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	result, err := repository.writeIssueClaim(
		queryContext,
		updateIssueClaimSQL,
		claim,
		claim.Version,
	)
	if !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	return account.IssueClaim{}, repository.ownedVersionFailure(
		queryContext,
		"issue_claims",
		claim.AccountID,
		claim.ID,
	)
}

// DeleteIssueClaim removes only the owned ID and matching version.
func (repository *AccountRepository) DeleteIssueClaim(
	ctx context.Context,
	accountID account.ID,
	claimID account.ResourceID,
	version int64,
) error {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	command, err := repository.executor.Exec(
		queryContext,
		`DELETE FROM issue_claims
		 WHERE account_id = $1 AND id = $2 AND version = $3`,
		accountID.String(),
		claimID.String(),
		version,
	)
	if err != nil {
		return ErrQueryFailed
	}
	if command.RowsAffected() == 1 {
		return nil
	}
	return repository.ownedVersionFailure(
		queryContext,
		"issue_claims",
		accountID,
		claimID,
	)
}

func (repository *AccountRepository) writeIssueClaim(
	ctx context.Context,
	query string,
	claim account.IssueClaim,
	finalArgument int64,
) (account.IssueClaim, error) {
	var pullOwner, pullRepository *string
	var pullNumber *int
	if claim.PullRequest != nil {
		pullOwner = &claim.PullRequest.RepositoryOwner
		pullRepository = &claim.PullRequest.RepositoryName
		pullNumber = &claim.PullRequest.Number
	}
	row := repository.executor.QueryRow(
		ctx,
		query,
		claim.AccountID.String(),
		claim.ID.String(),
		claim.Issue.RepositoryOwner,
		claim.Issue.RepositoryName,
		claim.Issue.IssueNumber,
		string(claim.Status),
		claim.Archived,
		pullOwner,
		pullRepository,
		pullNumber,
		finalArgument,
	)
	result, err := scanIssueClaimRow(row, claim.AccountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return account.IssueClaim{}, pgx.ErrNoRows
	}
	if err != nil {
		return account.IssueClaim{}, ErrQueryFailed
	}
	return result, nil
}

func scanIssueClaim(
	row rowScanner,
	accountID account.ID,
) (account.IssueClaim, account.IssueClaimSummary, error) {
	claim, err := scanIssueClaimValues(row, accountID, true)
	if err != nil {
		return account.IssueClaim{}, account.IssueClaimSummary{}, err
	}
	return claim.claim, claim.summary, nil
}

func scanIssueClaimRow(
	row rowScanner,
	accountID account.ID,
) (account.IssueClaim, error) {
	claim, err := scanIssueClaimValues(row, accountID, false)
	return claim.claim, err
}

type scannedIssueClaim struct {
	claim   account.IssueClaim
	summary account.IssueClaimSummary
}

func scanIssueClaimValues(
	row rowScanner,
	accountID account.ID,
	withSummary bool,
) (scannedIssueClaim, error) {
	var rawID, owner, repositoryName, rawStatus string
	var issueNumber int
	var archived bool
	var pullOwner, pullRepository *string
	var pullNumber *int
	var issueState, pullState string
	var version int64
	var createdAt, updatedAt time.Time
	values := []any{
		&rawID, &owner, &repositoryName, &issueNumber, &rawStatus, &archived,
		&pullOwner, &pullRepository, &pullNumber, &issueState, &pullState,
		&version, &createdAt, &updatedAt,
	}
	result := scannedIssueClaim{}
	if withSummary {
		values = append(values,
			&result.summary.Total,
			&result.summary.NotStarted,
			&result.summary.Researching,
			&result.summary.Implementing,
			&result.summary.PRSubmitted,
			&result.summary.Merged,
			&result.summary.Archived,
		)
	}
	if err := row.Scan(values...); err != nil {
		return scannedIssueClaim{}, err
	}
	id, err := account.ParseResourceID(rawID)
	if err != nil {
		return scannedIssueClaim{}, err
	}
	claim, err := account.NewIssueClaim(owner, repositoryName, issueNumber)
	if err != nil {
		return scannedIssueClaim{}, err
	}
	claim.ID = id
	claim.AccountID = accountID
	claim.Status = account.IssueClaimStatus(rawStatus)
	claim.Archived = archived
	claim.ObservedIssueState = account.UpstreamReferenceState(issueState)
	claim.ObservedPRState = account.UpstreamReferenceState(pullState)
	claim.Version = version
	claim.CreatedAt = createdAt
	claim.UpdatedAt = updatedAt
	if pullOwner != nil && pullRepository != nil && pullNumber != nil {
		reference, referenceErr := account.NewPullRequestReference(
			*pullOwner,
			*pullRepository,
			*pullNumber,
		)
		if referenceErr != nil {
			return scannedIssueClaim{}, referenceErr
		}
		claim.PullRequest = &reference
	}
	if _, err := account.UpdateIssueClaim(
		claim,
		claim.Status,
		claim.Archived,
		claim.PullRequest,
	); err != nil {
		return scannedIssueClaim{}, err
	}
	result.claim = claim
	return result, nil
}

const issueClaimColumns = `id, repository_owner, repository_name,
    issue_number, workflow_status, archived,
    pull_request_owner, pull_request_repository, pull_request_number,
    observed_issue_state, observed_pr_state, version, created_at, updated_at`

const listIssueClaimsSQL = `SELECT ` + issueClaimColumns + `,
    count(*) OVER (),
    count(*) FILTER (WHERE workflow_status = 'not_started') OVER (),
    count(*) FILTER (WHERE workflow_status = 'researching') OVER (),
    count(*) FILTER (WHERE workflow_status = 'implementing') OVER (),
    count(*) FILTER (WHERE workflow_status = 'pr_submitted') OVER (),
    count(*) FILTER (WHERE workflow_status = 'merged') OVER (),
    count(*) FILTER (WHERE archived) OVER ()
FROM issue_claims
WHERE account_id = $1
ORDER BY archived, updated_at DESC, id DESC
LIMIT $2 OFFSET $3`

const upsertIssueClaimSQL = `WITH account_lock AS (
    SELECT pg_advisory_xact_lock(hashtextextended($1::uuid::text, 2))
), allowed AS (
    SELECT 1 FROM account_lock
    WHERE EXISTS (
        SELECT 1 FROM issue_claims
        WHERE account_id = $1::uuid
          AND repository_owner = $3
          AND lower(repository_name) = lower($4)
          AND issue_number = $5
    ) OR (SELECT count(*) FROM issue_claims WHERE account_id = $1::uuid) < $11
), inserted AS (
    INSERT INTO issue_claims (
        id, account_id, repository_owner, repository_name, issue_number
    )
    SELECT $2::uuid, $1::uuid, $3, $4, $5 FROM allowed
    ON CONFLICT DO NOTHING
    RETURNING ` + issueClaimColumns + `
), existing AS (
    SELECT ` + issueClaimColumns + ` FROM issue_claims
    WHERE account_id = $1::uuid
      AND repository_owner = $3
      AND lower(repository_name) = lower($4)
      AND issue_number = $5
)
SELECT * FROM inserted
UNION ALL
SELECT * FROM existing
LIMIT 1`

const updateIssueClaimSQL = `UPDATE issue_claims SET
    workflow_status = $6,
    archived = $7,
    pull_request_owner = $8,
    pull_request_repository = $9,
    pull_request_number = $10,
    observed_pr_state = CASE WHEN $10::integer IS NULL
        THEN 'unverified' ELSE observed_pr_state END,
    version = version + 1,
    updated_at = now()
WHERE account_id = $1 AND id = $2 AND version = $11
RETURNING ` + issueClaimColumns
