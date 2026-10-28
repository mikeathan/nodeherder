package settings

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// HomeLayout is the private, versioned placement aggregate. Values and provider payloads belong
// to their existing sources; only references and presentation choices are persisted here.
type HomeLayout struct {
	Version  int         `json:"version"`
	Revision uint64      `json:"revision"`
	ETag     string      `json:"etag"`
	Blocks   []HomeBlock `json:"blocks"`
}

// HomeBlock is an ordered group, or one standalone card. ID survives group renames; Name is
// presentation only. Card blocks have an empty name and exactly one member.
type HomeBlock struct {
	ID    string     `json:"id"`
	Kind  string     `json:"kind"`
	Name  string     `json:"name"`
	Cards []HomeCard `json:"cards"`
}

// HomeCard identifies one placement, independently of its source's current value or list order.
// Panel controls wall-display visibility only. Device and integration references are disjoint.
type HomeCard struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Panel      bool   `json:"panel"`
	DeviceID   string `json:"device_id,omitempty"`
	Expose     string `json:"expose,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
	ServiceID  string `json:"service_id,omitempty"`
}

// HomeError is a sanitized layout failure suitable for the authenticated API. Code describes
// input, concurrency or persistence failure without exposing stored labels or provider data.
type HomeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns the safe explanation, never the submitted payload.
func (e *HomeError) Error() string { return e.Message }

func invalidHome() error {
	return &HomeError{Code: "invalid_input", Message: "The Home layout is invalid or exceeds its limits."}
}

func homeIdentifier(value string) bool {
	return value != "" && len(value) <= 256 && utf8.ValidString(value) && !strings.ContainsFunc(value, unicode.IsControl)
}

// Validate checks bounded structure and reference kinds without performing source/provider I/O.
// Missing accepted sources remain representable; API admission separately validates new sources.
func (layout HomeLayout) Validate() error {
	if layout.Version != 1 || layout.Blocks == nil || len(layout.Blocks) > 50 {
		return invalidHome()
	}
	blocks, cards, names := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, block := range layout.Blocks {
		if !homeIdentifier(block.ID) || blocks[block.ID] || block.Cards == nil {
			return invalidHome()
		}
		blocks[block.ID] = true
		switch block.Kind {
		case "group":
			name := strings.TrimSpace(block.Name)
			key := strings.ToLower(name)
			if name == "" || name != block.Name || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 60 || strings.ContainsFunc(name, unicode.IsControl) || names[key] {
				return invalidHome()
			}
			names[key] = true
		case "card":
			if block.Name != "" || len(block.Cards) != 1 {
				return invalidHome()
			}
		default:
			return invalidHome()
		}
		for _, card := range block.Cards {
			if !homeIdentifier(card.ID) || cards[card.ID] || len(cards) >= 500 {
				return invalidHome()
			}
			cards[card.ID] = true
			device := homeIdentifier(card.DeviceID) && homeIdentifier(card.Expose)
			integration := homeIdentifier(card.InstanceID) && homeIdentifier(card.ResourceID)
			switch card.Kind {
			case "device":
				if !device || card.InstanceID != "" || card.ResourceID != "" || card.ServiceID != "" {
					return invalidHome()
				}
			case "waste_service", "waste_summary":
				if !integration || card.DeviceID != "" || card.Expose != "" || (card.Kind == "waste_service" && !homeIdentifier(card.ServiceID)) || (card.Kind == "waste_summary" && card.ServiceID != "") {
					return invalidHome()
				}
			default:
				return invalidHome()
			}
		}
	}
	return nil
}

// Clone returns a completely owned placement snapshot, including empty slices. Neither callers
// nor JSON decoders may mutate the cache's accepted layout through a returned value.
func (layout HomeLayout) Clone() HomeLayout {
	layout.Blocks = slices.Clone(layout.Blocks)
	for index := range layout.Blocks {
		layout.Blocks[index].Cards = slices.Clone(layout.Blocks[index].Cards)
	}
	return layout
}

// HomeETag is a strong revision/content fence. Provider freshness is absent from this model,
// so a local resource read cannot cause a layout conflict or cancel an integration operation.
func HomeETag(layout HomeLayout) string {
	layout.ETag = ""
	encoded, _ := json.Marshal(layout) // This model contains only JSON-safe scalar fields/slices.
	hash := sha256.Sum256(encoded)
	return `"` + hex.EncodeToString(hash[:]) + `"`
}

// SourceKey identifies the source, not the placement. Moving or duplicating a placement never
// changes where its values come from; JSON tuple encoding prevents delimiter collisions.
func (card HomeCard) SourceKey() string {
	encoded, _ := json.Marshal([]string{card.Kind, card.DeviceID, card.Expose, card.InstanceID, card.ResourceID, card.ServiceID})
	return string(encoded)
}

func legacyHomeID(parts ...string) string {
	encoded, _ := json.Marshal(parts)
	hash := sha256.Sum256(encoded)
	return "legacy-" + hex.EncodeToString(hash[:16])
}

// HomeFromGroups derives a deterministic v1 layout without mutating legacy data or writing.
// Explicit area order precedes unordered areas; expose order is retained and device IDs are
// canonicalized like the legacy Go JSON map. Existing devices remain opted into panel mode.
func HomeFromGroups(groups map[string]*DashboardGroup) HomeLayout {
	ordered := make([]*DashboardGroup, 0, len(groups))
	for _, group := range groups {
		if group != nil {
			ordered = append(ordered, group)
		}
	}
	sort.SliceStable(ordered, func(a, b int) bool {
		left, right := ordered[a], ordered[b]
		if left.Order != nil && right.Order != nil && *left.Order != *right.Order {
			return *left.Order < *right.Order
		}
		if (left.Order == nil) != (right.Order == nil) {
			return left.Order != nil
		}
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	})
	layout := HomeLayout{Version: 1, Blocks: make([]HomeBlock, 0, len(ordered))}
	for _, group := range ordered {
		block := HomeBlock{ID: legacyHomeID("group", group.Name), Kind: "group", Name: group.Name, Cards: []HomeCard{}}
		block.Cards = legacyDeviceCards(block.ID, group.DeviceGroup)
		layout.Blocks = append(layout.Blocks, block)
	}
	return layout
}

func legacyDeviceCards(groupID string, groups map[string]*DeviceGroup) []HomeCard {
	ids := make([]string, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	cards := []HomeCard{}
	seen := map[string]bool{}
	for _, id := range ids {
		group := groups[id]
		if group == nil {
			continue
		}
		for _, expose := range group.Exposes {
			card := HomeCard{Kind: "device", DeviceID: group.DeviceId, Expose: expose, Panel: true}
			if seen[card.SourceKey()] {
				continue
			}
			seen[card.SourceKey()] = true
			card.ID = legacyHomeID("card", groupID, card.DeviceID, card.Expose)
			cards = append(cards, card)
		}
	}
	return cards
}

// DeviceGroups returns the device-only compatibility projection. Integration references and
// standalone cards never enter ordinary AppConfig/WS exports. The layout stays authoritative.
func (layout HomeLayout) DeviceGroups() map[string]*DashboardGroup {
	groups := map[string]*DashboardGroup{}
	for _, block := range layout.Blocks {
		if block.Kind != "group" {
			continue
		}
		order := len(groups)
		group := &DashboardGroup{Name: block.Name, Order: &order, DeviceGroup: map[string]*DeviceGroup{}}
		for _, card := range block.Cards {
			if card.Kind != "device" {
				continue
			}
			entry := group.DeviceGroup[card.DeviceID]
			if entry == nil {
				entry = &DeviceGroup{DeviceId: card.DeviceID, Exposes: []string{}}
				group.DeviceGroup[card.DeviceID] = entry
			}
			if !slices.Contains(entry.Exposes, card.Expose) {
				entry.Exposes = append(entry.Exposes, card.Expose)
			}
		}
		groups[group.Name] = group
	}
	return groups
}

func homeFailure(code, message string) error { return &HomeError{Code: code, Message: message} }

func homeSchemaError(version int) error {
	return homeFailure("unsupported_version", fmt.Sprintf("Home layout version %d is unsupported. Restore a compatible backup or update NodeHerder.", version))
}
