package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/account"
)

// ListBookmarks returns only rows owned by accountID in a stable newest-first
// order with the UUID as a deterministic tie-breaker.
func (repository *AccountRepository) ListBookmarks(
	ctx context.Context,
	accountID account.ID,
	page account.Page,
) (account.PageResult[account.Bookmark], error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	rows, err := repository.executor.Query(
		queryContext,
		listBookmarksSQL,
		accountID.String(),
		page.PerPage,
		page.Offset(),
	)
	if err != nil {
		return account.PageResult[account.Bookmark]{}, ErrQueryFailed
	}
	defer rows.Close()
	result := account.PageResult[account.Bookmark]{
		Items: make([]account.Bookmark, 0, page.PerPage),
		Page:  page,
	}
	for rows.Next() {
		bookmark, total, scanErr := scanBookmark(rows, accountID)
		if scanErr != nil {
			return account.PageResult[account.Bookmark]{}, ErrQueryFailed
		}
		result.Items = append(result.Items, bookmark)
		result.Total = total
	}
	if rows.Err() != nil {
		return account.PageResult[account.Bookmark]{}, ErrQueryFailed
	}
	return result, nil
}

// UpsertBookmark inserts one normalized reference or returns the existing
// account-owned row. Duplicate writes are idempotent and do not increment the
// optimistic version.
func (repository *AccountRepository) UpsertBookmark(
	ctx context.Context,
	bookmark account.Bookmark,
) (account.Bookmark, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	var rawID string
	var rawTarget string
	var owner string
	var repositoryName string
	var issueNumber *int
	var note, collection string
	var tags []string
	var version int64
	var createdAt, updatedAt time.Time
	err := repository.executor.QueryRow(
		queryContext,
		upsertBookmarkSQL,
		bookmark.AccountID.String(),
		bookmark.ID.String(),
		string(bookmark.Reference.TargetType),
		bookmark.Reference.RepositoryOwner,
		bookmark.Reference.RepositoryName,
		bookmark.Reference.IssueNumber,
		account.MaximumBookmarks,
	).Scan(
		&rawID,
		&rawTarget,
		&owner,
		&repositoryName,
		&issueNumber,
		&note,
		&collection,
		&tags,
		&version,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return account.Bookmark{}, account.ErrQuotaExceeded
	}
	if err != nil {
		return account.Bookmark{}, ErrQueryFailed
	}
	return mapBookmark(
		rawID,
		rawTarget,
		owner,
		repositoryName,
		issueNumber,
		note,
		collection,
		tags,
		version,
		createdAt,
		updatedAt,
		bookmark.AccountID,
	)
}

