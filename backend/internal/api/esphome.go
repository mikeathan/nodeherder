package api

// This file exposes the authenticated ESPHome enrollment, discovery and manage
// routes. Handlers validate bounded input and delegate lifecycle work to the
// provider manager; they never connect to a device inline.
import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"node-herder/internal/auth"
	"node-herder/internal/deviceproviders"
	"node-herder/internal/deviceproviders/esphome"
	model "node-herder/models/integrations"
)

const (
	esphomeDevicesPath   = "/api/esphome/devices"
	esphomeDevicePath    = esphomeDevicesPath + "/:id"
	esphomeDiscoveryPath = "/api/esphome/discover"

	esphomeMaxBodyBytes  = 4 * 1024
	esphomeMaxNameBytes  = 256
	esphomeMutationLimit = 10 * time.Second
	nodeHerderRequest    = "X-NodeHerder-Request"
	nodeHerderRequestSet = "1"
)

func esphomePath(path string) bool {
	switch path {
	case esphomeDevicesPath, esphomeDiscoveryPath:
		return true
	}
	return strings.HasPrefix(path, esphomeDevicesPath+"/")
}

type esphomeEnrollments interface {
	List() ([]deviceproviders.Enrollment, error)
	Add(deviceproviders.Enrollment) (deviceproviders.Enrollment, error)
	Remove(context.Context, string) error
	Update(context.Context, string, deviceproviders.EnrollmentUpdate) (deviceproviders.Enrollment, error)
}

type esphomeInput struct {
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	Port          int     `json:"port"`
	EncryptionKey *string `json:"encryption_key"`
}

// RegisterESPHomeRoutes installs authenticated enrollment routes. A nil service
// produces a classified unavailable response rather than disabling other devices.
func RegisterESPHomeRoutes(router *Router, service *deviceproviders.Manager) {
	var enrollments esphomeEnrollments
	if service != nil {
		enrollments = service
	}
	handler := newESPHomeHandler(enrollments)
	router.GET(esphomeDevicesPath, handler)
	router.POST(esphomeDevicesPath, handler)
	router.DELETE(esphomeDevicePath, handler)
	router.PUT(esphomeDevicePath, handler)
	router.POST(esphomeDiscoveryPath, newESPHomeDiscoveryHandler(esphome.Discover))
}

func newESPHomeDiscoveryHandler(scan func(context.Context) ([]esphome.Candidate, error)) http.Handler {
	allowed := allowedOriginsMap()
	active := make(chan struct{}, 1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		if _, ok := auth.GetUserFromContext(r); !ok {
			connectionError(w, model.Fail("unauthorized", "Sign in to discover ESPHome devices."))
			return
		}
		if cookie, err := r.Cookie(auth.CookieNameSession); err == nil && cookie.Value != "" {
			origin := r.Header.Get("Origin")
			if origin == "" || origin == "null" || !allowed[origin] || r.Header.Get(nodeHerderRequest) != nodeHerderRequestSet {
				connectionError(w, model.Fail("csrf_rejected", "Reload the page from the configured NodeHerder address."))
				return
			}
		}
		select {
		case active <- struct{}{}:
			defer func() { <-active }()
		default:
			connectionError(w, model.Fail("busy", "A discovery scan is already running."))
			return
		}
		candidates, err := scan(r.Context())
		if err != nil {
			connectionError(w, model.Fail("service_unavailable", "Discovery is unavailable. You can enter the address manually."))
			return
		}
		if candidates == nil {
			candidates = []esphome.Candidate{}
		}
		connectionJSON(w, http.StatusOK, candidates, nil)
	})
}

