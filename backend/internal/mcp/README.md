# MCP Server Architecture

NodeHerder includes an MCP (Model Context Protocol) server exposed via HTTP and SSE (Server-Sent Events). This allows AI assistants and other clients to query device metrics and receive real-time updates.

## Protocol Support

- **HTTP POST** (`/api/mcp`): Standard JSON-RPC 2.0 endpoint for tool calls and resource reading.
- **SSE** (`/api/mcp/events`): Notifications for resource updates (e.g., system prompt changes).

## Startup Flow

The MCP server runs as part of the main `nodeherder` process:

```
main.go
   │
   └─► hub.Register(port, store, ctx, ...enableMCP=true)
          │
          ├─► Initialize MCP Server (server.New)
          │
          └─► Register HTTP Routes:
                 ├─► POST /api/mcp
                 └─► GET  /api/mcp/events
```

## Configuration

The MCP server is enabled by default. You can control it via flags:

```bash
# Default: HTTP API + MCP enabled
./nodeherder

# Disable MCP endpoints
./nodeherder --mcp=false
```

## Testing with MCP Inspector

The MCP Inspector is a free tool from Anthropic for testing MCP servers without needing an LLM. It provides a web UI to interactively test tools and resources.

### Prerequisites

- Node.js installed (for npx)
- Backend built: `npx @modelcontextprotocol/inspector ./nodeherder --mcp-only`

### Running the Inspector

```bash
# From the backend directory
cd /path/to/nodeherder/backend
go build -o nodeherder .
npx @modelcontextprotocol/inspector ./nodeherder --mcp-only
```

This opens a web UI at `http://localhost:5173` (or similar).

### Using the Inspector

1. **Resources Tab**: View available resources
   - Click on `nodeherder://system-prompt` to see LLM guidance

2. **Tools Tab**: Test the `query_device` tool
   - Click on `query_device` to see its schema
   - Fill in test parameters:
     ```json
     {
       "target_name": "attic temperature",
       "metrics": ["temperature"],
       "time_scope": "today",
       "aggregation": "latest_value"
     }
     ```
   - Ambiguous: `{"status": "ambiguous", "candidates": [...]}`

---

## HTTP API Examples

You can interact with the MCP server directly using `curl` if the HTTP API is enabled.

### 1. List Resources

```bash
curl -X POST http://localhost:4110/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "resources/list"
  }'
```

### 2. Get System Prompt (and Device Context)

```bash
curl -X POST http://localhost:4110/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "resources/read",
    "params": {
      "uri": "nodeherder://system-prompt"
    }
  }'
```

### 3. Call Tool (query_device)

```bash
curl -X POST http://localhost:4110/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "query_device",
      "arguments": {
        "target_name": "attic temperature",
        "metrics": ["temperature"],
        "time_scope": "today",
        "aggregation": "last"
      }
    }
  }'
```

### Running Unit Tests

```bash
# Run all MCP tests
go test ./internal/mcp/... -v

# Run specific package tests
go test ./internal/mcp/tools/... -v
go test ./internal/mcp/resolver/... -v
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

**Example Usage**:

```bash
curl -X POST http://localhost:4110/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "resources/read",
    "params": {"uri": "nodeherder://system-prompt"}
  }' | jq '.result.contents[0].text'
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

5. **NodeHerder executes**: Resolves "attic" → device, queries temperature, returns result

6. **LLM formats response**: Claude turns the data into natural language:

   > "The attic is currently 22.5°C"

7. **Response returns** through proxy to frontend

The LLM proxy needs to be configured to connect to nodeherder as a remote MCP server (SSE). Example config:

```json
{
  "mcpServers": {
    "nodeherder": {
      "url": "http://localhost:4110/api/mcp/events",
      "transport": "sse"
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
│  │  - Auth              │    │  - Connects to nodeherder via HTTP   │   │
│  │  - Rate limiting     │    │  - Sends/receives JSON-RPC over HTTP │   │
│  └──────────────────────┘    └──────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
          │                                      │
          ▼ LLM API                              ▼ HTTP (JSON-RPC)
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

- **MCP uses HTTP/SSE**: The LLM proxy connects to nodeherder via HTTP endpoints
- **Shared data store**: MCP server runs in the same process as HTTP server, shares the database
- **LLM makes decisions**: The LLM decides when to call tools based on user questions
- **Stateless queries**: Each tool call is independent, no session state needed

---
