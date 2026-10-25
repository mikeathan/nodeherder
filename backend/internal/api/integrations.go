package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"node-herder/internal/auth"
	connections "node-herder/internal/integrations"
	model "node-herder/models/integrations"
)

type connectionEnvelope struct {
	Data  any            `json:"data"`
	Error *model.Failure `json:"error"`
}

func connectionError(w http.ResponseWriter, err error) {
	var failure *model.Failure
	if !errors.As(err, &failure) {
		failure = model.Fail("storage_error", "The change could not be saved. Check the Connections storage and try again.")
	}
	status := http.StatusBadRequest
	switch failure.Code {
	case "unauthorized":
		status = 401
	case "csrf_rejected":
		status = 403
	case "not_found":
		status = 404
	case "setup_expired", "operation_expired":
		status = 410
	case "revision_conflict", "already_configured", "definition_in_use", "destination_review_required":
		status = 409
	case "precondition_required":
		status = 428
	case "busy", "retry_later", "rate_limited":
		status = 429
	case "service_unavailable", "key_unavailable", "storage_unavailable", "config_directory_unwritable", "saved_not_applied", "storage_error":
		status = 503
	}
	if failure.RetryAt != nil {
		w.Header().Set("Retry-After", failure.RetryAt.UTC().Format(http.TimeFormat))
	}
	connectionJSON(w, status, nil, failure)
}
func connectionJSON(w http.ResponseWriter, status int, data any, failure *model.Failure) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(connectionEnvelope{Data: data, Error: failure})
}

type connectionInput struct {
	InstanceID        string                         `json:"instance_id"`
	Mode              string                         `json:"mode"`
	Sample            json.RawMessage                `json:"sample"`
	ProviderKey       string                         `json:"provider_key"`
	TargetInstanceID  string                         `json:"target_instance_id"`
	TargetETag        string                         `json:"target_etag"`
	Action            string                         `json:"action"`
	Postcode          string                         `json:"postcode"`
	PropertyID        string                         `json:"property_id"`
	Name              string                         `json:"name"`
	Enabled           bool                           `json:"enabled"`
	Value             string                         `json:"value"`
	YAML              string                         `json:"yaml"`
	Document          json.RawMessage                `json:"document"`
	DestinationReview *connections.DestinationReview `json:"destination_review"`
}

func decodeConnectionInput(r *http.Request) (connectionInput, error) {
	var input connectionInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&input); e != nil && e != io.EOF {
		return input, model.Fail("invalid_input", "The request body is invalid.")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return input, model.Fail("invalid_input", "Use one JSON request object.")
	}
	return input, nil
}

// RegisterIntegrationRoutes installs only protected routes, without changing device/MCP payloads.
func RegisterIntegrationRoutes(router *Router, service *connections.Service) {
	routes := []struct{ method, path, action string }{
		{"GET", "/catalog", "catalog"}, {"POST", "/definitions/reload", "reload_all"}, {"GET", "/instances", "instances"}, {"GET", "/instances/:id", "instance"}, {"GET", "/instances/:id/resources", "resources"}, {"PUT", "/instances/:id", "update"}, {"DELETE", "/instances/:id", "delete"}, {"POST", "/instances/:id/refresh", "refresh"}, {"POST", "/instances/:id/check-status", "check_instance"},
		{"POST", "/setup", "setup_create"}, {"GET", "/setup/:id", "setup"}, {"POST", "/setup/:id/step", "step"}, {"DELETE", "/setup/:id", "cancel"},
		{"GET", "/operations/:id", "operation"}, {"POST", "/operations/:id/check-status", "check"},
		{"PUT", "/setup/:id/credentials/:slot", "setup_credential"}, {"PUT", "/instances/:id/credentials/:slot", "credential"}, {"DELETE", "/instances/:id/credentials/:slot", "credential_delete"},
		{"GET", "/definitions/:type/:provider", "definition"}, {"PUT", "/definitions/:type/:provider", "definition_save"}, {"DELETE", "/definitions/:type/:provider", "definition_delete"}, {"POST", "/definitions/:type/:provider/validate", "validate"}, {"POST", "/definitions/:type/:provider/test", "test"}, {"POST", "/definitions/:type/:provider/reload", "reload"}, {"POST", "/definitions/:type/:provider/restore", "restore"}, {"POST", "/definitions/:type/:provider/shipped-reviewed", "shipped_reviewed"},
	}
	for _, route := range routes {
		handler := NewIntegrationHandler(service, route.action)
		path := "/api/integrations" + route.path
		switch route.method {
		case "GET":
			router.GET(path, handler)
		case "POST":
			router.POST(path, handler)
		case "PUT":
			router.PUT(path, handler)
		case "DELETE":
			router.DELETE(path, handler)
		}
	}
}

