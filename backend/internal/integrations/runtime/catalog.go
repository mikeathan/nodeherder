package runtime

// Derives read-time catalog summaries from configured instances.

import model "node-herder/models/integrations"

// catalogInstanceStatus maps one configured instance to its worst presentation state.
func catalogInstanceStatus(instance model.Instance) string {
	if instance.LastError != nil {
		return "error"
	}
	if !instance.Enabled {
		return "disabled"
	}
	if instance.Snapshot == nil {
		return "pending"
	}
	return "ready"
}

// catalogWorstStatus keeps the most severe status across configured instances of one type.
func catalogWorstStatus(current, next string) string {
	if catalogStatusRank(next) > catalogStatusRank(current) {
		return next
	}
	return current
}

// catalogStatusRank orders statuses by presentation severity.
func catalogStatusRank(value string) int {
	switch value {
	case "error":
		return 3
	case "disabled":
		return 2
	case "pending":
		return 1
	case "ready":
		return 0
	default:
		return -1
	}
}
