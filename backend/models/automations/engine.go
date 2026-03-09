package automations

type AutomationQuerier interface {
	IsAutomationEnabled(id string) bool
}

type AutomationTrigger interface {
	TriggerManual(automationId string, triggerName string) error
}
