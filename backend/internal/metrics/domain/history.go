package domain

import (
	"context"
	"fmt"
	"node-herder/utils"
	"time"
)

// ScanGuard lets bounded consumers stop a repository scan without changing legacy collectors.
type ScanGuard interface{ CheckScan() error }

// HistoryBudget is owned by one synchronous query, shared by its collectors.
type HistoryBudget struct {
	Context   context.Context
	Remaining int
}

func (b *HistoryBudget) CheckScan() error {
	if err := b.Context.Err(); err != nil {
		return err
	}
	b.Remaining--
	if b.Remaining < 0 {
		return fmt.Errorf("history scan budget exceeded; narrow the range")
	}
	return nil
}

type HistoryPoint struct {
	Timestamp int64 `json:"timestamp"`
	Value     any   `json:"value"`
}

// HistoryResult retains repeated reports; it does not merge event values.
type HistoryResult struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	From    int64          `json:"from"`
	To      int64          `json:"to"`
	Data    []HistoryPoint `json:"data"`
	Partial bool           `json:"partial"`
	Limit   int            `json:"-"`
	Budget  *HistoryBudget `json:"-"`
}

func (h *HistoryResult) GetType() string  { return h.Type }
func (h *HistoryResult) Size() int        { return len(h.Data) }
func (h *HistoryResult) Flush()           {}
func (h *HistoryResult) CheckScan() error { return h.Budget.CheckScan() }

func (h *HistoryResult) Collect(at time.Time, raw []byte) error {
	if len(h.Data) >= h.Limit {
		h.Partial = true
		return nil
	}
	value, err := decodeHistoryValue(h.Type, raw)
	if err != nil {
		return err
	}
	h.Data = append(h.Data, HistoryPoint{Timestamp: at.UnixMilli(), Value: value})
	return nil
}

func decodeHistoryValue(kind string, raw []byte) (any, error) {
	switch kind {
	case "numeric":
		var value float32
		err := utils.DecodeGobValue(raw, &value)
		return value, err
	case "binary":
		var value bool
		err := utils.DecodeGobValue(raw, &value)
		return value, err
	case "enum":
		var value string
		err := utils.DecodeGobValue(raw, &value)
		return value, err
	default:
		return nil, fmt.Errorf("unsupported history type %q", kind)
	}
}
