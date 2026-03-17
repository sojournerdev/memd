package ingest

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
)

func Walk(ctx context.Context, root string) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("ingest: walk: relative path for %q: %w", path, err)
		}
		if relPath == "." {
			return nil
		}

		if d.IsDir() {
			if ShouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !IsSupportedFile(relPath) {
			return nil
		}

		files = append(files, filepath.ToSlash(relPath))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ingest: walk repo: %w", err)
	}

	return files, nil
}
