# MCP Server Architecture

NodeHerder includes an MCP (Model Context Protocol) server for LLM tool integration. This allows AI assistants to query device metrics and understand your smart home state.

## What is MCP?

MCP (Model Context Protocol) is a standard protocol for LLMs to interact with external tools and resources. It uses JSON-RPC over stdio, allowing AI assistants like Claude to:

- **Call tools** to perform actions (query metrics, resolve devices)
- **Read resources** to get context (device list, system prompts)

## Startup Flow

```
main.go
   │
   └─► hub.Register(port, store, ctx, enableMCP)
          │
          ├─► Start WebSocket hub
          ├─► Connect MQTT
          ├─► Initialize automations
          │
          ├─► if enableMCP:
          │      └─► go startMCPServer(store)  ◄── runs in background goroutine
          │
          └─► Start HTTP API server
```

The MCP server runs alongside the HTTP server in a separate goroutine, sharing the same data store.

## Command Line Options

```bash
# Default: HTTP + MCP both running
./nodeherder

# Disable MCP server
./nodeherder --mcp=false

# Custom port (MCP still runs on stdio)
./nodeherder --port=8080
```

## Package Structure

```
internal/mcp/
├── server/           # MCP protocol layer
│   └── server.go       # mcp-go SDK integration, tool/resource registration
├── tools/            # Tool implementations
│   └── intent.go       # query_device handler
├── resources/        # Resource providers
│   ├── prompt.go       # System prompt for LLM guidance
│   └── devices.go      # Device context JSON
├── intent/           # Intent processing
│   ├── types.go        # Intent struct definition
│   ├── parse.go        # JSON parsing
│   ├── validate.go     # Validation rules
│   └── translate.go    # Intent → Query conversion
├── resolver/         # Device resolution
│   ├── resolver.go     # Main resolver logic
│   ├── scoring.go      # Token-based matching algorithm
│   └── errors.go       # AmbiguousDeviceError, DeviceNotFoundError
├── timescope/        # Time handling
│   ├── parse.go        # Scope string parsing (today, yesterday, etc.)
│   └── calendar.go     # Timezone-aware helpers
└── types.go          # Shared types (ToolResponse, etc.)
```

---

## MCP Tools

### query_device

**Purpose**: Query metrics using natural language device descriptions.

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| `target_name` | Yes | Natural language device name (e.g., "attic temperature sensor") |
| `metrics` | Yes | Array of metric names (e.g., `["temperature", "humidity"]`) |
| `time_scope` | No | Time range: `today`, `yesterday`, `last_24_hours`, `last_7_days`, `last_hour` |
| `aggregation` | No | `latest_value`, `min_value`, `max_value`, `avg_value`, `count_events` |

**Example**:

```json
{
  "target_name": "attic temperature",
  "metrics": ["temperature"],
  "time_scope": "today",
  "aggregation": "latest_value"
}
```

> **Note**: You can also pass exact device IDs as `target_name` - the resolver will match them directly.

---

## MCP Resources

### nodeherder://system-prompt

Returns LLM guidance text explaining:

- Available tools and their parameters
- Best practices for querying
- Aggregation type recommendations

### nodeherder://devices

Returns JSON with all available devices and their metrics:

```json
{
  "version": "1",
  "devices": [
    {
      "id": "0xa4c138e1b5658e68",
      "name": "Attic air sensor",
      "exposes": [
        { "name": "temperature", "type": "numeric", "unit": "°C" },
        { "name": "humidity", "type": "numeric", "unit": "%" }
      ]
    }
  ]
}
```

---

## Data Flow: query_device

```
┌─────────────────────────────────────────────────────────────────┐
│                        LLM Request                               │
│  {"target_name": "attic temp", "metrics": ["temperature"]}      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. PARSE (intent/parse.go)                                      │
│     JSON → Intent struct                                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. VALIDATE (intent/validate.go)                                │
│     - RequireTargetName                                          │
│     - RequireMetrics                                             │
│     - RejectCountEventsForNumeric                                │
│     - RejectRangeWithLatestValue                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. RESOLVE (resolver/)                                          │
│     "attic temp" → tokenize → score against all device names    │
│     → best match: "Attic temperature sensor" (device_id)        │
│                                                                  │
│     Scoring: token weights (common words like "sensor" = 0.15,  │
│              specific words like "attic" = 1.0)                  │
│                                                                  │
│     If ambiguous (multiple close scores) → return candidates    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  4. TRANSLATE (intent/translate.go)                              │
│     Intent + device_id → MetricsQueryRequest                     │
│     - Map aggregation: "latest_value" → AggLast                  │
│     - Parse time_scope: "today" → {from: midnight, to: now}      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  5. EXECUTE (metrics.QueryService)                               │
│     Query database with MetricsQueryRequest                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Response                                  │
│  {"status": "success", "data": {"temperature": 22.5}}           │
└─────────────────────────────────────────────────────────────────┘
```

---

## Device Resolution Algorithm

The resolver uses token-based scoring to match natural language to device names:

