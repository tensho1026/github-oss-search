ALTER TABLE saved_searches
    ADD COLUMN result_keys jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN last_checked_at timestamptz;

ALTER TABLE saved_searches
    ADD CONSTRAINT saved_searches_result_keys_check CHECK (
        jsonb_typeof(result_keys) = 'array'
        AND jsonb_array_length(result_keys) <= 50
        AND octet_length(result_keys::text) <= 12288
    );
