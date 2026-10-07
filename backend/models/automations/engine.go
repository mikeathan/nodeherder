package automations

import (
	"context"
	"errors"
)

// ErrLaneBusy reports that a manual trigger was not started because the
// device's processing lane was full, shut down, or did not reach it before the
// caller's deadline. The trigger did not run.
var ErrLaneBusy = errors.New("device is busy, trigger not started")

type AutomationQuerier interface {
	IsAutomationEnabled(id string) bool
}

type AutomationTrigger interface {
	TriggerManual(ctx context.Context, automationId string, triggerName string) error
}
