package ingest

import (
	"bytes"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	binarySampleSize      = 8000
	binaryControlRatioMax = 0.01
)

var allowedExtensions = stringSet(
	".go",
	".js",
	".ts",
	".jsx",
	".tsx",
	".py",
	".java",
	".cs",
	".rs",
	".rb",
	".php",
	".c",
	".h",
	".cpp",
	".hpp",
	".swift",
	".kt",
	".sh",
	".bash",
	".zsh",
	".fish",
	".html",
	".css",
	".scss",
	".json",
	".yaml",
	".yml",
	".toml",
	".xml",
	".md",
	".txt",
	".sql",
)

var skippedDirNames = map[string]struct{}{
	".git":          {},
	".gradle":       {},
	".next":         {},
	".nuxt":         {},
	".pytest_cache": {},
	".svelte-kit":   {},
	".tox":          {},
	".venv":         {},
	"__pycache__":   {},
	"bin":           {},
	"build":         {},
	"CMakeFiles":    {},
	"coverage":      {},
	"dist":          {},
	"node_modules":  {},
	"obj":           {},
	"out":           {},
	"target":        {},
	"vendor":        {},
	"venv":          {},
}

// ShouldSkipDir reports whether a directory should be skipped during ingest.
func ShouldSkipDir(name string) bool {
	_, ok := skippedDirNames[name]
	return ok
}

// IsSupportedFile reports whether a file path has an allowed extension.
func IsSupportedFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := allowedExtensions[ext]
	return ok
}

// IsBinary reports whether data looks like binary content.
func IsBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return true
	}
	if !utf8.Valid(data) {
		return true
	}

	sample := data
	if len(sample) > binarySampleSize {
		sample = sample[:binarySampleSize]
	}

	control := 0
	for _, b := range sample {
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
			control++
		}
		if b == 0x7f {
			control++
		}
	}

	return float64(control)/float64(len(sample)) > binaryControlRatioMax
}

func stringSet(values ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, v := range values {
		out[v] = struct{}{}
	}
	return out
}
