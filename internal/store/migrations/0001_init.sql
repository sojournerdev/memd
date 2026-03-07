PRAGMA foreign_keys = ON;

CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT '',
    checksum TEXT NOT NULL,
    created_at_ns INTEGER NOT NULL CHECK (created_at_ns > 0),
    updated_at_ns INTEGER NOT NULL CHECK (updated_at_ns > 0),
    archived_at_ns INTEGER NOT NULL DEFAULT 0 CHECK (archived_at_ns >= 0),
    CHECK (kind IN ('note', 'conversation', 'fact', 'summary', 'artifact')),
    CHECK (length(id) > 0),
    CHECK (length(checksum) > 0),
    CHECK (updated_at_ns >= created_at_ns)
) WITHOUT ROWID;

CREATE TABLE memory_chunks (
    id TEXT PRIMARY KEY,
    memory_id TEXT NOT NULL,
    ord INTEGER NOT NULL CHECK (ord >= 0),
    content TEXT NOT NULL,
    token_count INTEGER NOT NULL DEFAULT 0 CHECK (token_count >= 0),
    created_at_ns INTEGER NOT NULL CHECK (created_at_ns > 0),
    FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE,
    UNIQUE (memory_id, ord),
    CHECK (length(id) > 0)
) WITHOUT ROWID;

CREATE TABLE memory_links (
    from_memory_id TEXT NOT NULL,
    to_memory_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    created_at_ns INTEGER NOT NULL CHECK (created_at_ns > 0),
    PRIMARY KEY (from_memory_id, to_memory_id, kind),
    FOREIGN KEY (from_memory_id) REFERENCES memories(id) ON DELETE CASCADE,
    FOREIGN KEY (to_memory_id) REFERENCES memories(id) ON DELETE CASCADE,
    CHECK (kind IN ('references', 'relates_to', 'derived_from')),
    CHECK (from_memory_id <> to_memory_id)
) WITHOUT ROWID;

CREATE TABLE meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    CHECK (length(key) > 0)
) WITHOUT ROWID;

CREATE INDEX idx_memories_kind_created_at_ns
    ON memories(kind, created_at_ns DESC);

CREATE INDEX idx_memories_updated_at_ns
    ON memories(updated_at_ns DESC);

CREATE INDEX idx_memories_archived_at_ns
    ON memories(archived_at_ns);

CREATE INDEX idx_memory_chunks_memory_id_ord
    ON memory_chunks(memory_id, ord);

CREATE INDEX idx_memory_links_from_memory_id
    ON memory_links(from_memory_id);

CREATE INDEX idx_memory_links_to_memory_id
    ON memory_links(to_memory_id);

INSERT INTO meta(key, value) VALUES
    ('schema_version', '1');
    