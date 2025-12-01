import json
import requests
from datetime import datetime

LLM_URL = "http://127.0.0.1:8080/v1/chat/completions"

def call_llm(messages):
    payload = {
        "model": "demo",
        "messages": messages
    }
    return requests.post(LLM_URL, json=payload).json()

def run_tool(name, args):
    if name == "get_last_event":
        return { "timestamp": "2025-01-01T10:00:00Z" }

    if name == "get_events_in_range":
        return { "events": [
            "2025-01-01T02:15:10Z",
            "2025-01-01T03:44:22Z"
        ] }

    return { "error": "unknown tool" }

messages = [
    {"role": "user", "content": "Was the front door opened between midnight and 4am?"}
]

# First call
resp = call_llm(messages)
tool_call = resp["choices"][0]["message"].get("tool_calls")

if tool_call:
    tc = tool_call[0]
    name = tc["function"]["name"]
    args = json.loads(tc["function"]["arguments"])

    result = run_tool(name, args)

    # Send result back
    messages.append({
        "role": "assistant",
        "tool_calls": tool_call
    })
    messages.append({
        "role": "tool",
        "tool_call_id": tc["id"],
        "content": json.dumps(result)
    })

    resp2 = call_llm(messages)
    print(resp2["choices"][0]["message"]["content"])
else:
    print(resp["choices"][0]["message"]["content"])
