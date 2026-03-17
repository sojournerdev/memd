package ingest

const (
	defaultChunkSizeRunes    = 1200
	defaultChunkOverlapRunes = 120
)

// Chunk splits content into deterministic, overlapping text chunks.
func Chunk(content string, chunkSize int) []string {
	if chunkSize <= 0 {
		chunkSize = defaultChunkSizeRunes
	}

	runes := []rune(content)
	if len(runes) == 0 {
		return nil
	}

	chunks := make([]string, 0, (len(runes)+chunkSize-1)/chunkSize)
	start := 0
	for start < len(runes) {
		end := chunkEnd(runes, start, chunkSize)
		chunks = append(chunks, string(runes[start:end]))
		if end == len(runes) {
			break
		}
		nextStart := end - defaultChunkOverlapRunes
		if nextStart <= start {
			nextStart = end
		}
		start = nextStart
	}

	return chunks
}

// chunkEnd chooses a deterministic boundary, preferring the last newline near
// the target size to avoid splitting blocks mid-line.
func chunkEnd(runes []rune, start, chunkSize int) int {
	target := start + chunkSize
	if target >= len(runes) {
		return len(runes)
	}

	minEnd := start + chunkSize/2
	if minEnd <= start {
		minEnd = start + 1
	}
	if minEnd > target {
		minEnd = target
	}

	for i := target - 1; i >= minEnd; i-- {
		if runes[i] == '\n' {
			return i + 1
		}
	}
	return target
}
