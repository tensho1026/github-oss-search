package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/account"
)

// ListProfileSnapshots returns the oldest-to-newest bounded monthly history.
func (repository *AccountRepository) ListProfileSnapshots(
	ctx context.Context,
	accountID account.ID,
) ([]account.ProfileSnapshot, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	rows, err := repository.executor.Query(queryContext, `
		SELECT month, languages, frameworks, oss_activity,
		       merged_pull_requests, proficiency, completed_quests,
		       current_streak, longest_streak, created_at, updated_at
		FROM profile_snapshots
		WHERE account_id = $1
		ORDER BY month ASC
		LIMIT $2`, accountID.String(), account.MaximumProfileSnapshots)
	if err != nil {
		return nil, ErrQueryFailed
	}
	defer rows.Close()
	result := make([]account.ProfileSnapshot, 0, account.MaximumProfileSnapshots)
	for rows.Next() {
		snapshot, scanErr := scanProfileSnapshot(rows, accountID)
		if scanErr != nil {
			return nil, ErrQueryFailed
		}
		result = append(result, snapshot)
	}
	if rows.Err() != nil {
		return nil, ErrQueryFailed
	}
	return result, nil
}

// UpsertProfileSnapshot atomically retains at most 24 months and replaces the
// current calendar month for the authenticated account.
func (repository *AccountRepository) UpsertProfileSnapshot(
	ctx context.Context,
	snapshot account.ProfileSnapshot,
) (account.ProfileSnapshot, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	languages, err := json.Marshal(snapshot.Languages)
	if err != nil {
		return account.ProfileSnapshot{}, ErrQueryFailed
	}
	frameworks, err := json.Marshal(snapshot.Frameworks)
	if err != nil {
		return account.ProfileSnapshot{}, ErrQueryFailed
	}
	proficiency, err := json.Marshal(snapshot.Proficiency)
	if err != nil {
		return account.ProfileSnapshot{}, ErrQueryFailed
	}
	row := repository.executor.QueryRow(queryContext, `
		WITH pruned AS (
			DELETE FROM profile_snapshots
			WHERE account_id = $1
			  AND month = (SELECT min(month) FROM profile_snapshots WHERE account_id = $1)
			  AND (SELECT count(*) FROM profile_snapshots WHERE account_id = $1) >= $11
			  AND NOT EXISTS (SELECT 1 FROM profile_snapshots WHERE account_id = $1 AND month = $2)
		), written AS (
			INSERT INTO profile_snapshots (
				account_id, month, languages, frameworks, oss_activity,
				merged_pull_requests, proficiency, completed_quests,
				current_streak, longest_streak
			) SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			WHERE EXISTS (SELECT 1 FROM accounts WHERE id = $1)
			ON CONFLICT (account_id, month) DO UPDATE SET
				languages = EXCLUDED.languages,
				frameworks = EXCLUDED.frameworks,
				oss_activity = EXCLUDED.oss_activity,
				merged_pull_requests = EXCLUDED.merged_pull_requests,
				proficiency = EXCLUDED.proficiency,
				completed_quests = EXCLUDED.completed_quests,
				current_streak = EXCLUDED.current_streak,
				longest_streak = EXCLUDED.longest_streak,
				updated_at = now()
			RETURNING month, languages, frameworks, oss_activity,
			          merged_pull_requests, proficiency, completed_quests,
			          current_streak, longest_streak, created_at, updated_at
		)
		SELECT * FROM written`,
		snapshot.AccountID.String(), snapshot.Month, languages, frameworks,
		snapshot.OSSActivity, snapshot.MergedPullRequests, proficiency,
		snapshot.CompletedQuests, snapshot.CurrentStreak, snapshot.LongestStreak,
		account.MaximumProfileSnapshots,
	)
	result, err := scanProfileSnapshot(row, snapshot.AccountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return account.ProfileSnapshot{}, account.ErrNotFound
	}
	if err != nil {
		return account.ProfileSnapshot{}, ErrQueryFailed
	}
	return result, nil
}

func scanProfileSnapshot(row rowScanner, accountID account.ID) (account.ProfileSnapshot, error) {
	var month, createdAt, updatedAt time.Time
	var languagesJSON, frameworksJSON, proficiencyJSON []byte
	var ossActivity, merged, completed, current, longest int
	if err := row.Scan(&month, &languagesJSON, &frameworksJSON, &ossActivity, &merged, &proficiencyJSON, &completed, &current, &longest, &createdAt, &updatedAt); err != nil {
		return account.ProfileSnapshot{}, err
	}
	var languages, frameworks []string
	var proficiency []account.SnapshotProficiency
	if json.Unmarshal(languagesJSON, &languages) != nil || json.Unmarshal(frameworksJSON, &frameworks) != nil || json.Unmarshal(proficiencyJSON, &proficiency) != nil {
		return account.ProfileSnapshot{}, ErrQueryFailed
	}
	snapshot, err := account.NewProfileSnapshot(accountID, languages, frameworks, ossActivity, merged, proficiency, completed, current, longest, month)
	if err != nil {
		return account.ProfileSnapshot{}, err
	}
	snapshot.CreatedAt = createdAt.UTC()
	snapshot.UpdatedAt = updatedAt.UTC()
	return snapshot, nil
}
