// Package capabilities validates provider results into stable application contracts.
package capabilities

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

// WasteContext carries evidence not necessarily present below a response_root selector.
// The identity map is copied; persist the returned map and snapshot in the same transaction.
type WasteContext struct {
	NewIdentity                        func(kind string) (string, error)
	InstanceID, ResourceID, ProviderID string
	ProviderVersion                    string
	ProfileRevision, BindingRevision   uint64
	SelectedProperty                   string
	Now                                time.Time
	Previous                           *model.WasteSchedule
	ServiceIDs                         map[string]string
}
type WasteResult struct {
	Snapshot   model.WasteSchedule
	ServiceIDs map[string]string
}

func read(root any, p string) (any, bool) {
	if p == "" {
		return nil, false
	}
	return operations.Pointer(root, p)
}
func text(root any, p string) (string, bool) {
	v, ok := read(root, p)
	s, is := v.(string)
	return s, ok && is
}
func nullable(root any, p string) (*string, error) {
	v, ok := read(root, p)
	if !ok || v == nil {
		return nil, nil
	}
	s, ok := v.(string)
	if !ok || len(s) > 512 {
		return nil, invalid("A provider text field is invalid.")
	}
	return &s, nil
}
func invalid(message string) error { return model.Fail("invalid_response", message) }
func enum(value string, allowed ...string) string {
	for _, v := range allowed {
		if value == v {
			return value
		}
	}
	return "unknown"
}
func localDate(s string) bool {
	d, e := time.Parse("2006-01-02", s)
	return e == nil && d.Format("2006-01-02") == s
}
func timestamp(root any, p string) (*time.Time, error) {
	s, e := nullable(root, p)
	if e != nil || s == nil {
		return nil, e
	}
	t, e := time.Parse(time.RFC3339Nano, *s)
	if e != nil {
		return nil, invalid("A provider timestamp is invalid.")
	}
	t = t.UTC()
	return &t, nil
}

