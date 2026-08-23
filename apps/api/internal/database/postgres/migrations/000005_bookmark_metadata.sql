ALTER TABLE bookmarks
    ADD COLUMN note varchar(2000) NOT NULL DEFAULT '',
    ADD COLUMN collection_name varchar(320) NOT NULL DEFAULT '',
    ADD COLUMN tags text[] NOT NULL DEFAULT '{}';

ALTER TABLE bookmarks
    ADD CONSTRAINT bookmarks_tags_count_check
        CHECK (cardinality(tags) <= 10),
    ADD CONSTRAINT bookmarks_note_bytes_check
        CHECK (octet_length(note) <= 2000),
    ADD CONSTRAINT bookmarks_collection_bytes_check
        CHECK (octet_length(collection_name) <= 320);
