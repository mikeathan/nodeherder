import requests
import json
import readline  # optional, improves input UX

LLM_URL = "http://192.168.50.60:5001/v1/chat/completions"

# ---------------------------------------------------------------------------
# Fake backend tools
# ---------------------------------------------------------------------------

def mock_backend(name, args):
    if name == "get_last_event":
        return {
            "timestamp": "2025-02-01T13:00:00Z",
            "device": args.get("device_id", "unknown")
        }

    if name == "get_temp":
        return {
            "temperature": 21.4,
            "unit": "Celsius"
        }

    return {"error": f"unknown tool '{name}'"}

# ---------------------------------------------------------------------------
# Tools schema sent to LLM
# ---------------------------------------------------------------------------

TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "get_last_event",
            "description": "Returns last event timestamp for a given device.",
            "parameters": {
                "type": "object",
                "properties": {
                    "device_id": {"type": "string"}
                },
                "required": ["device_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_temp",
            "description": "Returns the current fake temperature.",
            "parameters": {
                "type": "object",
                "properties": {},
                "required": []
            }
        }
    }
]

# ---------------------------------------------------------------------------
# LLM call helper
# ---------------------------------------------------------------------------

def call_llm(messages, include_tools=True):
    payload = {
        "model": "local",
        "messages": messages,
        "tool_choice": "auto"
    }
    if include_tools:
        payload["tools"] = TOOLS

    r = requests.post(LLM_URL, json=payload)
    return r.json()

# ---------------------------------------------------------------------------
# Interactive loop
# ---------------------------------------------------------------------------

def main():
    print("Interactive LLM tool-calling demo. Type your query. Ctrl+C to quit.\n")

    while True:
        user_input = input("> ")

        # Build conversation history
        messages = [
            {"role": "user", "content": user_input}
        ]

        # FIRST CALL — model decides whether to call a tool
        resp = call_llm(messages, include_tools=True)

        try:
            message = resp["choices"][0]["message"]
        except:
            print("LLM error:", resp)
            continue

        # If model wants to call a tool
        if "tool_calls" in message:
            tc = message["tool_calls"][0]
            tool_name = tc["function"]["name"]
            args = json.loads(tc["function"]["arguments"])

            print(f"\nLLM requested tool: {tool_name}")
            print("Arguments:", args)

            # Run the fake backend tool
            result = mock_backend(tool_name, args)
            print("Mock backend returned:", result)

            # Append tool result into the conversation
            messages.append(message)
            messages.append({
                "role": "tool",
                "tool_call_id": tc["id"],
                "content": json.dumps(result)
            })

            # SECOND CALL — LLM completes answer
            resp2 = call_llm(messages, include_tools=False)
            final_msg = resp2["choices"][0]["message"]["content"]
            print("\nFinal answer:", final_msg)
            print("\n" + "-"*60 + "\n")

        else:
            # LLM didn't call a tool
            print("\nAssistant:", message["content"])
            print("\n" + "-"*60 + "\n")


if __name__ == "__main__":
    main()
