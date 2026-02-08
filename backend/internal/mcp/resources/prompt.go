package resources

import (
	"fmt"
	"log"
	"node-herder/models/devices"
	"strings"
	"sync"
)

// SystemPrompt is the domain rules content for LLM guidance.
// This resource provides instructions for how LLMs should interact with nodeherder.
const SystemPrompt = `# NodeHerder MCP Server

You are connected to NodeHerder, a smart home metrics system.

## CRITICAL: Context Awareness

The following devices are available in this home:
%s

**BEFORE calling query_device, CHECK the list above.**
This resource tells you exactly what metrics each device has. Do NOT guess metric names.
Use the EXACT "Friendly Name" from the list for the 'target_name' parameter.

Example workflow:
1. User asks: "Is the presence sensor on?"
2. You check the available devices list above -> find "Living room presence sensor" has metrics: ["presence", "illuminance", "battery"]
3. You call query_device with metrics: ["presence"] (the exact name from the resource)

## Available Tools

### query_device
Query device metrics. Only request metrics that exist on the device.
NOTE: This tool is READ-ONLY. It cannot control devices (e.g., cannot turn lights on/off).

Parameters:
- target_name: Natural language name for the device (e.g., "attic temperature sensor"). Matches 'Friendly Name'.
- metrics: Array of metric names to query - USE EXACT NAMES FROM THE LIST ABOVE
- time_scope: Time range scope - one of: today, yesterday, last_24_hours, last_7_days, last_hour
- aggregation: Use "last" for current/latest value. Options: last, min, max, avg, count

## Handling Tool Responses

1. **Success**: Use the returned data to answer the user's question.
2. **Ambiguous**: If the tool returns a status of "ambiguous", it will provide a list of candidates. ASK the user to clarify which device they meant from the list.
3. **Empty/Error with Hint**: If the tool returns empty results but provides a "hint", use that hint to retry with the correct metric names if applicable, or inform the user about available metrics.

## CRITICAL: Interpreting Boolean (true/false) Values

For binary sensors (alarm, contact, presence, motion, etc.), the query returns a boolean value:
- **true** = The metric IS active/triggered (alarm is sounding, door is open, presence detected)
- **false** = The metric is NOT active (alarm is silent/off, door is closed, no presence)

**IMPORTANT**: The metric name (e.g., "alarm") describes WHAT is being measured, NOT the state.
The "value" field contains the actual state (true/false).

Example: If querying "alarm" returns {"value": false}, report: "The alarm is OFF/silent" (NOT "the state is alarm").

## Available Resources

### nodeherder://system-prompt
Returns this system prompt.

## Best Practices

1. Use exact metric names from the device context, never guess
2. Aggregations (pick ONE):
   - **last**: Use for ALL "current value" or "when did X change" questions. Works for numeric AND binary.
   - min/max/avg: Numeric only (temperature, humidity, etc.)
   - count: Number of readings
3. All timestamps are in **UTC**. formatted_time is RFC3339 (e.g., 2026-02-07T19:01:05Z). Always mention UTC when giving times.
4. For current state queries, use time_scope: "today" or "last_hour"
5. For historical trends, use "last_7_days" or "last_24_hours"
`

// PromptResource provides the system prompt as an MCP resource.
type PromptResource struct {
	store         DeviceStore
	cachedSummary []byte
	mu            sync.RWMutex
}

// NewPromptResource creates a new PromptResource.
func NewPromptResource(store DeviceStore) *PromptResource {
	r := &PromptResource{store: store}
	r.buildCache()

	store.RegisterIsDirtyCallback(func() {
		r.buildCache()
	})

	return r
}

func (r *PromptResource) buildCache() {
	state, err := r.store.LoadHubState()
	if err != nil {
		log.Printf("Error loading hub state for prompt cache: %v", err)
		return
	}

	deviceSummary := buildDeviceSummary(state.Devices)
	content := fmt.Appendf(nil, SystemPrompt, deviceSummary)

	r.mu.Lock()
	r.cachedSummary = content
	r.mu.Unlock()
}

// GetContent returns the system prompt as text.
func (r *PromptResource) GetContent() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cachedSummary == nil {
		// Fallback if cache is empty (shouldn't happen if initialized correctly, but good for safety)
		// Or if initial load failed. Try one more time?
		// For now, let's just return an error or try to rebuild.
		r.mu.RUnlock()
		r.buildCache()
		r.mu.RLock()
		if r.cachedSummary == nil {
			return nil, fmt.Errorf("failed to load system prompt content")
		}
	}

	return r.cachedSummary, nil
}

func buildDeviceSummary(devs []*devices.Device) string {
	var b strings.Builder
	for _, d := range devs {
		b.WriteString("- ")
		b.WriteString(d.FriendlyName) // Use Friendly Name as primary
		b.WriteString(" (")
		b.WriteString(d.Id)
		b.WriteString(") | exposes: ")

		exposes := make([]string, 0)
		for _, e := range d.Exposes {
			desc := e.Name
			if valOn, ok := e.Values["on"]; ok {
				desc = fmt.Sprintf("%s(on=%v)", e.Name, valOn)
			}
			exposes = append(exposes, desc)
		}
		b.WriteString(strings.Join(exposes, ", "))
		b.WriteString("\n")
	}
	return b.String()
}
