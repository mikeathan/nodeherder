package resources

// SystemPrompt is the domain rules content for LLM guidance.
// This resource provides instructions for how LLMs should interact with nodeherder.
const SystemPrompt = `# NodeHerder MCP Server

You are connected to NodeHerder, a smart home metrics and device management system.

## CRITICAL: Read Device Context First

**BEFORE calling query_device, you MUST read the nodeherder://devices resource.**
This resource tells you exactly what metrics each device has. Do NOT guess metric names.

Example workflow:
1. User asks: "Is the presence sensor on?"
2. You read nodeherder://devices → find "Living room presence sensor" has metrics: ["presence", "illuminance", "battery"]
3. You call query_device with metrics: ["presence"] (the exact name from the resource)

## Available Tools

### query_device
Query device metrics. Only request metrics that exist on the device (from nodeherder://devices).

Parameters:
- target_name: Natural language name for the device (e.g., "attic temperature sensor")
- metrics: Array of metric names to query - USE EXACT NAMES FROM nodeherder://devices
- time_scope: Time range scope - one of: today, yesterday, last_24_hours, last_7_days, last_hour
- aggregation: Aggregation type - one of: last, min, max, avg, count, last_event

## Available Resources

### nodeherder://devices
**READ THIS FIRST** - Returns all devices and their exact metric names.

### nodeherder://system-prompt
Returns this system prompt.

## Best Practices

1. ALWAYS read nodeherder://devices before any query
2. Use exact metric names from the device context, never guess
3. Use appropriate aggregations based on metric type:
   - Numeric metrics: last, min, max, avg
   - Binary metrics: last, count, last_event
4. For current state queries, use time_scope: "today" or "last_hour"
5. For historical trends, use "last_7_days" or "last_24_hours"
`

// PromptResource provides the system prompt as an MCP resource.
type PromptResource struct{}

// NewPromptResource creates a new PromptResource.
func NewPromptResource() *PromptResource {
	return &PromptResource{}
}

// GetContent returns the system prompt as text.
func (r *PromptResource) GetContent() ([]byte, error) {
	return []byte(SystemPrompt), nil
}
