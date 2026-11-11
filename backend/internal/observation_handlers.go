package hub

// This file wires observation triage actions onto the existing WebSocket hub. Payloads arrive
// from the browser, so they are validated here before changing persisted state (NH-03).

import (
	"errors"
	"strings"

	"node-herder/internal/ws"
	"node-herder/models/settings"
	"node-herder/utils"
)

// maxObservationLabelRunes bounds an operator-supplied display label.
const maxObservationLabelRunes = 120

// observationIdentityRequest is the validated form of a triage action payload.
type observationIdentityRequest struct {
	key      string
	decision settings.ObservationDecision
	label    string
}

// registerObservationHandlers binds the triage actions. The cache is the same instance the
// observation service reads, so a saved decision takes effect on the next snapshot.
func registerObservationHandlers(hub ws.EventHub, cache *settings.AppConfigCache) {
	if hub == nil || cache == nil {
		return
	}
	hub.OnSaveObservationIdentity(func(payload interface{}) error {
		request, err := parseObservationIdentityRequest(payload)
		if err != nil {
			return err
		}
		return cache.SaveObservationIdentity(request.key, settings.ObservationIdentity{
			Decision: request.decision,
			Label:    request.label,
		})
	})
	hub.OnDeleteObservationIdentity(func(payload interface{}) error {
		key, err := observationIdentityKey(payload)
		if err != nil {
			return err
		}
		return cache.DeleteObservationIdentity(key)
	})
}

// parseObservationIdentityRequest validates one save payload. An absent decision is rejected
// rather than defaulted, because the browser must state the intent it is persisting.
func parseObservationIdentityRequest(payload interface{}) (observationIdentityRequest, error) {
	fields, ok := payload.(map[string]interface{})
	if !ok {
		return observationIdentityRequest{}, errors.New("invalid observation identity payload")
	}
	key, err := observationIdentityKey(fields)
	if err != nil {
		return observationIdentityRequest{}, err
	}
	rawDecision, _ := fields["decision"].(string)
	decision := settings.ObservationDecision(rawDecision)
	if !settings.ValidObservationDecision(decision) {
		return observationIdentityRequest{}, errors.New("invalid observation decision")
	}
	label := ""
	if raw, ok := fields["label"].(string); ok {
		label = strings.TrimSpace(raw)
		if len([]rune(label)) > maxObservationLabelRunes {
			return observationIdentityRequest{}, errors.New("observation label is too long")
		}
	}
	return observationIdentityRequest{key: key, decision: decision, label: label}, nil
}

// observationIdentityKey extracts the opaque identity key from a payload.
func observationIdentityKey(payload interface{}) (string, error) {
	fields, ok := payload.(map[string]interface{})
	if !ok {
		return "", errors.New("invalid observation identity payload")
	}
	key, _ := fields["key"].(string)
	if key == "" || len(key) > 256 || strings.ContainsRune(key, '\x00') {
		return "", errors.New("invalid observation identity key")
	}
	return key, nil
}

// logObservationHandlerFailure keeps rejection reasons observable without echoing household
// identifiers into the log.
func logObservationHandlerFailure(action string, err error) {
	utils.LogErrorf("Observation %s rejected: %v", action, err)
}
