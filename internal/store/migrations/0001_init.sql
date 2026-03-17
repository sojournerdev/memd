CREATE TABLE repo_files (
    repo_root TEXT NOT NULL,
    file_path TEXT NOT NULL,
    content_sha256 TEXT NOT NULL,
    byte_size INTEGER NOT NULL CHECK (byte_size >= 0),
    chunk_count INTEGER NOT NULL CHECK (chunk_count >= 0),
    PRIMARY KEY (repo_root, file_path),
    CHECK (length(repo_root) > 0),
    CHECK (length(file_path) > 0),
    CHECK (length(content_sha256) > 0)
) WITHOUT ROWID;

CREATE TABLE repo_chunks (
    chunk_id TEXT PRIMARY KEY,
    repo_root TEXT NOT NULL,
    file_path TEXT NOT NULL,
    chunk_index INTEGER NOT NULL CHECK (chunk_index >= 0),
    content TEXT NOT NULL,
    content_sha256 TEXT NOT NULL,
    FOREIGN KEY (repo_root, file_path) REFERENCES repo_files(repo_root, file_path) ON DELETE CASCADE,
    UNIQUE (repo_root, file_path, chunk_index),
    CHECK (length(chunk_id) > 0),
    CHECK (length(content_sha256) > 0)
) WITHOUT ROWID;

CREATE INDEX idx_repo_chunks_repo_file_chunk
    ON repo_chunks(repo_root, file_path, chunk_index);

CREATE INDEX idx_repo_chunks_repo_root
    ON repo_chunks(repo_root);

CREATE VIRTUAL TABLE repo_chunks_fts USING fts5(
    chunk_id UNINDEXED,
    content,
    tokenize = 'unicode61'
);
