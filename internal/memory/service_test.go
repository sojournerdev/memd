package memory

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type testContextKey struct{}

func TestNewService_ReturnsErrorForNilRepository(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil)
	if !errors.Is(err, errNilRepository) {
		t.Fatalf("NewService(nil) error = %v, want %v", err, errNilRepository)
	}
}

func TestService_MethodsReturnErrorForNilReceiver(t *testing.T) {
	t.Parallel()

	var svc *Service

	_, err := svc.Create(context.Background(), CreateInput{})
	if !errors.Is(err, errNilRepository) {
		t.Fatalf("nil Create() error = %v, want %v", err, errNilRepository)
	}

	_, err = svc.Get(context.Background(), "mem_123")
	if !errors.Is(err, errNilRepository) {
		t.Fatalf("nil Get() error = %v, want %v", err, errNilRepository)
	}

	svc = &Service{}

	_, err = svc.Create(context.Background(), CreateInput{})
	if !errors.Is(err, errNilRepository) {
		t.Fatalf("empty Create() error = %v, want %v", err, errNilRepository)
	}

	_, err = svc.Get(context.Background(), "mem_123")
	if !errors.Is(err, errNilRepository) {
		t.Fatalf("empty Get() error = %v, want %v", err, errNilRepository)
	}
}

func TestService_DelegatesToRepository(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{
		createResult: Memory{ID: "mem_123"},
		getResult:    Memory{ID: "mem_456"},
	}
	svc, err := NewService(repo)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	created, err := svc.Create(context.Background(), CreateInput{ProjectKey: "project-a"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != "mem_123" {
		t.Fatalf("Create() ID = %q, want %q", created.ID, "mem_123")
	}

	got, err := svc.Get(context.Background(), "mem_456")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != "mem_456" {
		t.Fatalf("Get() ID = %q, want %q", got.ID, "mem_456")
	}
	if repo.createCalls != 1 || repo.getCalls != 1 {
		t.Fatalf("calls = create:%d get:%d, want 1 and 1", repo.createCalls, repo.getCalls)
	}
}

func TestService_PassesContextAndInputToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := testContextKey{}
	ctx := context.WithValue(context.Background(), ctxKey, "trace")
	input := CreateInput{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
	}

	repo := &stubRepository{}
	svc, err := NewService(repo)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if _, err := svc.Create(ctx, input); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repo.createCtx != ctx {
		t.Fatal("Create() context was not forwarded to repository")
	}
	if !reflect.DeepEqual(repo.createInput, input) {
		t.Fatalf("Create() input = %#v, want %#v", repo.createInput, input)
	}

	if _, err := svc.Get(ctx, "mem_123"); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if repo.getCtx != ctx {
		t.Fatal("Get() context was not forwarded to repository")
	}
	if repo.getID != "mem_123" {
		t.Fatalf("Get() id = %q, want %q", repo.getID, "mem_123")
	}
}

func TestService_PropagatesRepositoryErrors(t *testing.T) {
	t.Parallel()

	createErr := errors.New("create failed")
	getErr := errors.New("get failed")
	repo := &stubRepository{
		createErr: createErr,
		getErr:    getErr,
	}
	svc, err := NewService(repo)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = svc.Create(context.Background(), CreateInput{})
	if !errors.Is(err, createErr) {
		t.Fatalf("Create() error = %v, want %v", err, createErr)
	}

	_, err = svc.Get(context.Background(), "mem_123")
	if !errors.Is(err, getErr) {
		t.Fatalf("Get() error = %v, want %v", err, getErr)
	}
}

type stubRepository struct {
	createCalls int
	getCalls    int

	createCtx   context.Context
	getCtx      context.Context
	createInput CreateInput
	getID       string

	createResult Memory
	getResult    Memory
	createErr    error
	getErr       error
}

func (r *stubRepository) Create(ctx context.Context, input CreateInput) (Memory, error) {
	r.createCalls++
	r.createCtx = ctx
	r.createInput = input
	return r.createResult, r.createErr
}

func (r *stubRepository) Get(ctx context.Context, id string) (Memory, error) {
	r.getCalls++
	r.getCtx = ctx
	r.getID = id
	return r.getResult, r.getErr
}
