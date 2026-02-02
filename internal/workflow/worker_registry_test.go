package workflow

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"
)

type testWorker struct {
	name string
}

func (t *testWorker) Name() string                                 { return t.name }
func (t *testWorker) Register(ctx context.Context) error           { return nil }
func (t *testWorker) Start(ctx context.Context, addr string) error { return nil }

func TestWorkerRegistryRegisterAndGet(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := NewWorkerRegistry(logger)

	worker := &testWorker{name: "restate"}
	if err := registry.Register(worker); err != nil {
		t.Fatalf("register worker failed: %v", err)
	}

	got, err := registry.Get("restate")
	if err != nil {
		t.Fatalf("get worker failed: %v", err)
	}
	if got.Name() != "restate" {
		t.Fatalf("expected worker restate, got %s", got.Name())
	}
}

func TestWorkerRegistryDuplicate(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := NewWorkerRegistry(logger)

	worker := &testWorker{name: "restate"}
	if err := registry.Register(worker); err != nil {
		t.Fatalf("register worker failed: %v", err)
	}
	if err := registry.Register(worker); err == nil {
		t.Fatal("expected duplicate registration error")
	}
}
