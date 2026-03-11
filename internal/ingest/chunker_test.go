package ingest

import (
	"strings"
	"testing"
)

func TestChunk_DeterministicAndNewlineAware(t *testing.T) {
	t.Parallel()

	content := strings.Repeat("line\n", 700)
	first := Chunk(content, 120)
	second := Chunk(content, 120)

	if len(first) == 0 {
		t.Fatalf("Chunk() produced no chunks")
	}
	if strings.Join(first, "") == "" {
		t.Fatalf("Chunk() produced empty output")
	}
	if len(first) != len(second) {
		t.Fatalf("chunk count mismatch: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("chunk %d mismatch between deterministic runs", i)
		}
	}

	if len(first) < 2 {
		t.Fatalf("Chunk() produced %d chunks, want at least 2", len(first))
	}
	for i := 0; i < len(first)-1; i++ {
		if !strings.HasSuffix(first[i], "\n") {
			t.Fatalf("chunk %d does not end at newline boundary: %q", i, first[i])
		}
	}
}

func TestChunk_OverlapsAdjacentChunks(t *testing.T) {
	t.Parallel()

	content := strings.Repeat("abcdefghij", 300)
	chunks := Chunk(content, 120)
	if len(chunks) < 2 {
		t.Fatalf("Chunk() produced %d chunks, want at least 2", len(chunks))
	}

	prevRunes := []rune(chunks[0])
	nextRunes := []rune(chunks[1])
	wantOverlap := string(prevRunes[len(prevRunes)-defaultChunkOverlapRunes:])
	gotPrefix := string(nextRunes[:defaultChunkOverlapRunes])

	if gotPrefix != wantOverlap {
		t.Fatalf("overlap mismatch: got %q, want %q", gotPrefix, wantOverlap)
	}
}

func TestChunk_FallsBackToHardBoundaryWithoutNewline(t *testing.T) {
	t.Parallel()

	content := strings.Repeat("a", 250)
	chunks := Chunk(content, 120)
	if len(chunks) < 2 {
		t.Fatalf("Chunk() produced %d chunks, want at least 2", len(chunks))
	}

	firstLen := len([]rune(chunks[0]))
	if firstLen != 120 {
		t.Fatalf("first chunk length = %d, want %d", firstLen, 120)
	}
}
