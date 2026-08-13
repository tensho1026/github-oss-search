CREATE TABLE issue_claims (
    id uuid PRIMARY KEY,
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    repository_owner varchar(39) NOT NULL CHECK (
        repository_owner = lower(repository_owner)
    ),
    repository_name varchar(100) NOT NULL,
    issue_number integer NOT NULL CHECK (issue_number > 0),
    workflow_status text NOT NULL DEFAULT 'not_started' CHECK (
        workflow_status IN (
            'not_started',
            'researching',
            'implementing',
            'pr_submitted',
            'merged'
        )
    ),
    archived boolean NOT NULL DEFAULT false,
    pull_request_owner varchar(39) CHECK (
        pull_request_owner IS NULL
        OR pull_request_owner = lower(pull_request_owner)
    ),
    pull_request_repository varchar(100),
    pull_request_number integer CHECK (pull_request_number > 0),
    observed_issue_state text NOT NULL DEFAULT 'unverified' CHECK (
        observed_issue_state IN (
            'unverified', 'open', 'closed', 'missing', 'inaccessible'
        )
    ),
    observed_pr_state text NOT NULL DEFAULT 'unverified' CHECK (
        observed_pr_state IN (
            'unverified', 'open', 'closed', 'merged', 'missing', 'inaccessible'
        )
    ),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        (pull_request_owner IS NULL
            AND pull_request_repository IS NULL
            AND pull_request_number IS NULL)
        OR (pull_request_owner IS NOT NULL
            AND pull_request_repository IS NOT NULL
            AND pull_request_number IS NOT NULL)
    ),
    CHECK (
        workflow_status NOT IN ('pr_submitted', 'merged')
        OR pull_request_number IS NOT NULL
    )
);

CREATE UNIQUE INDEX issue_claims_reference_unique_idx
    ON issue_claims (
        account_id,
        repository_owner,
        lower(repository_name),
        issue_number
    );

CREATE INDEX issue_claims_account_order_idx
    ON issue_claims (account_id, archived, updated_at DESC, id DESC);
