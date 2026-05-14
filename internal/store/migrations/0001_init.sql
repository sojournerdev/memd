CREATE TABLE memories (
    memory_id TEXT PRIMARY KEY,
    project_key TEXT NOT NULL,
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at_ns INTEGER NOT NULL,
    updated_at_ns INTEGER NOT NULL,

    CHECK (length(trim(memory_id)) > 0),
    CHECK (length(trim(project_key)) > 0),
    CHECK (length(trim(title)) > 0),
    CHECK (length(trim(summary)) > 0),
    CHECK (length(trim(content)) > 0),
    CHECK (created_at_ns > 0),
    CHECK (updated_at_ns > 0),
    CHECK (updated_at_ns >= created_at_ns)
) STRICT, WITHOUT ROWID;

CREATE INDEX idx_memories_project_updated
    ON memories(project_key, updated_at_ns DESC, memory_id);

CREATE VIRTUAL TABLE memory_search USING fts5(
    memory_id UNINDEXED,
    project_key UNINDEXED,
    title,
    summary,
    content,
    tokenize = 'unicode61'
);