// NormalizeWaste rejects non-exact/provisional results and never invents coverage or empty evidence.
func NormalizeWaste(root any, m model.Mapping, c WasteContext) (WasteResult, error) {
	var out WasteResult
	property, ok := text(root, m.Metadata.PropertyID)
	if !ok || property != c.SelectedProperty {
		return out, invalid("The result does not match the selected address. Set up the connection again.")
	}
	match, _ := text(root, m.Metadata.AddressMatch)
	evidence, _ := text(root, m.Metadata.EvidenceGranularity)
	provisional, has := read(root, m.Metadata.Provisional)
	if has {
		b, ok := provisional.(bool)
		if !ok {
			return out, invalid("Invalid provisional flag.")
		}
		if b {
			return out, model.Fail("provisional_result", "The provider returned an approximate result. No collection dates were saved.")
		}
	}
	if match != "exact" || evidence != "property" {
		return out, model.Fail("exact_address_required", "This connection needs dates for the selected property.")
	}
	retrieved, e := timestamp(root, m.Metadata.SourceRetrievedAt)
	if e != nil {
		return out, e
	}
	expiry, e := timestamp(root, m.Metadata.RefreshDueAt)
	if e != nil {
		return out, e
	}
	if retrieved != nil && retrieved.After(c.Now.Add(5*time.Minute)) {
		return out, invalid("Provider verification time is in the future.")
	}
	if retrieved != nil && expiry != nil && expiry.Before(*retrieved) {
		return out, invalid("Provider freshness expires before verification.")
	}
	confidence, _ := text(root, m.Metadata.DateConfidence)
	if len(confidence) > 512 {
		return out, invalid("The provider confidence label is too long.")
	}
	completeness, _ := text(root, m.Metadata.DateCompleteness)
	quality := model.Quality{AddressMatch: "exact", EvidenceGranularity: "property", DateConfidence: enum(confidence, "published_calendar", "council_projection", "next_collection_only")}
	if quality.DateConfidence == "unknown" && confidence != "" {
		quality.SourceConfidenceLabel = &confidence
	}
	newIdentity := c.NewIdentity
	if newIdentity == nil {
		newIdentity = func(string) (string, error) { id, e := uuid.NewRandom(); return id.String(), e }
	}
	resourceID := c.ResourceID
	if resourceID == "" {
		var e error
		resourceID, e = newIdentity("resource")
		if e != nil {
			return out, e
		}
	}
	revision := uint64(1)
	if c.Previous != nil {
		revision = c.Previous.Revision + 1
	}
	if c.ProviderVersion == "" {
		c.ProviderVersion = "unversioned"
	}
	snapshot := model.WasteSchedule{ContractVersion: model.WasteCapability, InstanceID: c.InstanceID, ResourceID: resourceID, Revision: revision, Kind: "exact", Timezone: "Europe/London", CapturedAt: c.Now.UTC(), LastCheckedAt: c.Now.UTC(), SourceRetrievedAt: retrieved, FreshUntil: expiry, Quality: quality, Provenance: model.Provenance{ProviderID: c.ProviderID, ProviderVersion: c.ProviderVersion, ProfileRevision: c.ProfileRevision, BindingRevision: c.BindingRevision}, RepresentativeConsent: "not_required", Notices: []string{}, Services: []model.WasteService{}}
	if u, ok := text(root, m.Metadata.SourceURL); ok {
		parsed, err := url.Parse(u)
		if err == nil && parsed.Scheme == "https" && parsed.Hostname() != "" && parsed.User == nil && len(u) <= 2048 {
			snapshot.Provenance.SourceURL = &u
		} else {
			snapshot.Notices = append(snapshot.Notices, "The provider supplied an unsupported source link.")
		}
	}
	if notes, ok := read(root, m.Metadata.Notes); ok && notes != nil {
		switch n := notes.(type) {
		case string:
			if len(n) > 2048 {
				return out, invalid("Provider notes are too long.")
			}
			snapshot.Notices = append(snapshot.Notices, n)
		case []any:
			if len(n) > 32 {
				return out, invalid("Too many provider notes.")
			}
			for _, item := range n {
				s, ok := item.(string)
				if !ok || len(s) > 2048 {
					return out, invalid("Provider notes are invalid.")
				}
				snapshot.Notices = append(snapshot.Notices, s)
			}
		default:
			return out, invalid("Provider notes are invalid.")
		}
	}
	if len(snapshot.Notices) > 32 {
		return out, invalid("Too many provider notices.")
	}
	items, ok := read(root, m.Collections.Items)
	list, is := items.([]any)
	if !ok || !is || len(list) > 64 || len(list) == 0 {
		return out, invalid("No verified collection services were returned. Previous dates were kept.")
	}
	ids := map[string]string{}
	usedIDs := map[string]bool{}
	for key, id := range c.ServiceIDs {
		if _, e := uuid.Parse(id); e != nil || usedIDs[id] {
			return out, model.Fail("reconciliation_required", "Saved service identities need review.")
		}
		ids[key] = id
		usedIDs[id] = true
	}
	knownServices := 0
	seen := map[string]bool{}
	present := map[string]bool{}
	for _, item := range list {
		f := m.Collections.Fields
		name, ok := text(item, f.Name)
		if !ok || strings.TrimSpace(name) == "" || len(name) > 512 {
			return out, invalid("A collection name is missing or invalid.")
		}
		rawType, e := nullable(item, f.WasteType)
		if e != nil {
			return out, e
		}
		container, e := nullable(item, f.Container)
		if e != nil {
			return out, e
		}
		colour, e := nullable(item, f.Colour)
		if e != nil {
			return out, e
		}
		sourceKey, e := nullable(item, f.SourceServiceKey)
		if e != nil {
			return out, e
		}
		key := ""
		if sourceKey != nil && *sourceKey != "" {
			key = "source:" + *sourceKey
		} else {
			tuple, _ := json.Marshal([]*string{&name, rawType, container})
			sum := sha256.Sum256(tuple)
			key = "descriptor:v1:" + hex.EncodeToString(sum[:])
		}
		if seen[key] {
			return out, model.Fail("reconciliation_required", "The provider returned ambiguous collection services.")
		}
		seen[key] = true
		id := ids[key]
		if id == "" {
			for attempt := 0; attempt < 3; attempt++ {
				candidate, e := newIdentity("service")
				if e != nil {
					return out, e
				}
				if !usedIDs[candidate] {
					id = candidate
					break
				}
			}
			if id == "" {
				return out, invalid("Could not allocate a service identity.")
			}
			ids[key] = id
			usedIDs[id] = true
		}
		present[id] = true
		service := model.WasteService{ServiceID: id, SourceServiceKey: sourceKey, Name: name, WasteType: "unknown", SourceWasteType: rawType, BinColour: colour, Container: container, Subscription: "unknown", DateState: "unavailable", Dates: []string{}, Coverage: model.Coverage{State: "unknown"}}
		if rawType != nil {
			service.WasteType = enum(*rawType, "refuse", "recycling", "paper_card", "glass", "garden", "food", "mixed", "other")
		}
		if value, ok := read(item, f.SubscriptionRequired); ok && value != nil {
			b, ok := value.(bool)
			if !ok {
				return out, invalid("Invalid subscription flag.")
			}
			if b {
				service.Subscription = "required"
			} else {
				service.Subscription = "not_required"
			}
		}
		value, ok := read(item, f.Dates)
		dates, is := value.([]any)
		if !ok || !is || len(dates) > 730 {
			return out, invalid("Collection dates are missing or invalid.")
		}
		unique := map[string]bool{}
		for _, d := range dates {
			s, ok := d.(string)
			if !ok || !localDate(s) {
				return out, invalid("The provider returned an invalid collection date.")
			}
			unique[s] = true
		}
		for d := range unique {
			service.Dates = append(service.Dates, d)
		}
		sort.Strings(service.Dates)
		if len(service.Dates) > 0 {
			knownServices++
			service.DateState = "known"
			switch completeness {
			case "limited_horizon":
				service.Coverage.State = "limited"
			case "next_only":
				service.Coverage.State = "next_only"
			}
			if quality.DateConfidence == "next_collection_only" {
				service.Coverage.State = "next_only"
			}
		} else {
			snapshot.Kind = "partial"
		}
		if value, ok := read(item, f.DatesComplete); ok && value != nil {
			complete, ok := value.(bool)
			if !ok {
				return out, invalid("Invalid date completeness flag.")
			}
			if !complete {
				snapshot.Kind = "partial"
			}
		}
		snapshot.Services = append(snapshot.Services, service)
	}
	if knownServices == 0 {
		return out, model.Fail("no_collection_dates", "The provider supplied no verified collection dates. Previous saved dates were kept.")
	}
	if c.Previous != nil {
		for _, old := range c.Previous.Services {
			if !present[old.ServiceID] {
				old.DateState = "unavailable"
				old.Dates = []string{}
				old.Coverage = model.Coverage{State: "unknown"}
				snapshot.Services = append(snapshot.Services, old)
				snapshot.Kind = "partial"
			}
		}
	}
	if len(snapshot.Services) > 64 {
		return out, model.Fail("reconciliation_required", "Too many changed collection services. Review the provider mapping before replacing saved data.")
	}
	snapshot = snapshot.Project(c.Now)
	out.Snapshot = snapshot
	out.ServiceIDs = ids
	return out, nil
}
