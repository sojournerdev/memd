package memory

import (
	"context"
	"errors"
	"testing"
)

func TestService_CreateReturnsRepositoryMemory(t *testing.T) {
	t.Parallel()

	want := Memory{ID: "mem_123"}
	repo := &stubRepository{
		createResult: want,
	}
	svc := NewService(repo)

	got, err := svc.Create(context.Background(), CreateInput{ProjectKey: "project-a"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got != want {
		t.Fatalf("Create() = %#v, want %#v", got, want)
	}
}

func TestService_SearchReturnsRepositoryMemories(t *testing.T) {
	t.Parallel()

	want := []Memory{{ID: "mem_123"}}
	repo := &stubRepository{
		searchResult: want,
	}
	svc := NewService(repo)

	got, err := svc.Search(context.Background(), SearchInput{ProjectKey: "project-a", Query: "bootstrap"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("Search() = %#v, want %#v", got, want)
	}
}

func TestService_CreatePassesInputToRepository(t *testing.T) {
	t.Parallel()

	want := CreateInput{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
	}

	repo := &stubRepository{}
	svc := NewService(repo)

	if _, err := svc.Create(context.Background(), want); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repo.createInput != want {
		t.Fatalf("Create() input = %#v, want %#v", repo.createInput, want)
	}
}

func TestService_SearchPassesInputToRepository(t *testing.T) {
	t.Parallel()

	want := SearchInput{
		ProjectKey: "project-a",
		Query:      "bootstrap",
		Limit:      10,
	}

	repo := &stubRepository{}
	svc := NewService(repo)

	if _, err := svc.Search(context.Background(), want); err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if repo.searchInput != want {
		t.Fatalf("Search() input = %#v, want %#v", repo.searchInput, want)
	}
}

func TestService_PropagatesRepositoryErrors(t *testing.T) {
	t.Parallel()

	createErr := errors.New("create failed")
	searchErr := errors.New("search failed")
	repo := &stubRepository{
		createErr: createErr,
		searchErr: searchErr,
	}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{})
	if !errors.Is(err, createErr) {
		t.Fatalf("Create() error = %v, want %v", err, createErr)
	}

	_, err = svc.Search(context.Background(), SearchInput{})
	if !errors.Is(err, searchErr) {
		t.Fatalf("Search() error = %v, want %v", err, searchErr)
	}
}

type stubRepository struct {
	createInput CreateInput
	searchInput SearchInput

	createResult Memory
	searchResult []Memory
	createErr    error
	searchErr    error
}

func (r *stubRepository) Create(_ context.Context, input CreateInput) (Memory, error) {
	r.createInput = input
	return r.createResult, r.createErr
}

func (r *stubRepository) Search(_ context.Context, input SearchInput) ([]Memory, error) {
	r.searchInput = input
	return r.searchResult, r.searchErr
}
