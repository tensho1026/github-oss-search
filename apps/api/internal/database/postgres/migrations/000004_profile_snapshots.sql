CREATE TABLE profile_snapshots (
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    month date NOT NULL,
    languages jsonb NOT NULL CHECK (jsonb_typeof(languages) = 'array'),
    frameworks jsonb NOT NULL CHECK (jsonb_typeof(frameworks) = 'array'),
    oss_activity integer NOT NULL CHECK (oss_activity BETWEEN 0 AND 1000000),
    merged_pull_requests integer NOT NULL CHECK (merged_pull_requests BETWEEN 0 AND 1000000),
    proficiency jsonb NOT NULL CHECK (jsonb_typeof(proficiency) = 'array'),
    completed_quests integer NOT NULL CHECK (completed_quests BETWEEN 0 AND 1000000),
    current_streak integer NOT NULL CHECK (current_streak BETWEEN 0 AND 1000000),
    longest_streak integer NOT NULL CHECK (longest_streak BETWEEN 0 AND 1000000),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, month),
    CHECK (month = date_trunc('month', month)::date),
    CHECK (octet_length(languages::text) <= 4096),
    CHECK (octet_length(frameworks::text) <= 4096),
    CHECK (octet_length(proficiency::text) <= 8192)
);

CREATE INDEX profile_snapshots_account_month_idx
    ON profile_snapshots (account_id, month DESC);
