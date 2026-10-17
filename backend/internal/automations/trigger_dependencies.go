package automations

// triggerDependencies is a read-only index of fields used by a configured recipe.
// It is built before the recipe is published and replaced with that generation,
// rather than maintained as separate mutable automation state. Most sensor packets
// can then be rejected with a map lookup, before allocating a run context or testing
// conditions. Trigger.AffectedBy remains the final authority for actual execution.
type triggerDependencies struct {
	names    map[string]struct{}
	fallback bool
}

func newTriggerDependencies(triggers TriggerList) *triggerDependencies {
	inputs := &triggerDependencies{names: make(map[string]struct{})}
	for _, trigger := range triggers {
		if _, known := trigger.(*DeviceTrigger); !known {
			inputs.fallback = true
		}
		inputs.names[trigger.GetName()] = struct{}{}
		for _, condition := range trigger.GetConditions() {
			inputs.addCondition(condition)
		}
	}
	return inputs
}

// Unknown extensions keep full routing because their dependencies cannot be inferred safely.
func (d *triggerDependencies) addCondition(condition Condition) {
	switch c := condition.(type) {
	case *ExposeCondition:
		d.names[c.Name] = struct{}{}
	case *TimeCondition: // No device-state dependency.
	default:
		d.fallback = true
	}
}

func (d *triggerDependencies) accepts(changes map[string]interface{}) bool {
	if d == nil || d.fallback {
		return true
	}
	for property := range changes {
		if _, ok := d.names[property]; ok {
			return true
		}
	}
	return false
}
