package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sojournerdev/memd/internal/memory"
)

func TestCreateMemoryRequest_ToCreateInput(t *testing.T) {
	t.Parallel()

	got, err := (createMemoryRequest{
		ProjectKey: "  project-a  ",
		Title:      "  Title  ",
		Summary:    "  Summary  ",
		Content:    "  Content  ",
	}).toCreateInput()
	if err != nil {
		t.Fatalf("toCreateInput() error = %v", err)
	}

	want := memory.CreateInput{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
	}
	if got != want {
		t.Fatalf("toCreateInput() = %#v, want %#v", got, want)
	}
}

func TestCreateMemoryRequest_ToCreateInputRejectsBlankFields(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		req  createMemoryRequest
	}{
		{
			name: "blank project key",
			req: createMemoryRequest{
				ProjectKey: " ",
				Title:      "Title",
				Summary:    "Summary",
				Content:    "Content",
			},
		},
		{
			name: "blank title",
			req: createMemoryRequest{
				ProjectKey: "project-a",
				Title:      " ",
				Summary:    "Summary",
				Content:    "Content",
			},
		},
		{
			name: "blank summary",
			req: createMemoryRequest{
				ProjectKey: "project-a",
				Title:      "Title",
				Summary:    " ",
				Content:    "Content",
			},
		},
		{
			name: "blank content",
			req: createMemoryRequest{
				ProjectKey: "project-a",
				Title:      "Title",
				Summary:    "Summary",
				Content:    " ",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.req.toCreateInput()
			if err == nil {
				t.Fatal("toCreateInput() error = nil, want error")
			}
		})
	}
}

func TestSearchMemoriesRequest_ToSearchInput(t *testing.T) {
	t.Parallel()

	got, err := (searchMemoriesRequest{
		ProjectKey: "  project-a  ",
		Query:      "  bootstrap plan  ",
		Limit:      3,
	}).toSearchInput()
	if err != nil {
		t.Fatalf("toSearchInput() error = %v", err)
	}

	want := memory.SearchInput{
		ProjectKey: "project-a",
		Query:      "bootstrap plan",
		Limit:      3,
	}
	if got != want {
		t.Fatalf("toSearchInput() = %#v, want %#v", got, want)
	}
}

func TestSearchMemoriesRequest_ToSearchInputDefaultsAndClampsLimit(t *testing.T) {
	t.Parallel()

	got, err := (searchMemoriesRequest{
		ProjectKey: "project-a",
		Query:      "bootstrap",
	}).toSearchInput()
	if err != nil {
		t.Fatalf("toSearchInput() default limit error = %v", err)
	}
	if got.Limit != defaultSearchLimit {
		t.Fatalf("default Limit = %d, want %d", got.Limit, defaultSearchLimit)
	}

	got, err = (searchMemoriesRequest{
		ProjectKey: "project-a",
		Query:      "bootstrap",
		Limit:      maxSearchLimit + 1,
	}).toSearchInput()
	if err != nil {
		t.Fatalf("toSearchInput() clamp limit error = %v", err)
	}
	if got.Limit != maxSearchLimit {
		t.Fatalf("clamped Limit = %d, want %d", got.Limit, maxSearchLimit)
	}
}

func TestSearchMemoriesRequest_ToSearchInputRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		req  searchMemoriesRequest
	}{
		{
			name: "blank project key",
			req: searchMemoriesRequest{
				ProjectKey: " ",
				Query:      "bootstrap",
			},
		},
		{
			name: "blank query",
			req: searchMemoriesRequest{
				ProjectKey: "project-a",
				Query:      " ",
			},
		},
		{
			name: "negative limit",
			req: searchMemoriesRequest{
				ProjectKey: "project-a",
				Query:      "bootstrap",
				Limit:      -1,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.req.toSearchInput()
			if err == nil {
				t.Fatal("toSearchInput() error = nil, want error")
			}
		})
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
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	})

	want := memoryResponse{
		ID:         "mem_123",
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
		CreatedAt:  "2026-04-18T10:11:12.000000013Z",
		UpdatedAt:  "2026-04-18T10:11:13.000000013Z",
	}
	if got != want {
		t.Fatalf("toMemoryResponse() = %#v, want %#v", got, want)
	}
}

func TestToMemoryResponses_MapsMemoryList(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.April, 18, 10, 11, 12, 13, time.UTC)
	got := toMemoryResponses([]memory.Memory{
		{
			ID:         "mem_123",
			ProjectKey: "project-a",
			Title:      "Title",
			Summary:    "Summary",
			Content:    "Content",
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		},
	})

	want := []memoryResponse{
		{
			ID:         "mem_123",
			ProjectKey: "project-a",
			Title:      "Title",
			Summary:    "Summary",
			Content:    "Content",
			CreatedAt:  "2026-04-18T10:11:12.000000013Z",
			UpdatedAt:  "2026-04-18T10:11:12.000000013Z",
		},
	}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("toMemoryResponses() = %#v, want %#v", got, want)
	}
}

func TestNew_AcceptsValidMemoryService(t *testing.T) {
	t.Parallel()

	memoryService := memory.NewService(stubRepository{})
	srv := New(memoryService, Options{Version: "test-version"})
	if srv == nil || srv.srv == nil {
		t.Fatal("srv = nil, want initialized MCP server")
	}
}

type stubRepository struct{}

func (stubRepository) Create(context.Context, memory.CreateInput) (memory.Memory, error) {
	return memory.Memory{}, errors.New("not implemented")
}

func (stubRepository) Search(context.Context, memory.SearchInput) ([]memory.Memory, error) {
	return nil, errors.New("not implemented")
}
