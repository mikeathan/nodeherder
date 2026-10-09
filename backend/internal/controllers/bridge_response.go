package controllers

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/ws"
)

type bridgeDeviceResponseData struct {
	ID string `json:"id"`
}

type bridgeRenameResponseData struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type decodedBridgeResponse[T any] struct {
	Status      string
	Error       string
	Transaction string
	Data        T
}

// Decode success data only for successful responses: error envelopes need not
// contain the operation's success fields. Unknown fields remain compatible.
func decodeBridgeResponse[T any](topic string, payload []byte) (decodedBridgeResponse[T], error) {
	var envelope struct {
		Status      string          `json:"status"`
		Error       string          `json:"error"`
		Transaction string          `json:"transaction"`
		Data        json.RawMessage `json:"data"`
	}
	var response decodedBridgeResponse[T]
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return response, fmt.Errorf("%s: decode response: %w", topic, err)
	}
	if envelope.Status != "ok" && envelope.Status != "error" {
		return response, fmt.Errorf("%s: missing or invalid response status", topic)
	}
	response.Status, response.Error, response.Transaction = envelope.Status, envelope.Error, envelope.Transaction
	if envelope.Status == "ok" && len(envelope.Data) != 0 {
		if err := json.Unmarshal(envelope.Data, &response.Data); err != nil {
			return response, fmt.Errorf("%s: decode response data: %w", topic, err)
		}
	}
	return response, nil
}

func broadcastBridgeResult(hub ws.EventHub, topic, event, message string) error {
	if err := hub.Broadcast(event, message); err != nil {
		return fmt.Errorf("%s: broadcast %s: %w", topic, event, err)
	}
	return nil
}