// UpdateBookmark replaces bounded metadata on one owned optimistic version.
func (repository *AccountRepository) UpdateBookmark(
	ctx context.Context,
	bookmark account.Bookmark,
) (account.Bookmark, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	var rawID, rawTarget, owner, repositoryName string
	var issueNumber *int
	var note, collection string
	var tags []string
	var version int64
	var createdAt, updatedAt time.Time
	err := repository.executor.QueryRow(
		queryContext,
		updateBookmarkSQL,
		bookmark.AccountID.String(),
		bookmark.ID.String(),
		bookmark.Note,
		bookmark.Collection,
		bookmark.Tags,
		bookmark.Version,
	).Scan(
		&rawID, &rawTarget, &owner, &repositoryName, &issueNumber,
		&note, &collection, &tags, &version, &createdAt, &updatedAt,
	)
	if err == nil {
		return mapBookmark(rawID, rawTarget, owner, repositoryName, issueNumber,
			note, collection, tags, version, createdAt, updatedAt, bookmark.AccountID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return account.Bookmark{}, ErrQueryFailed
	}
	return account.Bookmark{}, repository.ownedVersionFailure(
		queryContext, "bookmarks", bookmark.AccountID, bookmark.ID,
	)
}

// DeleteBookmark removes an owned bookmark only when its version matches.
func (repository *AccountRepository) DeleteBookmark(
	ctx context.Context,
	accountID account.ID,
	bookmarkID account.ResourceID,
	version int64,
) error {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	command, err := repository.executor.Exec(
		queryContext,
		`DELETE FROM bookmarks
		 WHERE account_id = $1 AND id = $2 AND version = $3`,
		accountID.String(),
		bookmarkID.String(),
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
		"bookmarks",
		accountID,
		bookmarkID,
	)
}

// ListSavedSearches returns only account-owned filter documents with stable
// updated-at and UUID ordering.
func (repository *AccountRepository) ListSavedSearches(
	ctx context.Context,
	accountID account.ID,
	page account.Page,
) (account.PageResult[account.SavedSearch], error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	rows, err := repository.executor.Query(
		queryContext,
		listSavedSearchesSQL,
		accountID.String(),
		page.PerPage,
		page.Offset(),
	)
	if err != nil {
		return account.PageResult[account.SavedSearch]{}, ErrQueryFailed
	}
	defer rows.Close()
	result := account.PageResult[account.SavedSearch]{
		Items: make([]account.SavedSearch, 0, page.PerPage),
		Page:  page,
	}
	for rows.Next() {
		savedSearch, total, scanErr := scanSavedSearch(rows, accountID)
		if scanErr != nil {
			return account.PageResult[account.SavedSearch]{}, ErrQueryFailed
		}
		result.Items = append(result.Items, savedSearch)
		result.Total = total
	}
	if rows.Err() != nil {
		return account.PageResult[account.SavedSearch]{}, ErrQueryFailed
	}
	return result, nil
}

// CreateSavedSearch inserts a bounded normalized filter document.
func (repository *AccountRepository) CreateSavedSearch(
	ctx context.Context,
	savedSearch account.SavedSearch,
) (account.SavedSearch, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	result, err := repository.writeSavedSearch(
		queryContext,
		createSavedSearchSQL,
		savedSearch,
		account.MaximumSavedSearches,
	)
	if !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	var duplicate bool
	if classifyErr := repository.executor.QueryRow(
		queryContext,
		`SELECT EXISTS (
		    SELECT 1 FROM saved_searches
		    WHERE account_id = $1 AND lower(name) = lower($2)
		)`,
		savedSearch.AccountID.String(),
		savedSearch.Name,
	).Scan(&duplicate); classifyErr != nil {
		return account.SavedSearch{}, ErrQueryFailed
	}
	if duplicate {
		return account.SavedSearch{}, account.ErrDuplicateSavedSearch
	}
	return account.SavedSearch{}, account.ErrQuotaExceeded
}

// UpdateSavedSearch updates an owned row only when its version is current.
func (repository *AccountRepository) UpdateSavedSearch(
	ctx context.Context,
	savedSearch account.SavedSearch,
) (account.SavedSearch, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	result, err := repository.writeSavedSearch(
		queryContext,
		updateSavedSearchSQL,
		savedSearch,
		savedSearch.Version,
	)
	if !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	failure := repository.ownedVersionFailure(
		queryContext,
		"saved_searches",
		savedSearch.AccountID,
		savedSearch.ID,
	)
	if !errors.Is(failure, account.ErrVersionConflict) {
		return account.SavedSearch{}, failure
	}
	var duplicate bool
	if classifyErr := repository.executor.QueryRow(
		queryContext,
		`SELECT EXISTS (
		    SELECT 1 FROM saved_searches
		    WHERE account_id = $1
		      AND lower(name) = lower($2)
		      AND id <> $3
		)`,
		savedSearch.AccountID.String(),
		savedSearch.Name,
		savedSearch.ID.String(),
	).Scan(&duplicate); classifyErr != nil {
		return account.SavedSearch{}, ErrQueryFailed
	}
	if duplicate {
		return account.SavedSearch{}, account.ErrDuplicateSavedSearch
	}
	return account.SavedSearch{}, account.ErrVersionConflict
}

// UpdateSavedSearchSnapshot replaces one explicit-check baseline without
// changing the saved filter or display name.
func (repository *AccountRepository) UpdateSavedSearchSnapshot(
	ctx context.Context,
	savedSearch account.SavedSearch,
) (account.SavedSearch, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	resultKeys, err := json.Marshal(savedSearch.ResultKeys)
	if err != nil {
		return account.SavedSearch{}, ErrQueryFailed
	}
	row := repository.executor.QueryRow(
		queryContext,
		updateSavedSearchSnapshotSQL,
		savedSearch.AccountID.String(),
		savedSearch.ID.String(),
		resultKeys,
		savedSearch.LastCheckedAt,
		savedSearch.Version,
	)
	result, scanErr := scanSavedSearchRow(row, savedSearch.AccountID)
	if scanErr == nil {
		return result, nil
	}
	if !errors.Is(scanErr, pgx.ErrNoRows) {
		return account.SavedSearch{}, ErrQueryFailed
	}
	return account.SavedSearch{}, repository.ownedVersionFailure(
		queryContext, "saved_searches", savedSearch.AccountID, savedSearch.ID,
	)
}

// DeleteSavedSearch removes an owned row only when its version matches.
func (repository *AccountRepository) DeleteSavedSearch(
	ctx context.Context,
	accountID account.ID,
	savedSearchID account.ResourceID,
	version int64,
) error {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	command, err := repository.executor.Exec(
		queryContext,
		`DELETE FROM saved_searches
		 WHERE account_id = $1 AND id = $2 AND version = $3`,
		accountID.String(),
		savedSearchID.String(),
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
		"saved_searches",
		accountID,
		savedSearchID,
	)
}

// GetPreferences returns stored preferences or deterministic defaults with
// version zero when the account has not persisted a preference row.
func (repository *AccountRepository) GetPreferences(
	ctx context.Context,
	accountID account.ID,
) (account.Preferences, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	preferences := account.Preferences{AccountID: accountID}
	var theme string
	var reducedMotion string
	var createdAt, updatedAt time.Time
	err := repository.executor.QueryRow(
		queryContext,
		`SELECT theme, reduced_motion, results_per_page, version,
		        created_at, updated_at
		 FROM user_preferences
		 WHERE account_id = $1`,
		accountID.String(),
	).Scan(
		&theme,
		&reducedMotion,
		&preferences.ResultsPerPage,
		&preferences.Version,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		preferences.Theme = account.ThemeSystem
		preferences.ReducedMotion = account.ReducedMotionSystem
		preferences.ResultsPerPage = account.DefaultPageSize
		return preferences, nil
	}
	if err != nil {
		return account.Preferences{}, ErrQueryFailed
	}
	preferences.Theme = account.Theme(theme)
	preferences.ReducedMotion = account.ReducedMotion(reducedMotion)
	preferences.CreatedAt = createdAt
	preferences.UpdatedAt = updatedAt
	return preferences, nil
}

// UpsertPreferences creates defaults version one or updates an existing row
// only when expectedVersion matches.
func (repository *AccountRepository) UpsertPreferences(
	ctx context.Context,
	preferences account.Preferences,
	expectedVersion int64,
) (account.Preferences, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	var theme string
	var reducedMotion string
	var createdAt, updatedAt time.Time
	err := repository.executor.QueryRow(
		queryContext,
		upsertPreferencesSQL,
		preferences.AccountID.String(),
		string(preferences.Theme),
		string(preferences.ReducedMotion),
		preferences.ResultsPerPage,
		expectedVersion,
	).Scan(
		&theme,
		&reducedMotion,
		&preferences.ResultsPerPage,
		&preferences.Version,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return account.Preferences{}, account.ErrVersionConflict
	}
	if err != nil {
		return account.Preferences{}, ErrQueryFailed
	}
	preferences.Theme = account.Theme(theme)
	preferences.ReducedMotion = account.ReducedMotion(reducedMotion)
	preferences.CreatedAt = createdAt
	preferences.UpdatedAt = updatedAt
	return preferences, nil
}

func (repository *AccountRepository) writeSavedSearch(
	ctx context.Context,
	query string,
	savedSearch account.SavedSearch,
	finalArgument int64,
) (account.SavedSearch, error) {
	var rawID string
	var rawType string
	var name string
	var filters []byte
	var resultKeys []byte
	var lastCheckedAt *time.Time
	var version int64
	var createdAt, updatedAt time.Time
	err := repository.executor.QueryRow(
		ctx,
		query,
		savedSearch.AccountID.String(),
		savedSearch.ID.String(),
		string(savedSearch.SearchType),
		savedSearch.Name,
		[]byte(savedSearch.Filters),
		finalArgument,
	).Scan(
		&rawID,
		&rawType,
		&name,
		&filters,
		&resultKeys,
		&lastCheckedAt,
		&version,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return account.SavedSearch{}, pgx.ErrNoRows
	}
	if err != nil {
		return account.SavedSearch{}, ErrQueryFailed
	}
	return mapSavedSearch(rawID, rawType, name, filters, resultKeys,
		lastCheckedAt, version, createdAt, updatedAt, savedSearch.AccountID)
}

func (repository *AccountRepository) ownedVersionFailure(
	ctx context.Context,
	table string,
	accountID account.ID,
	resourceID account.ResourceID,
) error {
	query := ""
	switch table {
	case "bookmarks":
		query = "SELECT version FROM bookmarks WHERE account_id = $1 AND id = $2"
	case "saved_searches":
		query = "SELECT version FROM saved_searches WHERE account_id = $1 AND id = $2"
	case "issue_claims":
		query = "SELECT version FROM issue_claims WHERE account_id = $1 AND id = $2"
	default:
		return ErrQueryFailed
	}
	var currentVersion int64
	err := repository.executor.QueryRow(
		ctx,
		query,
		accountID.String(),
		resourceID.String(),
	).Scan(&currentVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return account.ErrNotFound
	}
	if err != nil {
		return ErrQueryFailed
	}
	return account.ErrVersionConflict
}

type rowScanner interface {
	Scan(destinations ...any) error
}

func scanBookmark(
	row rowScanner,
	accountID account.ID,
) (account.Bookmark, int, error) {
	var rawID string
	var rawTarget string
	var owner string
	var repositoryName string
	var issueNumber *int
	var note, collection string
	var tags []string
	var version int64
	var createdAt, updatedAt time.Time
	var total int
	if err := row.Scan(
		&rawID,
		&rawTarget,
		&owner,
		&repositoryName,
		&issueNumber,
		&note,
		&collection,
		&tags,
		&version,
		&createdAt,
		&updatedAt,
		&total,
	); err != nil {
		return account.Bookmark{}, 0, err
	}
	bookmark, err := mapBookmark(
		rawID,
		rawTarget,
		owner,
		repositoryName,
		issueNumber,
		note,
		collection,
		tags,
		version,
		createdAt,
		updatedAt,
		accountID,
	)
	return bookmark, total, err
}

func mapBookmark(
	rawID string,
	rawTarget string,
	owner string,
	repositoryName string,
	issueNumber *int,
	note string,
	collection string,
	tags []string,
	version int64,
	createdAt time.Time,
	updatedAt time.Time,
	accountID account.ID,
) (account.Bookmark, error) {
	id, err := account.ParseResourceID(rawID)
	if err != nil {
		return account.Bookmark{}, err
	}
	reference, err := account.NewBookmarkReference(
		account.BookmarkTarget(rawTarget),
		owner,
		repositoryName,
		issueNumber,
	)
	if err != nil {
		return account.Bookmark{}, err
	}
	return account.Bookmark{
		ID:         id,
		AccountID:  accountID,
		Reference:  reference,
		Note:       note,
		Collection: collection,
		Tags:       append([]string{}, tags...),
		Version:    version,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func scanSavedSearch(
	row rowScanner,
	accountID account.ID,
) (account.SavedSearch, int, error) {
	var rawID string
	var rawType string
	var name string
	var filters []byte
	var resultKeys []byte
	var lastCheckedAt *time.Time
	var version int64
	var createdAt, updatedAt time.Time
	var total int
	if err := row.Scan(
		&rawID,
		&rawType,
		&name,
		&filters,
		&resultKeys,
		&lastCheckedAt,
		&version,
		&createdAt,
		&updatedAt,
		&total,
	); err != nil {
		return account.SavedSearch{}, 0, err
	}
	savedSearch, err := mapSavedSearch(rawID, rawType, name, filters, resultKeys,
		lastCheckedAt, version, createdAt, updatedAt, accountID)
	return savedSearch, total, err
}

func scanSavedSearchRow(row rowScanner, accountID account.ID) (account.SavedSearch, error) {
	var rawID, rawType, name string
	var filters, resultKeys []byte
	var lastCheckedAt *time.Time
	var version int64
	var createdAt, updatedAt time.Time
	if err := row.Scan(&rawID, &rawType, &name, &filters, &resultKeys,
		&lastCheckedAt, &version, &createdAt, &updatedAt); err != nil {
		return account.SavedSearch{}, err
	}
	return mapSavedSearch(rawID, rawType, name, filters, resultKeys,
		lastCheckedAt, version, createdAt, updatedAt, accountID)
}

func mapSavedSearch(
	rawID, rawType, name string,
	filters, rawResultKeys []byte,
	lastCheckedAt *time.Time,
	version int64,
	createdAt, updatedAt time.Time,
	accountID account.ID,
) (account.SavedSearch, error) {
	id, err := account.ParseResourceID(rawID)
	if err != nil || !json.Valid(filters) || !json.Valid(rawResultKeys) {
		return account.SavedSearch{}, ErrQueryFailed
	}
	var resultKeys []string
	if err := json.Unmarshal(rawResultKeys, &resultKeys); err != nil {
		return account.SavedSearch{}, ErrQueryFailed
	}
	return account.SavedSearch{
		ID: id, AccountID: accountID, SearchType: account.SearchType(rawType),
		Name: name, Filters: json.RawMessage(filters), ResultKeys: resultKeys,
		LastCheckedAt: lastCheckedAt, Version: version,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

const listBookmarksSQL = `SELECT
    id,
    target_type,
    repository_owner,
    repository_name,
    issue_number,
    note,
    collection_name,
    tags,
    version,
    created_at,
    updated_at,
    count(*) OVER ()
FROM bookmarks
WHERE account_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3`

const upsertBookmarkSQL = `WITH account_lock AS (
    SELECT pg_advisory_xact_lock(hashtextextended($1::uuid::text, 0))
),
allowed AS (
    SELECT 1
    FROM account_lock
    WHERE EXISTS (
        SELECT 1
        FROM bookmarks
        WHERE account_id = $1::uuid
          AND target_type = $3
          AND repository_owner = $4
          AND lower(repository_name) = lower($5)
          AND issue_number IS NOT DISTINCT FROM $6
    )
    OR (
        SELECT count(*) FROM bookmarks WHERE account_id = $1::uuid
    ) < $7
),
inserted AS (
    INSERT INTO bookmarks (
        id,
        account_id,
        target_type,
        repository_owner,
        repository_name,
        issue_number
    )
    SELECT $2::uuid, $1::uuid, $3, $4, $5, $6
    FROM allowed
    ON CONFLICT DO NOTHING
    RETURNING id, target_type, repository_owner, repository_name,
              issue_number, note, collection_name, tags, version, created_at, updated_at
),
existing AS (
    SELECT id, target_type, repository_owner, repository_name,
           issue_number, note, collection_name, tags, version, created_at, updated_at
    FROM bookmarks
    WHERE account_id = $1::uuid
      AND target_type = $3
      AND repository_owner = $4
      AND lower(repository_name) = lower($5)
      AND issue_number IS NOT DISTINCT FROM $6
)
SELECT * FROM inserted
UNION ALL
SELECT * FROM existing
LIMIT 1`

const updateBookmarkSQL = `UPDATE bookmarks SET
    note = $3,
    collection_name = $4,
    tags = $5,
    version = version + 1,
    updated_at = now()
WHERE account_id = $1 AND id = $2 AND version = $6
RETURNING id, target_type, repository_owner, repository_name,
          issue_number, note, collection_name, tags, version, created_at, updated_at`

const listSavedSearchesSQL = `SELECT
    id,
    search_type,
    name,
    filters,
    result_keys,
    last_checked_at,
    version,
    created_at,
    updated_at,
    count(*) OVER ()
FROM saved_searches
WHERE account_id = $1
ORDER BY updated_at DESC, id DESC
LIMIT $2 OFFSET $3`

const createSavedSearchSQL = `WITH account_lock AS (
    SELECT pg_advisory_xact_lock(hashtextextended($1::uuid::text, 1))
),
allowed AS (
    SELECT 1
    FROM account_lock
    WHERE (
        SELECT count(*) FROM saved_searches WHERE account_id = $1::uuid
    ) < $6
)
INSERT INTO saved_searches (
    id,
    account_id,
    search_type,
    name,
    filters
)
SELECT $2::uuid, $1::uuid, $3, $4, $5
FROM allowed
ON CONFLICT DO NOTHING
RETURNING id, search_type, name, filters, result_keys, last_checked_at,
          version, created_at, updated_at`

const updateSavedSearchSQL = `UPDATE saved_searches
SET search_type = $3,
    name = $4,
    filters = $5,
    version = version + 1,
    updated_at = now()
WHERE account_id = $1
  AND id = $2
  AND version = $6
  AND NOT EXISTS (
      SELECT 1
      FROM saved_searches duplicate
      WHERE duplicate.account_id = $1
        AND lower(duplicate.name) = lower($4)
        AND duplicate.id <> $2
  )
RETURNING id, search_type, name, filters, result_keys, last_checked_at,
          version, created_at, updated_at`

const updateSavedSearchSnapshotSQL = `UPDATE saved_searches SET
    result_keys = $3,
    last_checked_at = $4,
    version = version + 1,
    updated_at = now()
WHERE account_id = $1 AND id = $2 AND version = $5
RETURNING id, search_type, name, filters, result_keys, last_checked_at,
          version, created_at, updated_at`

const upsertPreferencesSQL = `INSERT INTO user_preferences (
    account_id,
    theme,
    reduced_motion,
    results_per_page
)
VALUES ($1, $2, $3, $4)
ON CONFLICT (account_id) DO UPDATE
SET theme = EXCLUDED.theme,
    reduced_motion = EXCLUDED.reduced_motion,
    results_per_page = EXCLUDED.results_per_page,
    version = user_preferences.version + 1,
    updated_at = now()
WHERE user_preferences.version = $5
RETURNING theme, reduced_motion, results_per_page, version,
          created_at, updated_at`
