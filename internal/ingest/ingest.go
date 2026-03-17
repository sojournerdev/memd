package ingest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sojournerdev/memd/internal/store"
)

// Result reports how many files and chunks were indexed.
type Result struct {
	Files  int
	Chunks int
}

// Run ingests one repository into the local store.
func Run(ctx context.Context, db *sql.DB, repoPath string) (Result, error) {
	if db == nil {
		return Result{}, errors.New("ingest: nil db")
	}
	if repoPath == "" {
		return Result{}, errors.New("ingest: empty repo path")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	absRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return Result{}, fmt.Errorf("ingest: resolve repo path: %w", err)
	}
	absRepoPath = filepath.Clean(absRepoPath)

	fi, err := os.Stat(absRepoPath)
	if err != nil {
		return Result{}, fmt.Errorf("ingest: stat repo path: %w", err)
	}
	if !fi.IsDir() {
		return Result{}, fmt.Errorf("ingest: repo path is not a directory: %s", absRepoPath)
	}

	paths, err := Walk(ctx, absRepoPath)
	if err != nil {
		return Result{}, err
	}

	files := make([]store.RepoFileRecord, 0, len(paths))
	chunks := make([]store.RepoChunkRecord, 0, len(paths)*2)

	for _, relPath := range paths {
		if err := ctx.Err(); err != nil {
			return Result{}, fmt.Errorf("ingest: canceled: %w", err)
		}

		fullPath := filepath.Join(absRepoPath, filepath.FromSlash(relPath))
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return Result{}, fmt.Errorf("ingest: read file %q: %w", relPath, err)
		}
		if IsBinary(data) {
			continue
		}

		content := string(data)
		fileChunks := Chunk(content, defaultChunkSizeRunes)

		fileRec := store.RepoFileRecord{
			RepoRoot:      absRepoPath,
			FilePath:      relPath,
			ContentSHA256: sha256Hex(data),
			ByteSize:      len(data),
			ChunkCount:    len(fileChunks),
		}
		files = append(files, fileRec)

		for idx, chunk := range fileChunks {
			chunks = append(chunks, store.RepoChunkRecord{
				ChunkID:       stableChunkID(absRepoPath, relPath, idx, chunk),
				RepoRoot:      absRepoPath,
				FilePath:      relPath,
				ChunkIndex:    idx,
				Content:       chunk,
				ContentSHA256: sha256Hex([]byte(chunk)),
			})
		}
	}

	if err := store.ReplaceRepoIndex(ctx, db, absRepoPath, files, chunks); err != nil {
		return Result{}, err
	}

	return Result{Files: len(files), Chunks: len(chunks)}, nil
}

func stableChunkID(repoRoot, filePath string, chunkIndex int, chunk string) string {
	payload := repoRoot + "\n" + filePath + "\n" + strconv.Itoa(chunkIndex) + "\n" + chunk
	return sha256Hex([]byte(payload))
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
