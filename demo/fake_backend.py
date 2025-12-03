import requests
import json
import readline  # optional

LLM_URL = "http://192.168.50.60:5001/v1/chat/completions"

# ---------------------------------------------------------------------------
# Fake backend tools
# ---------------------------------------------------------------------------

def mock_backend(name, args):
    if name == "get_last_event":
        device = args.get("device_id", "front_door")
        return {
            "timestamp": "2025-02-01T13:00:00Z",
            "device": device
        }

    if name == "get_temp":
        return {
            "temperature": 21.4,
            "unit": "Celsius"
        }

    return {"error": f"unknown tool '{name}'"}

# ---------------------------------------------------------------------------
# Tools schema (device_id NOT required)
# ---------------------------------------------------------------------------

TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "get_last_event",
            "description": "Returns last event timestamp for a given device. If missing, assume front_door.",
            "parameters": {
                "type": "object",
                "properties": {
                    "device_id": {"type": "string"}
                },
                "required": []     # <---- FIXED
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
    print("Interactive LLM tool-calling demo. Ctrl+C to quit.\n")

    while True:
        user_input = input("> ")

        messages = [
            {
                "role": "system",
                "content": (
                    "You MUST use the provided tools when the user asks for "
                    "information that a tool can retrieve. "
                    "Never ask the user for missing parameters. "
                    "If a device_id is missing, assume device_id='front_door'. "
                    "Return only valid JSON when calling tools."
                )
            },
            {
                "role": "user",
                "content": user_input
            }
        ]

        # FIRST CALL — model decides whether to call a tool
        resp = call_llm(messages, include_tools=True)

        try:
            message = resp["choices"][0]["message"]
        except Exception:
            print("LLM error:", resp)
            continue

        # ------------------------------------------------------------------
        # MODEL CALLS A TOOL
        # ------------------------------------------------------------------
        if "tool_calls" in message:
            tc = message["tool_calls"][0]
            tool_name = tc["function"]["name"]

            # Handles cases where the model returns empty args like "{}"
            try:
                args = json.loads(tc["function"]["arguments"])
            except:
                args = {}

            print(f"\nLLM requested tool: {tool_name}")
            print("Arguments:", args)

            # Execute fake backend
            result = mock_backend(tool_name, args)
            print("Mock backend returned:", result)

            # Prepare messages for second call
            messages.append(message)
            messages.append({
                "role": "tool",
                "tool_call_id": tc["id"],
                "content": json.dumps(result)
            })

            # SECOND CALL — completion
            resp2 = call_llm(messages, include_tools=False)

            try:
                final_msg = resp2["choices"][0]["message"]["content"]
            except:
                print("LLM error:", resp2)
                continue

            print("\nFinal answer:", final_msg)
            print("\n" + "-"*60 + "\n")
            continue

        # ------------------------------------------------------------------
        # MODEL DID NOT CALL A TOOL
        # ------------------------------------------------------------------
        else:
            print("\nAssistant:", message.get("content"))
            print("\n" + "-"*60 + "\n")


if __name__ == "__main__":
    main()
