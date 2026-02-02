package compute

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"go.uber.org/zap"
)

func TestRegistryRegister(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	provider := &testProvider{name: "test"}

	// Test successful registration
	err := registry.Register(provider)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	// Test duplicate registration
	err = registry.Register(provider)
	if !errors.Is(err, ErrProviderConflict) {
		t.Errorf("expected ErrProviderConflict, got: %v", err)
	}
}

func TestRegistryGet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	provider := &testProvider{name: "test"}
	registry.Register(provider)

	// Test getting existing provider
	p, err := registry.Get("test")
	if err != nil {
		t.Fatalf("failed to get provider: %v", err)
	}
	if p.Name() != "test" {
		t.Errorf("expected provider name 'test', got: %s", p.Name())
	}

	// Test getting non-existent provider
	_, err = registry.Get("nonexistent")
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got: %v", err)
	}
}

func TestRegistryList(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	// Test empty registry
	names := registry.List()
	if len(names) != 0 {
		t.Errorf("expected empty list, got: %v", names)
	}

	// Register multiple providers
	registry.Register(&testProvider{name: "provider1"})
	registry.Register(&testProvider{name: "provider2"})
	registry.Register(&testProvider{name: "provider3"})

	names = registry.List()
	if len(names) != 3 {
		t.Errorf("expected 3 providers, got: %d", len(names))
	}

	// Verify all names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}
	if !nameMap["provider1"] || !nameMap["provider2"] || !nameMap["provider3"] {
		t.Errorf("missing expected providers in list: %v", names)
	}
}

func TestRegistryHas(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	provider := &testProvider{name: "test"}
	registry.Register(provider)

	// Test existing provider
	if !registry.Has("test") {
		t.Error("expected Has to return true for registered provider")
	}

	// Test non-existent provider
	if registry.Has("nonexistent") {
		t.Error("expected Has to return false for non-existent provider")
	}
}

func TestRegistryConcurrency(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	// Test concurrent registration
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			provider := &testProvider{name: fmt.Sprintf("provider%d", id)}
			registry.Register(provider)
		}(i)
	}
	wg.Wait()

	// Verify all providers registered
	names := registry.List()
	if len(names) != 10 {
		t.Errorf("expected 10 providers, got: %d", len(names))
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("provider%d", id)
			p, err := registry.Get(name)
			if err != nil {
				t.Errorf("failed to get provider %s: %v", name, err)
				return
			}
			if p.Name() != name {
				t.Errorf("expected provider name %s, got: %s", name, p.Name())
			}
		}(i)
	}
	wg.Wait()
}

func TestRegistryEmptyProviderName(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	provider := &testProvider{name: ""}
	err := registry.Register(provider)
	if err == nil {
		t.Error("expected error when registering provider with empty name")
	}
}
