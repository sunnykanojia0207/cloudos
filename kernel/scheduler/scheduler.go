// Package scheduler provides a lightweight, in-memory task scheduler for
// running periodic and one-shot background jobs within the kernel.
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/packages/logging"
)

// TaskFunc is the function signature for a scheduled task.
type TaskFunc func(ctx context.Context) error

// Task represents a scheduled job.
type Task struct {
	Name     string
	Interval time.Duration
	RunOnce  bool
	Func     TaskFunc
}

// Scheduler manages the execution of periodic and one-shot tasks.
type Scheduler struct {
	mu      sync.Mutex
	tasks   map[string]*Task
	cancels map[string]context.CancelFunc
	log     *logging.Logger
	running bool
}

// New creates a new Scheduler.
func New(log *logging.Logger) *Scheduler {
	return &Scheduler{
		tasks:   make(map[string]*Task),
		cancels: make(map[string]context.CancelFunc),
		log:     log,
	}
}

// Start enables task scheduling.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()
	s.log.Debug("scheduler started")
}

// Stop cancels all running tasks and prevents new ones from starting.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	for name, cancel := range s.cancels {
		cancel()
		delete(s.cancels, name)
	}
	s.log.Debug("scheduler stopped")
}

// Schedule adds a task to the scheduler. The task begins running immediately
// if the scheduler is running.
func (s *Scheduler) Schedule(task Task) error {
	if task.Func == nil {
		return fmt.Errorf("scheduler: task %q has nil function", task.Name)
	}
	if task.Interval <= 0 && !task.RunOnce {
		return fmt.Errorf("scheduler: task %q has no interval and is not one-shot", task.Name)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.Name]; exists {
		return fmt.Errorf("scheduler: task %q already scheduled", task.Name)
	}
	s.tasks[task.Name] = &task

	if s.running {
		s.startTask(&task)
	}

	return nil
}

// Unschedule removes and stops a previously scheduled task.
func (s *Scheduler) Unschedule(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, ok := s.cancels[name]; ok {
		cancel()
		delete(s.cancels, name)
	}
	delete(s.tasks, name)
}

// startTask launches a task in a background goroutine.
func (s *Scheduler) startTask(task *Task) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancels[task.Name] = cancel

	go func() {
		s.log.Debug("task scheduled", "name", task.Name, "interval", task.Interval)

		if task.RunOnce {
			s.executeTask(ctx, task)
			return
		}

		ticker := time.NewTicker(task.Interval)
		defer ticker.Stop()

		// Run immediately on schedule, then on each tick.
		s.executeTask(ctx, task)
		for {
			select {
			case <-ticker.C:
				s.executeTask(ctx, task)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Scheduler) executeTask(ctx context.Context, task *Task) {
	if err := task.Func(ctx); err != nil {
		s.log.Error("task failed", "name", task.Name, "error", err)
	}
}

// Running returns whether the scheduler is accepting tasks.
func (s *Scheduler) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Tasks returns the names of all scheduled tasks.
func (s *Scheduler) Tasks() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]string, 0, len(s.tasks))
	for n := range s.tasks {
		names = append(names, n)
	}
	return names
}
