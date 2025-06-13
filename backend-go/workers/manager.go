package workers

import (
	"context"
	"fmt"
	"ftoolbox/models"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WorkerManager struct {
	db          *gorm.DB
	workers     map[string]Worker
	cancelFuncs map[string]context.CancelFunc
	running     map[string]bool
	mu          sync.RWMutex
	wg          sync.WaitGroup
	enabled     bool
}

func NewWorkerManager(db *gorm.DB, enabled bool) *WorkerManager {
	return &WorkerManager{
		db:          db,
		workers:     make(map[string]Worker),
		cancelFuncs: make(map[string]context.CancelFunc),
		running:     make(map[string]bool),
		enabled:     enabled,
	}
}

// Register adds a worker to the manager
func (m *WorkerManager) Register(worker Worker) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := worker.Name()
	if _, exists := m.workers[name]; exists {
		return fmt.Errorf("worker %s already registered", name)
	}

	m.workers[name] = worker

	// Ensure worker exists in database
	var dbWorker models.Worker
	if err := m.db.Where("name = ?", name).First(&dbWorker).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new worker record
			dbWorker = models.Worker{
				Name:      name,
				Status:    "idle",
				IsEnabled: true,
			}
			if err := m.db.Create(&dbWorker).Error; err != nil {
				return fmt.Errorf("failed to create worker record: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check worker existence: %w", err)
		}
	}

	zap.L().Info("Worker registered", zap.String("worker", name))
	return nil
}

// Start begins running a worker
func (m *WorkerManager) Start(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		zap.L().Info("Workers disabled, skipping start", zap.String("worker", name))
		return nil
	}

	worker, exists := m.workers[name]
	if !exists {
		return fmt.Errorf("worker %s not found", name)
	}

	// Check if already running
	if _, running := m.cancelFuncs[name]; running {
		zap.L().Warn("Worker already running", zap.String("worker", name))
		return fmt.Errorf("worker %s already running", name)
	}

	// Check if worker is enabled in database
	var dbWorker models.Worker
	if err := m.db.Where("name = ?", name).First(&dbWorker).Error; err != nil {
		return fmt.Errorf("failed to fetch worker status: %w", err)
	}

	if !dbWorker.IsEnabled {
		zap.L().Info("Worker is disabled", zap.String("worker", name))
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFuncs[name] = cancel

	m.wg.Add(1)
	go m.runWorker(ctx, worker)

	zap.L().Info("Worker started", zap.String("worker", name))
	return nil
}

// Stop halts a running worker
func (m *WorkerManager) Stop(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, exists := m.cancelFuncs[name]
	if !exists {
		return fmt.Errorf("worker %s not running", name)
	}

	cancel()
	delete(m.cancelFuncs, name)

	zap.L().Info("Worker stopped", zap.String("worker", name))
	return nil
}

// StopAll stops all running workers
func (m *WorkerManager) StopAll() {
	m.mu.Lock()
	for name, cancel := range m.cancelFuncs {
		cancel()
		delete(m.cancelFuncs, name)
		zap.L().Info("Worker stopped", zap.String("worker", name))
	}
	m.mu.Unlock()

	// Wait for all workers to finish
	m.wg.Wait()
	zap.L().Info("All workers stopped")
}

// GetStatus returns the status of all workers
func (m *WorkerManager) GetStatus() ([]models.Worker, error) {
	var workers []models.Worker
	if err := m.db.Find(&workers).Error; err != nil {
		return nil, err
	}
	return workers, nil
}

func (m *WorkerManager) runWorker(ctx context.Context, worker Worker) {
	defer m.wg.Done()

	ticker := time.NewTicker(worker.Interval())
	defer ticker.Stop()

	// Run immediately
	m.executeWorker(ctx, worker)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.executeWorker(ctx, worker)
		}
	}
}

func (m *WorkerManager) executeWorker(ctx context.Context, worker Worker) {
	name := worker.Name()

	// Check if already running
	m.mu.RLock()
	if m.running[name] {
		m.mu.RUnlock()
		zap.L().Debug("Worker already running, skipping", zap.String("worker", name))
		return
	}
	m.mu.RUnlock()

	// Mark as running
	m.mu.Lock()
	m.running[name] = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.running[name] = false
		m.mu.Unlock()
	}()

	// Update status to running
	now := time.Now()
	if err := m.db.Model(&models.Worker{}).
		Where("name = ?", name).
		Updates(map[string]interface{}{
			"status":      "running",
			"last_run_at": now,
			"updated_at":  now,
		}).Error; err != nil {
		zap.L().Error("Failed to update worker status", zap.String("worker", name), zap.Error(err))
		return
	}

	// Run the worker
	startTime := time.Now()
	err := worker.Run(ctx)
	duration := time.Since(startTime)

	// Update worker status based on result
	status := "idle"
	updates := map[string]interface{}{
		"status":      status,
		"run_count":   gorm.Expr("run_count + 1"),
		"updated_at":  time.Now(),
		"next_run_at": time.Now().Add(worker.Interval()),
	}

	if err != nil {
		status = "failed"
		updates["status"] = status
		updates["failure_count"] = gorm.Expr("failure_count + 1")
		updates["last_error"] = err.Error()

		zap.L().Error("Worker failed",
			zap.String("worker", name),
			zap.Error(err),
			zap.Duration("duration", duration))
	} else {
		updates["success_count"] = gorm.Expr("success_count + 1")
		updates["last_error"] = nil

		zap.L().Info("Worker completed",
			zap.String("worker", name),
			zap.Duration("duration", duration))
	}

	if err := m.db.Model(&models.Worker{}).
		Where("name = ?", name).
		Updates(updates).Error; err != nil {
		zap.L().Error("Failed to update worker status", zap.String("worker", name), zap.Error(err))
	}
}