1. **Tokenize** both the target name and each device name
2. **Weight** each token (common words get lower weight):
   ```go
   "sensor": 0.15, "device": 0.15, "room": 0.4,
   "temperature": 0.3, "humidity": 0.3, "light": 0.4
   ```
3. **Score** = matched_weight / total_weight
4. **Ambiguity check**: If top two scores are within 0.15, return both as candidates
5. **Minimum threshold**: Score must be ≥ 0.3 to match

---

## Time Scopes

| Scope           | From                | To             |
| --------------- | ------------------- | -------------- |
| `today`         | Midnight (local TZ) | Now            |
| `yesterday`     | Yesterday midnight  | Today midnight |
| `last_24_hours` | Now - 24h           | Now            |
| `last_7_days`   | Now - 7d            | Now            |
| `last_hour`     | Now - 1h            | Now            |
| `last_30_days`  | Now - 30d           | Now            |

---

## Response Format

All tool responses use a unified format:

```go
type ToolResponse struct {
    Status     string      // "success", "error", "ambiguous"
    Data       any         // Query results (on success)
    Error      *ToolError  // Error details (on error)
    Candidates []Candidate // Device options (on ambiguous)
}
```

**Success**:

```json
{"status": "success", "data": {...}}
```

**Error**:

```json
{ "status": "error", "error": { "code": "validation_error", "message": "..." } }
```

**Ambiguous**:

```json
{
  "status": "ambiguous",
  "candidates": [
    { "device_id": "...", "name": "Attic temperature sensor", "score": 0.85 },
    { "device_id": "...", "name": "Attic air sensor", "score": 0.8 }
  ]
}
```

---

## End-to-End Integration

When NodeHerder is used with a frontend and LLM proxy, here's the complete flow:

```
┌──────────────┐     ┌──────────────┐     ┌─────────┐     ┌────────────┐
│   Frontend   │────►│  LLM Proxy   │────►│   LLM   │────►│ NodeHerder │
│  (Vue.js)    │◄────│              │◄────│ (Claude)│◄────│ MCP Server │
└──────────────┘     └──────────────┘     └─────────┘     └────────────┘
```

### Step-by-Step Flow

1. **User asks question** in frontend:

   > "What's the temperature in the attic?"

2. **Frontend → LLM Proxy**: Sends the question via HTTP/WebSocket

3. **LLM Proxy → LLM (Claude)**: Forwards the question with MCP tools configured

4. **LLM decides to use MCP**: Claude sees `query_device` tool and calls it:

   ```json
   {
     "target_name": "attic",
     "metrics": ["temperature"],
     "aggregation": "latest_value"
   }
   ```

5. **LLM Proxy → NodeHerder MCP**: The proxy spawns/connects to nodeherder's MCP server via **stdio** and sends the tool call

6. **NodeHerder executes**: Resolves "attic" → device, queries temperature, returns result

7. **LLM formats response**: Claude turns the data into natural language:

   > "The attic is currently 22.5°C"

8. **Response returns** through proxy to frontend

### LLM Proxy Configuration

The LLM proxy needs to be configured to spawn nodeherder as an MCP server. Example MCP config:

```json
{
  "mcpServers": {
    "nodeherder": {
      "command": "./nodeherder",
      "args": [],
      "env": {}
    }
  }
}
```

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND                                    │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Chat UI: "What's the temperature in the attic?"                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼ HTTP/WebSocket
┌─────────────────────────────────────────────────────────────────────────┐
│                            LLM PROXY                                     │
│  ┌──────────────────────┐    ┌──────────────────────────────────────┐   │
│  │  API Gateway         │    │  MCP Client                          │   │
│  │  - Auth              │    │  - Spawns nodeherder process         │   │
│  │  - Rate limiting     │    │  - Sends/receives JSON-RPC via stdio │   │
│  └──────────────────────┘    └──────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
          │                                      │
          ▼ LLM API                              ▼ stdio (JSON-RPC)
┌─────────────────────┐                ┌──────────────────────────────────┐
│        LLM          │                │         NODEHERDER               │
│  (Claude/GPT/etc)   │                │  ┌────────────────────────────┐  │
│                     │ ◄─tool call──► │  │  MCP Server (goroutine)    │  │
│  - Understands      │                │  │  - query_device            │  │
│    user intent      │                │  └────────────────────────────┘  │
│  - Calls MCP tools  │                │                │                 │
└─────────────────────┘                │                ▼                 │
                                       │  ┌────────────────────────────┐  │
                                       │  │  Metrics Store (BoltDB)    │  │
                                       │  └────────────────────────────┘  │
                                       └──────────────────────────────────┘
```

### Key Points

- **MCP uses stdio**: The LLM proxy connects to nodeherder via stdin/stdout, not HTTP
- **Shared data store**: MCP server runs in the same process as HTTP server, shares the database
- **LLM makes decisions**: The LLM decides when to call tools based on user questions
- **Stateless queries**: Each tool call is independent, no session state needed
