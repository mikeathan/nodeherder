package automations

type AutomationQuerier interface {
	IsAutomationEnabled(id string) bool
}