// NewIntegrationHandler checks the established auth context even when called outside the router.
func NewIntegrationHandler(service *connections.Service, action string) http.Handler {
	allowed := allowedOriginsMap()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		if _, ok := auth.GetUserFromContext(r); !ok {
			connectionError(w, model.Fail("unauthorized", "Sign in to use Connections."))
			return
		}
		mutation := r.Method != "GET"
		if mutation {
			if cookie, e := r.Cookie(auth.CookieNameSession); e == nil && cookie.Value != "" {
				origin := r.Header.Get("Origin")
				if origin == "" || origin == "null" || !allowed[origin] || r.Header.Get("X-NodeHerder-Request") != "1" {
					connectionError(w, model.Fail("csrf_rejected", "Reload the page from the configured NodeHerder address before changing Connections."))
					return
				}
			}
		}
		var input connectionInput
		var err error
		if mutation {
			input, err = decodeConnectionInput(r)
			if err != nil {
				connectionError(w, err)
				return
			}
		}
		id := Param(r, "id")
		key := Param(r, "type") + "/" + Param(r, "provider")
		expected := r.Header.Get("If-Match")
		status := 200
		var data any
		switch action {
		case "reload_all":
			data, err = service.ReloadDefinitions()
		case "catalog":
			data, err = service.Catalog()
		case "instances":
			data, err = service.Instances()
		case "instance":
			data, err = service.Instance(id)
		case "resources":
			var i connections.InstanceDTO
			i, err = service.Instance(id)
			data = i.Resources
		case "update":
			data, err = service.UpdateInstance(id, input.Name, input.Enabled, expected)
		case "delete":
			err = service.DeleteInstance(id, expected)
			data = map[string]bool{"deleted": err == nil}
		case "refresh":
			data, err = service.Refresh(id, expected)
			status = 202
		case "check_instance":
			var i connections.InstanceDTO
			i, err = service.Instance(id)
			if err == nil {
				if i.OperationID == "" {
					err = model.Fail("not_found", "No pending lookup exists for this connection.")
				} else {
					data, err = service.CheckOperation(i.OperationID, expected)
				}
			}
			status = 202
		case "setup_create":
			data, err = service.CreateSetup(input.ProviderKey, input.TargetInstanceID, input.TargetETag)
			status = 201
		case "setup":
			data, err = service.Setup(id)
		case "step":
			switch input.Action {
			case "search":
				data, err = service.SearchSetup(id, input.Postcode, expected)
			case "save":
				data, err = service.SaveSetup(id, input.PropertyID, input.Name, expected)
			default:
				err = model.Fail("invalid_input", "Choose search or save.")
			}
			status = 202
		case "cancel":
			err = service.CancelSetup(id, expected)
			data = map[string]bool{"cancelled": err == nil}
		case "operation":
			data, err = service.Operation(id)
		case "check":
			data, err = service.CheckOperation(id, expected)
			status = 202
		case "setup_credential":
			data, err = service.SetCredential("setup", id, Param(r, "slot"), input.Value, expected, false)
		case "credential", "credential_delete":
			data, err = service.SetCredential("instances", id, Param(r, "slot"), input.Value, expected, action == "credential_delete")
		case "definition":
			data, err = service.Definition(key)
		case "definition_save":
			create := r.Header.Get("If-None-Match") == "*"
			if expected == "" && !create {
				err = model.Fail("precondition_required", "Reload the provider before saving.")
			} else {
				data, err = service.SaveDefinition(key, input.YAML, expected, create, input.DestinationReview)
			}
		case "definition_delete":
			err = service.DeleteDefinition(key, expected)
			data = map[string]bool{"deleted": err == nil}
		case "validate":
			if input.YAML != "" && input.Document != nil {
				err = model.Fail("invalid_input", "Send either YAML or a form document.")
			} else {
				data, err = service.ValidateDefinition(key, input.YAML, input.Document)
			}
		case "test":
			data, err = service.TestDefinition(key, input.YAML, input.InstanceID, input.Mode, input.Sample, input.DestinationReview)
			if input.Mode == "live" {
				status = 202
			}
		case "reload":
			data, err = service.ReloadDefinition(key, expected, input.DestinationReview)
		case "shipped_reviewed":
			data, err = service.AcknowledgeShipped(key, expected)
		case "restore":
			data, err = service.RestoreDefinition(key, expected, input.DestinationReview)
		default:
			err = model.Fail("not_found", "Connection operation not found.")
		}
		if err != nil {
			connectionError(w, err)
			return
		}
		switch dto := data.(type) {
		case connections.SetupDTO:
			w.Header().Set("ETag", dto.ETag)
			if dto.OperationID != "" {
				w.Header().Set("Location", "/api/integrations/operations/"+dto.OperationID)
			}
		case connections.InstanceDTO:
			w.Header().Set("ETag", dto.ETag)
		case connections.DefinitionDTO:
			w.Header().Set("ETag", dto.ETag)
		case connections.OperationDTO:
			if !dto.RetryAt.Before(time.Now()) {
				w.Header().Set("Retry-After", dto.RetryAt.Format(http.TimeFormat))
			}
			w.Header().Set("Location", "/api/integrations/operations/"+dto.ID)
		}
		connectionJSON(w, status, data, nil)
	})
}

// integrationPath is intentionally narrower than all API traffic for new browser headers.
func integrationPath(path string) bool { return strings.HasPrefix(path, "/api/integrations/") }
