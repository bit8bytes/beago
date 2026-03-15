package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/bit8bytes/beago/inputs/roles"
	"github.com/bit8bytes/beago/llms"
)

var ctx = context.Background()

func msg(role roles.Role, content string) llms.Message {
	return llms.Message{Role: role, Content: content}
}

func TestAddAndList(t *testing.T) {
	s := New()
	msgs := []llms.Message{
		msg(roles.User, "hello"),
		msg(roles.Assistant, "world"),
	}
	if err := s.Add(ctx, msgs...); err != nil {
		t.Fatal(err)
	}
	got, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != msgs[0] || got[1] != msgs[1] {
		t.Fatalf("unexpected messages: %v", got)
	}
}

func TestListReturnsCopy(t *testing.T) {
	s := New()
	_ = s.Add(ctx, msg(roles.User, "original"))
	got, _ := s.List(ctx)
	got[0].Content = "mutated"

	got2, _ := s.List(ctx)
	if got2[0].Content != "original" {
		t.Fatal("List returned a reference to internal slice")
	}
}

func TestListEmptyStore(t *testing.T) {
	s := New()
	got, err := s.List(ctx)
	if err != nil || len(got) != 0 {
		t.Fatalf("expected empty list, got %v %v", got, err)
	}
}

func TestClear(t *testing.T) {
	s := New()
	_ = s.Add(ctx, msg(roles.User, "hello"))
	if err := s.Clear(ctx); err != nil {
		t.Fatal(err)
	}
	got, _ := s.List(ctx)
	if len(got) != 0 {
		t.Fatalf("expected empty after Clear, got %v", got)
	}
}

func TestClearThenAdd(t *testing.T) {
	s := New()
	_ = s.Add(ctx, msg(roles.User, "first"))
	_ = s.Clear(ctx)
	_ = s.Add(ctx, msg(roles.User, "second"))
	got, _ := s.List(ctx)
	if len(got) != 1 || got[0].Content != "second" {
		t.Fatalf("unexpected messages after Clear+Add: %v", got)
	}
}

func TestConcurrentAddList(t *testing.T) {
	s := New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); _ = s.Add(ctx, msg(roles.User, "x")) }()
		go func() { defer wg.Done(); _, _ = s.List(ctx) }()
	}
	wg.Wait()
}
