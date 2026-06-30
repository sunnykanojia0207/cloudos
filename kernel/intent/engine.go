package intent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/plugin"
	"github.com/cloudos/cloudos/kernel/registry"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// IntentEngine orchestrates the full intent lifecycle: parse → plan → execute
type IntentEngine struct {
	parser  *Parser
	planner *Planner
	exec    *Executor
	bus     *events.Bus
	log     *logging.Logger

	mu      sync.RWMutex
	intents map[string]*Intent
	plans   map[string]*ExecutionPlan
	counter int
}

// NewIntentEngine creates a new IntentEngine with all dependencies
func NewIntentEngine(
	resRegistry *resource.Registry,
	ctrlManager *controller.Manager,
	healthMgr *health.Manager,
	pluginReg *plugin.Registry,
	capRegistry *registry.Manager,
	providerReg *registry.Manager,
	bus *events.Bus,
	log *logging.Logger,
) *IntentEngine {
	exec := NewExecutor(ExecutorDeps{
		ResourceRegistry:   resRegistry,
		ControllerManager:  ctrlManager,
		HealthManager:      healthMgr,
		PluginRegistry:     pluginReg,
		CapabilityRegistry: capRegistry,
		ProviderRegistry:   providerReg,
		EventBus:           bus,
		Logger:             log,
	})

	return &IntentEngine{
		parser:  NewParser(),
		planner: NewPlanner(),
		exec:    exec,
		bus:     bus,
		log:     log,
		intents: make(map[string]*Intent),
		plans:   make(map[string]*ExecutionPlan),
	}
}

// Submit processes a user input string through the full intent pipeline
func (ie *IntentEngine) Submit(ctx context.Context, input string) (*Intent, error) {
	id := ie.nextID()

	intent := &Intent{
		ID:        id,
		Raw:       input,
		Status:    IntentPending,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	ie.mu.Lock()
	ie.intents[id] = intent
	ie.mu.Unlock()

	ie.log.Info("intent submitted", "id", id, "input", input)
	ie.publishEvent("intent.submitted", map[string]interface{}{
		"intent_id": id,
		"raw":       input,
	})

	// Phase 1: Parse
	intent.Status = IntentParsing
	parsed, err := ie.parser.Parse(input)
	if err != nil {
		intent.Status = IntentFailed
		intent.Error = err.Error()
		ie.publishEvent("intent.failed", map[string]interface{}{
			"intent_id": id,
			"error":     err.Error(),
		})
		return intent, nil
	}
	parsed.ID = id
	parsed.Status = IntentValidating
	parsed.CreatedAt = intent.CreatedAt
	*intent = *parsed

	// Phase 2: Plan
	intent.Status = IntentPlanning
	plan, err := ie.planner.Plan(intent)
	if err != nil {
		intent.Status = IntentFailed
		intent.Error = err.Error()
		ie.publishEvent("intent.failed", map[string]interface{}{
			"intent_id": id,
			"error":     err.Error(),
		})
		return intent, nil
	}
	plan.Status = IntentPending

	intent.PlanID = plan.ID

	ie.mu.Lock()
	ie.plans[plan.ID] = plan
	ie.mu.Unlock()

	ie.publishEvent("intent.planned", map[string]interface{}{
		"intent_id": id,
		"plan_id":   plan.ID,
		"steps":     len(plan.Steps),
	})

	// Phase 3: Execute (async)
	intent.Status = IntentExecuting
	go ie.executePlan(ctx, intent, plan)

	return intent, nil
}

func (ie *IntentEngine) executePlan(ctx context.Context, intent *Intent, plan *ExecutionPlan) {
	ie.log.Info("executing plan", "intent_id", intent.ID, "plan_id", plan.ID)
	ie.publishEvent("intent.executing", map[string]interface{}{
		"intent_id": intent.ID,
		"plan_id":   plan.ID,
	})

	result := ie.exec.Execute(ctx, plan)

	intent.Result = result
	if result.Success {
		intent.Status = IntentCompleted
		intent.CompletedAt = time.Now().UTC().Format(time.RFC3339)
		ie.log.Info("intent completed", "id", intent.ID)
		ie.publishEvent("intent.completed", map[string]interface{}{
			"intent_id": intent.ID,
			"summary":   result.Summary,
			"success":   true,
		})
	} else {
		intent.Status = IntentFailed
		intent.Error = result.Summary
		ie.log.Warn("intent failed", "id", intent.ID, "error", result.Summary)
		ie.publishEvent("intent.failed", map[string]interface{}{
			"intent_id": intent.ID,
			"summary":   result.Summary,
			"success":   false,
		})
	}
}

// GetIntent returns an intent by ID
func (ie *IntentEngine) GetIntent(id string) (*Intent, bool) {
	ie.mu.RLock()
	defer ie.mu.RUnlock()
	intent, ok := ie.intents[id]
	return intent, ok
}

// GetPlan returns an execution plan by ID
func (ie *IntentEngine) GetPlan(id string) (*ExecutionPlan, bool) {
	ie.mu.RLock()
	defer ie.mu.RUnlock()
	plan, ok := ie.plans[id]
	return plan, ok
}

// ListIntents returns all submitted intents
func (ie *IntentEngine) ListIntents() []*Intent {
	ie.mu.RLock()
	defer ie.mu.RUnlock()
	result := make([]*Intent, 0, len(ie.intents))
	for _, intent := range ie.intents {
		result = append(result, intent)
	}
	return result
}

// ListPlans returns all execution plans
func (ie *IntentEngine) ListPlans() []*ExecutionPlan {
	ie.mu.RLock()
	defer ie.mu.RUnlock()
	result := make([]*ExecutionPlan, 0, len(ie.plans))
	for _, plan := range ie.plans {
		result = append(result, plan)
	}
	return result
}

func (ie *IntentEngine) nextID() string {
	ie.mu.Lock()
	defer ie.mu.Unlock()
	ie.counter++
	return fmt.Sprintf("intent_%d", ie.counter)
}

func (ie *IntentEngine) publishEvent(eventType string, data map[string]interface{}) {
	if ie.bus != nil {
		ie.bus.Publish(context.Background(), events.Event{
			Type:    eventType,
			Source:  "intent.engine",
			Payload: data,
		})
	}
}
