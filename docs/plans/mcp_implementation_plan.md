# NodeHerder MCP Server Implementation Plan

## Overview

Add MCP server to nodeherder for LLM tool integration. Fresh implementation with clean design.

---

## Package Structure

```
internal/mcp/
├── server/
│   ├── server.go          # MCP server setup, transport
│   └── handler.go         # Request routing
├── tools/
│   ├── intent.go          # declare_intent tool
│   ├── metrics.go         # query_metrics tool
│   └── response.go        # Unified response types
├── resources/
│   ├── prompt.go          # system-prompt resource
│   └── devices.go         # devices resource
├── intent/
│   ├── types.go           # Intent struct
│   ├── parse.go           # JSON parsing
│   ├── validate.go        # Validation rules
│   └── translate.go       # Intent → Query translation
├── resolver/
│   ├── resolver.go        # Device resolution
│   ├── scoring.go         # Token scoring algorithm
│   └── errors.go          # AmbiguousDeviceError
└── timescope/
    ├── parse.go           # Scope string → TimeRange
    └── calendar.go        # today/yesterday with timezone
```

---

## Key Improvements Over Reference

### 1. Single Tool Response Type

```go
// Instead of different structs everywhere
type ToolResponse struct {
    Status     string      `json:"status"`  // success, error, ambiguous
    Data       any         `json:"data,omitempty"`
    Error      *ToolError  `json:"error,omitempty"`
    Candidates []Candidate `json:"candidates,omitempty"`
}
```

### 2. Simpler Intent Validation

```go
// Reference has complex nested checks. Simplify to rule list:
var validationRules = []Rule{
    RequireFields("target_name", "metrics"),
    RejectCountEventsForNumeric(),
    RejectRangeWithLatestValue(),
}

func Validate(intent Intent, ctx DeviceContext) error {
    for _, rule := range validationRules {
        if err := rule(intent, ctx); err != nil {
            return err
        }
    }
    return nil
}
```

### 3. Declarative Token Weights

```go
// Reference hardcodes weights. Use config:
var tokenConfig = TokenConfig{
    CommonWords: map[string]float64{
        "sensor": 0.15, "device": 0.15, "room": 0.4,
        "temperature": 0.3, "humidity": 0.3,
    },
    MinScore:       0.3,
    AmbiguityDelta: 0.15,
}
```

### 4. Timezone-Aware Time Parsing

```go
// Reference handles timezone inconsistently. Centralize:
type TimeScope struct {
    tz       *time.Location
    clock    func() time.Time  // injectable for testing
}

func (ts *TimeScope) Parse(scope string) *TimeRange {
    // All time calculations use ts.tz consistently
}
```

### 5. Table-Driven Aggregation Mapping

```go
// Instead of switch statement:
var aggregationMap = map[string]string{
    "count_events": "count",
    "latest_value": "last",
    "last_event":   "last_event",
    "min_value":    "min",
    "max_value":    "max",
    "avg_value":    "avg",
}
```

### 6. Binary Sensor Logic Encapsulated

```go
// Reference spreads this logic. Centralize:
type BinarySensorMapper struct {
    OnValue  any
    OffValue any
}

func (m *BinarySensorMapper) MapOutcome(positive bool) any {
    if positive {
        return m.OnValue
    }
    return m.OffValue
}
```

---

## MCP Tools

| Tool             | Purpose                                            |
| ---------------- | -------------------------------------------------- |
| `declare_intent` | Full pipeline: validate → resolve → query → format |
| `query_metrics`  | Direct query with known device_id                  |

## MCP Resources

| URI                          | Content              |
| ---------------------------- | -------------------- |
| `nodeherder://system-prompt` | Domain rules for LLM |
| `nodeherder://devices`       | Device context JSON  |

---

## Implementation Phases

### Phase 1: Core Types & Interfaces

- [ ] Create package skeleton
- [ ] Define Intent, TimeRange, ToolResponse types
- [ ] Add `github.com/mark3labs/mcp-go` dependency

### Phase 2: Time & Resolver Logic

- [ ] Implement `timescope/` package
- [ ] Implement `resolver/` package with scoring
- [ ] Unit tests (≥90% coverage)

### Phase 3: Intent Logic

- [ ] Implement `intent/` package (parse, validate, translate)
- [ ] Unit tests

### Phase 4: MCP Server

- [ ] Implement `server/` package
- [ ] Implement `resources/` package
- [ ] Implement `tools/` package
- [ ] Wire to existing nodeherder APIs

### Phase 5: Testing

- [ ] Integration tests with MCP Inspector
- [ ] End-to-end tests

---

## Dependencies

| Package                       | Purpose          |
| ----------------------------- | ---------------- |
| `github.com/mark3labs/mcp-go` | MCP protocol SDK |

---

## Running

```bash
./nodeherder --mcp  # stdio mode for LLM clients
```

---

## Success Criteria

- [ ] `tools/list` returns both tools
- [ ] `declare_intent` works end-to-end
- [ ] All tests pass ≥85% coverage