func newESPHomeHandler(service esphomeEnrollments) http.Handler {
	allowed := allowedOriginsMap()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		if _, ok := auth.GetUserFromContext(r); !ok {
			connectionError(w, model.Fail("unauthorized", "Sign in to manage ESPHome devices."))
			return
		}
		if service == nil {
			connectionError(w, model.Fail("service_unavailable", "Device provider storage is unavailable."))
			return
		}
		if r.Method == http.MethodGet {
			items, err := service.List()
			if err != nil {
				connectionError(w, model.Fail("storage_unavailable", "Device provider storage is unavailable."))
				return
			}
			if items == nil {
				items = []deviceproviders.Enrollment{}
			}
			connectionJSON(w, http.StatusOK, items, nil)
			return
		}
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if cookie, err := r.Cookie(auth.CookieNameSession); err == nil && cookie.Value != "" {
			origin := r.Header.Get("Origin")
			if origin == "" || origin == "null" || !allowed[origin] || r.Header.Get(nodeHerderRequest) != nodeHerderRequestSet {
				connectionError(w, model.Fail("csrf_rejected", "Reload the page from the configured NodeHerder address."))
				return
			}
		}
		if r.Method == http.MethodDelete {
			id := Param(r, "id")
			if parsed, err := uuid.Parse(id); err != nil || parsed.String() != id {
				connectionError(w, model.Fail("invalid_input", "The device ID is invalid."))
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), esphomeMutationLimit)
			defer cancel()
			err := service.Remove(ctx, id)
			if errors.Is(err, deviceproviders.ErrEnrollmentNotFound) {
				connectionError(w, model.Fail("not_found", "Device not found."))
				return
			}
			if err != nil {
				connectionError(w, model.Fail("storage_unavailable", "The device could not be removed. Try again."))
				return
			}
			connectionJSON(w, http.StatusOK, map[string]bool{"removed": true}, nil)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, esphomeMaxBodyBytes)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var input esphomeInput
		if err := decoder.Decode(&input); err != nil {
			connectionError(w, model.Fail("invalid_input", "The device configuration is invalid."))
			return
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			connectionError(w, model.Fail("invalid_input", "Send one device configuration."))
			return
		}
		if _, err := deviceproviders.NewEndpoint(input.Address, input.Port); err != nil || len(input.Name) > esphomeMaxNameBytes || !utf8.ValidString(input.Name) {
			connectionError(w, model.Fail("invalid_input", "Enter a private IP address, valid port and device name."))
			return
		}
		if input.EncryptionKey != nil && esphome.ValidateEncryptionKey(*input.EncryptionKey) != nil {
			connectionError(w, model.Fail("invalid_input", "The encryption key must be a valid 32-byte ESPHome key."))
			return
		}
		var item deviceproviders.Enrollment
		var err error
		if r.Method == http.MethodPut {
			id := Param(r, "id")
			if parsed, parseErr := uuid.Parse(id); parseErr != nil || parsed.String() != id {
				connectionError(w, model.Fail("invalid_input", "The device ID is invalid."))
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), esphomeMutationLimit)
			defer cancel()
			item, err = service.Update(ctx, id, deviceproviders.EnrollmentUpdate{Name: input.Name, Address: input.Address, Port: input.Port, Credential: input.EncryptionKey})
		} else {
			key := ""
			if input.EncryptionKey != nil {
				key = *input.EncryptionKey
			}
			item, err = service.Add(deviceproviders.Enrollment{Provider: esphome.ProviderName, Name: input.Name, Address: input.Address, Port: input.Port, Credential: key})
		}
		if errors.Is(err, deviceproviders.ErrEnrollmentNotFound) {
			connectionError(w, model.Fail("not_found", "Device not found."))
			return
		}
		if errors.Is(err, deviceproviders.ErrEnrollmentExists) {
			connectionError(w, model.Fail("already_configured", "This ESPHome endpoint is already configured."))
			return
		}
		if err != nil {
			connectionError(w, model.Fail("storage_unavailable", "The device could not be saved."))
			return
		}
		status := http.StatusCreated
		if r.Method == http.MethodPut {
			status = http.StatusOK
		}
		connectionJSON(w, status, item, nil)
	})
}
