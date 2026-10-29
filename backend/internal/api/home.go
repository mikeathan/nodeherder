package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"node-herder/internal/auth"
	"node-herder/internal/integrations"
	"node-herder/models/devices"
	"node-herder/models/settings"
)

// HomeLayouts is the handler's revisioned placement boundary; persistence belongs to the cache.
type HomeLayouts interface {
	LoadHomeLayout() (settings.HomeLayout, error)
	SaveHomeLayout(settings.HomeLayout, string) (settings.HomeLayout, error)
}

// HomeDevices resolves local accepted devices without commanding them.
type HomeDevices interface {
	FindDeviceById(string) (*devices.Device, error)
}

// HomeIntegrations reads local validated projections only, never provider refresh/setup work.
type HomeIntegrations interface {
	Instances() ([]integrations.InstanceDTO, error)
}

type homeEnvelope struct {
	Data  *settings.HomeLayout `json:"data"`
	Error *settings.HomeError  `json:"error"`
}

func homeJSON(w http.ResponseWriter, status int, layout *settings.HomeLayout, failure *settings.HomeError) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	if layout != nil {
		w.Header().Set("ETag", layout.ETag)
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(homeEnvelope{Data: layout, Error: failure})
}

func homeError(w http.ResponseWriter, err error) {
	var failure *settings.HomeError
	if !errors.As(err, &failure) {
		failure = &settings.HomeError{Code: "storage_error", Message: "Home could not be loaded or saved. Try again shortly."}
	}
	status := http.StatusBadRequest
	switch failure.Code {
	case "unauthorized":
		status = http.StatusUnauthorized
	case "csrf_rejected":
		status = http.StatusForbidden
	case "precondition_required":
		status = http.StatusPreconditionRequired
	case "revision_conflict", "unsupported_version":
		status = http.StatusConflict
	case "storage_error", "service_unavailable":
		status = http.StatusServiceUnavailable
	}
	homeJSON(w, status, nil, failure)
}

func homeFailure(code, message string) error {
	return &settings.HomeError{Code: code, Message: message}
}

// NewHomeLayoutHandler protects local Home preferences, including when invoked outside Router.
// GET is read-only; PUT validates new source refs then uses the cache's atomic revision fence.
func NewHomeLayoutHandler(layouts HomeLayouts, devices HomeDevices, connections HomeIntegrations) http.Handler {
	allowed := allowedOriginsMap()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		if _, ok := auth.GetUserFromContext(r); !ok {
			homeError(w, homeFailure("unauthorized", "Sign in to use Home."))
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodPut {
			w.Header().Set("Allow", "GET, PUT")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Method == http.MethodPut {
			if cookie, err := r.Cookie(auth.CookieNameSession); err == nil && cookie.Value != "" {
				origin := r.Header.Get("Origin")
				if origin == "" || origin == "null" || !allowed[origin] || r.Header.Get("X-NodeHerder-Request") != "1" {
					homeError(w, homeFailure("csrf_rejected", "Reload Home from the configured NodeHerder address before editing."))
					return
				}
			}
		}
		current, err := layouts.LoadHomeLayout()
		if err != nil {
			homeError(w, err)
			return
		}
		if r.Method == http.MethodGet {
			homeJSON(w, http.StatusOK, &current, nil)
			return
		}
		expected := r.Header.Get("If-Match")
		if expected == "" {
			homeError(w, homeFailure("precondition_required", "Reload Home before saving its layout."))
			return
		}
		if expected != current.ETag {
			homeError(w, homeFailure("revision_conflict", "Home changed in another session. Reload the layout and try again."))
			return
		}
		var proposal settings.HomeLayout
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&proposal); err != nil {
			homeError(w, homeFailure("invalid_input", "Send one valid Home layout."))
			return
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			homeError(w, homeFailure("invalid_input", "Send one Home layout object."))
			return
		}
		if err = proposal.Validate(); err == nil {
			err = validateHomeSources(proposal, current, devices, connections)
		}
		if err != nil {
			homeError(w, err)
			return
		}
		// Recheck the fence under the cache lock: legacy writers may have changed it during reads.
		saved, err := layouts.SaveHomeLayout(proposal, expected)
		if err != nil {
			homeError(w, err)
			return
		}
		homeJSON(w, http.StatusOK, &saved, nil)
	})
}

func validateHomeSources(proposal, current settings.HomeLayout, devices HomeDevices, connections HomeIntegrations) error {
	accepted := map[string]bool{}
	for _, block := range current.Blocks {
		for _, card := range block.Cards {
			accepted[card.SourceKey()] = true
		}
	}
	var instances []integrations.InstanceDTO
	loaded := false
	for _, block := range proposal.Blocks {
		for _, card := range block.Cards {
			if accepted[card.SourceKey()] {
				continue // Accepted missing sources remain explicit, removable placeholders.
			}
			if card.Kind == "device" {
				if devices == nil {
					return homeFailure("invalid_input", "That device value is unavailable.")
				}
				device, err := devices.FindDeviceById(card.DeviceID)
				if err != nil || device == nil || device.Exposes[card.Expose] == nil {
					return homeFailure("invalid_input", "That device value is unavailable.")
				}
				continue
			}
			if !loaded {
				if connections == nil {
					return homeFailure("service_unavailable", "Integrations are unavailable. Existing device areas still work.")
				}
				var err error
				instances, err = connections.Instances()
				if err != nil {
					return homeFailure("service_unavailable", "Integration results could not be read. Existing layout is kept.")
				}
				loaded = true
			}
			if !homeIntegrationExists(card, instances) {
				return homeFailure("invalid_input", "That integration result is no longer available. Reload Home before adding it.")
			}
		}
	}
	return nil
}

func homeIntegrationExists(card settings.HomeCard, instances []integrations.InstanceDTO) bool {
	for _, instance := range instances {
		if instance.ID != card.InstanceID {
			continue
		}
		for _, source := range instance.Sources {
			if source.ResourceID == card.ResourceID && source.Kind == card.Kind && source.ServiceID == card.ServiceID {
				return true
			}
		}
	}
	return false
}
