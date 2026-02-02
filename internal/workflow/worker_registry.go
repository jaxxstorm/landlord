package workflow

import (
	"fmt"
	"sort"
	"sync"

	"go.uber.org/zap"
)

// WorkerRegistry stores registered workflow worker engines.
type WorkerRegistry struct {
	mu      sync.RWMutex
	workers map[string]WorkerEngine
	logger  *zap.Logger
}

// NewWorkerRegistry creates a new worker registry.
func NewWorkerRegistry(logger *zap.Logger) *WorkerRegistry {
	return &WorkerRegistry{
		workers: make(map[string]WorkerEngine),
		logger:  logger.With(zap.String("component", "workflow-worker-registry")),
	}
}

// Register registers a workflow worker engine.
func (r *WorkerRegistry) Register(worker WorkerEngine) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := worker.Name()
	if name == "" {
		return fmt.Errorf("worker name cannot be empty")
	}

	if _, exists := r.workers[name]; exists {
		r.logger.Error("worker already registered",
			zap.String("worker", name),
		)
		return fmt.Errorf("%w: %s", ErrProviderConflict, name)
	}

	r.workers[name] = worker
	r.logger.Info("registered workflow worker",
		zap.String("worker", name),
	)

	return nil
}

// Get retrieves a worker engine by name.
func (r *WorkerRegistry) Get(workerType string) (WorkerEngine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	worker, exists := r.workers[workerType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, workerType)
	}

	return worker, nil
}

// List returns all registered worker engine names.
func (r *WorkerRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.workers))
	for name := range r.workers {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

// Has checks if a worker engine is registered.
func (r *WorkerRegistry) Has(workerType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.workers[workerType]
	return exists
}
