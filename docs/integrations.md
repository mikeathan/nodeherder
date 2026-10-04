# API, MCP, and assistant integrations

[Project overview](../README.md) · [Setup](setup.md)

## MCP

Enable MCP in Settings → MCP Server. It initializes from saved configuration and
defaults disabled; the current CLI has no `--mcp` or `--mcp-only` option.

| Endpoint | Method | Purpose |
| --- | --- | --- |
| `/api/mcp` | POST | JSON-RPC requests |
| `/api/mcp`, `/api/mcp/events` | GET | SSE notifications |

These routes are currently public; restrict network exposure as described in
[deployment](setup.md#deployment-and-access).

Discover the actual registered resources/tools:

```bash
curl -sS http://localhost:4110/api/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"resources/list"}'

curl -sS http://localhost:4110/api/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
```

The registered resource is `nodeherder://system-prompt`, which includes domain guidance
and device context. Read it before choosing exact metric names:

```bash
curl -sS http://localhost:4110/api/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"nodeherder://system-prompt"}}'
```

`query_device` resolves a device name and queries metrics:

```bash
curl -sS http://localhost:4110/api/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"query_device","arguments":{"target_name":"Living room presence sensor","metrics":["presence"],"aggregation":"last"}}}'
```

Replace target/metrics with values from your hub. Ambiguous targets return candidates;
do not assume one was selected. Discover current schemas through `tools/list`.

For update notifications, keep an SSE connection open:

```bash
curl -N http://localhost:4110/api/mcp/events
```

Resource changes publish `notifications/resources/updated` for the system-prompt URI.
The server accepts subscribe/unsubscribe requests; subscriptions are handled by the SSE
connection. The old `nodeherder://devices` examples do not match current registration.
MCP tests: `go test ./internal/mcp/...` from `backend/`.

## Authenticated metrics API

Configure service credentials using the [setup guide](setup.md#backend-environment).
Request a bearer token; placeholders below are not usable credentials:

```bash
curl -sS http://localhost:4110/api/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"llm-proxy","client_secret":"<your-service-secret>"}'
```

Use the returned `access_token` in a metrics request:

```bash
curl -sS http://localhost:4110/api/metrics/query \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <access_token>' \
  -d '{"deviceIds":["<device-id>"],"exposes":["temperature"],"time":{"lookback":"1h"},"aggregation":"last"}'
```

Replace the device/expose with your own. Queries also accept explicit time bounds,
filters, limits, and ordering; see the
[request model](../backend/internal/metrics/query/request.go).
Metrics must have been enabled/recorded for historical queries to contain samples.
Device context is also available at `GET /api/context/devices` (currently public).

## Assistant chat

Set the external assistant URL in Settings → Assistant. Browser messages go through
`POST /api/assistant/message`; the backend forwards JSON to that URL and stores
conversation history. The frontend sends `conversation_id`, `context_version`, and
`message`, and expects a response with `reply`.

The external service must implement this contract; a model provider's generic API
URL is not necessarily compatible. Conversation listing/history/deletion use protected
`/api/assistant/conversations` and `/api/assistant/history/:id` routes.
