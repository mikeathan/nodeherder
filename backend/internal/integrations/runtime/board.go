package runtime

import (
	"encoding/json"
	"sort"

	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// BoardEntry references one configured connection. Position in Entries is its display order;
// hiding a widget preserves the connection, credentials and saved resources.
type BoardEntry struct {
	InstanceID string `json:"instance_id"`
	Hidden     bool   `json:"hidden"`
}

// BoardDTO is the shared installation layout. ETag fences membership and preference changes,
// independently of snapshot freshness and the connection's lifecycle revision.
type BoardDTO struct {
	ETag    string       `json:"etag"`
	Entries []BoardEntry `json:"entries"`
}

// readBoardInstances reads private aggregates within the caller's transaction. No provider
// request or read-time resource projection is needed for layout operations.
func readBoardInstances(tx *repository.IntegrationTx) ([]model.Instance, error) {
	instances := []model.Instance{}
	err := tx.Each("instances", func(_ string, data []byte) error {
		var instance model.Instance
		if err := repository.DecodeIntegrationRecord(data, &instance); err != nil {
			return err
		}
		instances = append(instances, instance)
		return nil
	})
	return instances, err
}

func boardDTO(instances []model.Instance) BoardDTO {
	// Hash canonical IDs/revisions, not list-read timing, names or resource revisions.
	sort.Slice(instances, func(a, b int) bool { return instances[a].ID < instances[b].ID })
	type version struct {
		ID       string
		Revision uint64
	}
	versions := make([]version, 0, len(instances))
	for _, instance := range instances {
		var revision uint64
		if instance.Board != nil {
			revision = instance.Board.Revision
		}
		versions = append(versions, version{instance.ID, revision})
	}
	encoded, _ := json.Marshal(versions) // Fixed strings/integers cannot fail JSON encoding.
	out := BoardDTO{ETag: definitions.StrongETag(definitions.ContentHash(encoded)), Entries: make([]BoardEntry, 0, len(instances))}
	sort.SliceStable(instances, func(a, b int) bool {
		left, right := instances[a], instances[b]
		if left.Board != nil && right.Board == nil {
			return true
		}
		if left.Board == nil && right.Board != nil {
			return false
		}
		if left.Board != nil && right.Board != nil && left.Board.Order != right.Board.Order {
			return left.Board.Order < right.Board.Order
		}
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.Before(right.CreatedAt)
		}
		return left.ID < right.ID
	})
	for _, instance := range instances {
		out.Entries = append(out.Entries, BoardEntry{InstanceID: instance.ID, Hidden: instance.Board != nil && instance.Board.Hidden})
	}
	return out
}

// Board reads shared order/visibility without exposing addresses, provider state or secrets.
func (s *Service) Board() (BoardDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out BoardDTO
	if err := s.ready(); err != nil {
		return out, err
	}
	err := s.repo.View(func(tx *repository.IntegrationTx) error {
		instances, err := readBoardInstances(tx)
		if err != nil {
			return err
		}
		out = boardDTO(instances)
		return nil
	})
	return out, err
}

// SaveBoard validates the complete live membership and commits every placement atomically.
// It intentionally leaves lifecycle revisions, generations, timestamps and snapshots untouched.
func (s *Service) SaveBoard(entries []BoardEntry, expected string) (BoardDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out BoardDTO
	if err := s.ready(); err != nil {
		return out, err
	}
	if entries == nil || len(entries) > 50 {
		return out, model.Fail("invalid_input", "Send the complete integration layout (up to 50 entries).")
	}
	err := s.repo.Update(func(tx *repository.IntegrationTx) error {
		instances, err := readBoardInstances(tx)
		if err != nil {
			return err
		}
		if err = checkRevision(expected, boardDTO(instances).ETag); err != nil {
			return err
		}
		if len(entries) != len(instances) {
			return model.Fail("invalid_input", "Include each configured integration exactly once.")
		}
		byID := make(map[string]int, len(instances))
		for index, instance := range instances {
			byID[instance.ID] = index
		}
		seen := make(map[string]bool, len(entries))
		// Validate the whole proposal before writing; the repository transaction also rolls back errors.
		for _, entry := range entries {
			_, exists := byID[entry.InstanceID]
			if !exists || entry.InstanceID == "" || len(entry.InstanceID) > 256 || seen[entry.InstanceID] {
				return model.Fail("invalid_input", "Include each configured integration exactly once.")
			}
			seen[entry.InstanceID] = true
		}
		for order, entry := range entries {
			instance := &instances[byID[entry.InstanceID]]
			previous := instance.Board
			if previous != nil && previous.Order == order && previous.Hidden == entry.Hidden {
				continue
			}
			revision := uint64(1)
			if previous != nil {
				revision = previous.Revision + 1
			}
			instance.Board = &model.BoardPreference{Order: order, Hidden: entry.Hidden, Revision: revision}
			if err = tx.Put("instances", instance.ID, instance); err != nil {
				return err
			}
		}
		out = boardDTO(instances)
		return nil
	})
	return out, err
}
