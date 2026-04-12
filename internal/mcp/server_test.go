package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sojournerdev/memd/internal/memory"
)

func TestNew_ReturnsErrorForNilMemoryService(t *testing.T) {
	t.Parallel()

	_, err := New(nil, Options{})
	if err == nil {
		t.Fatal("New(nil) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "nil memory service") {
		t.Fatalf("New(nil) error = %q, want substring %q", err.Error(), "nil memory service")
	}
}

func TestCreateMemoryRequest_ToCreateInput(t *testing.T) {
	t.Parallel()

	got, err := (createMemoryRequest{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
		Tags:       []string{"alpha", "beta"},
		Metadata:   map[string]any{"source": "chat"},
	}).toCreateInput()
	if err != nil {
		t.Fatalf("toCreateInput() error = %v", err)
	}

	want := memory.CreateInput{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
		Tags:       []string{"alpha", "beta", "codex", "mcp"},
		Metadata:   json.RawMessage(`{"capture_client":"codex","capture_source":"mcp","capture_tool":"create_memory","source":"chat"}`),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("toCreateInput() = %#v, want %#v", got, want)
	}
}

func TestCreateMemoryRequest_ToCreateInputClonesTags(t *testing.T) {
	t.Parallel()

	input := createMemoryRequest{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
		Tags:       []string{"alpha"},
	}

	got, err := input.toCreateInput()
	if err != nil {
		t.Fatalf("toCreateInput() error = %v", err)
	}

	input.Tags[0] = "mutated"
	if !reflect.DeepEqual(got.Tags, []string{"alpha", "codex", "mcp"}) {
		t.Fatalf("Tags after caller mutation = %#v, want %#v", got.Tags, []string{"alpha", "codex", "mcp"})
	}
}

func TestCreateMemoryRequest_ToCreateInputRejectsInvalidMetadata(t *testing.T) {
	t.Parallel()

	input := createMemoryRequest{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
		Metadata: map[string]any{
			"bad": math.NaN(),
		},
	}

	_, err := input.toCreateInput()
	if err == nil {
		t.Fatal("toCreateInput() error = nil, want error")
	}
}

func TestCreateMemoryRequest_ToCreateInputUsesSystemMetadata(t *testing.T) {
	t.Parallel()

	got, err := (createMemoryRequest{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
		Tags:       []string{"mcp", "user"},
		Metadata: map[string]any{
			"capture_client": "custom-client",
			"source":         "chat",
		},
	}).toCreateInput()
	if err != nil {
		t.Fatalf("toCreateInput() error = %v", err)
	}

	if !reflect.DeepEqual(got.Tags, []string{"codex", "mcp", "user"}) {
		t.Fatalf("Tags = %#v, want %#v", got.Tags, []string{"codex", "mcp", "user"})
	}
	if string(got.Metadata) != `{"capture_client":"codex","capture_source":"mcp","capture_tool":"create_memory","source":"chat"}` {
		t.Fatalf("Metadata = %q, want system-owned metadata to override reserved keys", got.Metadata)
	}
}

func TestToMemoryResponse_MapsMemoryFields(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.April, 18, 10, 11, 12, 13, time.UTC)
	updatedAt := createdAt.Add(time.Second)

	got := toMemoryResponse(memory.Memory{
		ID:         "mem_123",
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
		Tags:       []string{"alpha", "beta"},
		Metadata:   json.RawMessage(`{"source":"chat"}`),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	})

	if got.ID != "mem_123" || got.ProjectKey != "project-a" {
		t.Fatalf("identity fields = %#v, want id/project_key mapped", got)
	}
	if got.CreatedAt != "2026-04-18T10:11:12.000000013Z" {
		t.Fatalf("CreatedAt = %q, want %q", got.CreatedAt, "2026-04-18T10:11:12.000000013Z")
	}
	if got.UpdatedAt != "2026-04-18T10:11:13.000000013Z" {
		t.Fatalf("UpdatedAt = %q, want %q", got.UpdatedAt, "2026-04-18T10:11:13.000000013Z")
	}
	if got.Metadata["source"] != "chat" {
		t.Fatalf("Metadata[source] = %#v, want %q", got.Metadata["source"], "chat")
	}
	if !reflect.DeepEqual(got.Tags, []string{"alpha", "beta"}) {
		t.Fatalf("Tags = %#v, want %#v", got.Tags, []string{"alpha", "beta"})
	}
}

func TestToMemoryResponse_ClonesTagsAndToleratesInvalidMetadata(t *testing.T) {
	t.Parallel()

	source := memory.Memory{
		Tags:     []string{"alpha"},
		Metadata: json.RawMessage(`{invalid`),
	}

	got := toMemoryResponse(source)
	source.Tags[0] = "mutated"

	if !reflect.DeepEqual(got.Tags, []string{"alpha"}) {
		t.Fatalf("Tags after source mutation = %#v, want %#v", got.Tags, []string{"alpha"})
	}
	if len(got.Metadata) != 0 {
		t.Fatalf("Metadata = %#v, want empty object map on invalid input", got.Metadata)
	}
}

func TestEncodeMetadata(t *testing.T) {
	t.Parallel()

	got, err := encodeMetadata(map[string]any{"source": "chat"})
	if err != nil {
		t.Fatalf("encodeMetadata() error = %v", err)
	}
	if string(got) != `{"source":"chat"}` {
		t.Fatalf("encodeMetadata() = %q, want %q", got, `{"source":"chat"}`)
	}
}

func TestDecodeMetadata(t *testing.T) {
	t.Parallel()

	got := decodeMetadata(json.RawMessage(`{"source":"chat"}`))
	if got["source"] != "chat" {
		t.Fatalf("decodeMetadata()[source] = %#v, want %q", got["source"], "chat")
	}
}

func TestDecodeMetadata_ReturnsEmptyMapForInvalidJSON(t *testing.T) {
	t.Parallel()

	got := decodeMetadata(json.RawMessage(`{invalid`))
	if len(got) != 0 {
		t.Fatalf("decodeMetadata(invalid) = %#v, want empty map", got)
	}
}

func TestServer_RunStdioNilSafe(t *testing.T) {
	t.Parallel()

	var srv *Server
	err := srv.RunStdio(context.Background())
	if err == nil {
		t.Fatal("RunStdio(nil) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "nil server") {
		t.Fatalf("RunStdio(nil) error = %q, want substring %q", err.Error(), "nil server")
	}
}

func TestNew_AcceptsValidMemoryService(t *testing.T) {
	t.Parallel()

	svc, err := memory.NewService(stubRepository{})
	if err != nil {
		t.Fatalf("memory.NewService() error = %v", err)
	}

	server, err := New(svc, Options{Version: "test-version"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if server == nil || server.server == nil {
		t.Fatal("server = nil, want initialized MCP server")
	}
}

type stubRepository struct{}

func (stubRepository) Create(context.Context, memory.CreateInput) (memory.Memory, error) {
	return memory.Memory{}, errors.New("not implemented")
}

func (stubRepository) Get(context.Context, string) (memory.Memory, error) {
	return memory.Memory{}, errors.New("not implemented")
}
