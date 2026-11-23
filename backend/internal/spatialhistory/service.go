package spatialhistory

import (
	"sort"
	"time"
)

// Service provides read models over the recorder store.
type Service struct{ repo Repository }

// NewService wraps a repository for API and UI queries.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Sessions groups chunks by session id and summarises them.
func (s *Service) Sessions(device string, from, to time.Time) ([]Session, error) {
	// A chunk that started before the requested range can still contain points inside it (an
	// ongoing session); look back one chunk interval and filter sessions by overlap.
	chunks, err := s.repo.Chunks(device, from.Add(-sourceIdleClose), to)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*Session)
	slots := make(map[string]map[int]bool)
	// The activity summary only needs each session's recorded timestamps in order, not its points:
	// a 7-day query can span hundreds of sessions, and retaining every Point would duplicate the
	// whole store in memory for the duration of one request.
	sampleTimes := make(map[string][]int64)
	for _, chunk := range chunks {
		entry := byID[chunk.Session]
		if entry == nil {
			// Slots starts non-nil: the wire contract is `number[]`, and a nil slice marshals as
			// `null`, which the browser's slot label code does not expect.
			entry = &Session{ID: chunk.Session, From: chunk.From, To: chunk.To, Slots: make([]int, 0)}
			byID[chunk.Session] = entry
			slots[chunk.Session] = make(map[int]bool)
		}
		if chunk.From < entry.From {
			entry.From = chunk.From
		}
		if chunk.To > entry.To {
			entry.To = chunk.To
		}
		entry.Capped = entry.Capped || chunk.Capped
		entry.To = max(entry.To, chunk.To)
		for _, point := range chunk.Points {
			if point.At >= from.UnixMilli() && point.At <= to.UnixMilli() {
				entry.Points++
				slots[chunk.Session][point.Slot] = true
				// Collected in the same pass that counts points: no second read of the chunks.
				sampleTimes[chunk.Session] = append(sampleTimes[chunk.Session], point.At)
			}
		}
	}
	sessions := make([]Session, 0, len(byID))
	for id, entry := range byID {
		if entry.To < from.UnixMilli() {
			continue
		}
		for slot := range slots[id] {
			entry.Slots = append(entry.Slots, slot)
		}
		sort.Ints(entry.Slots)
		times := sampleTimes[id]
		// Chunk keys sort by start time, so concatenation is already ascending; sorting keeps the
		// drawn activity tied to real time order even if chunks overlap.
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
		entry.Samples = samplesFrom(times)
		sessions = append(sessions, *entry)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].From < sessions[j].From })
	return sessions, nil
}

// Points returns the ordered points for one session inside the requested range.
func (s *Service) Points(device, session string, from, to time.Time) ([]Point, error) {
	// Same lookback as Sessions: a chunk that started before the window can still hold points
	// inside it. Without it the list offers a recording whose detail loads empty, and the player
	// reports "No recorded positions in this session" for a recording it just listed with points.
	chunks, err := s.repo.Chunks(device, from.Add(-sourceIdleClose), to)
	if err != nil {
		return nil, err
	}
	var points []Point
	for _, chunk := range chunks {
		if chunk.Session != session {
			continue
		}
		for _, point := range chunk.Points {
			if point.At >= from.UnixMilli() && point.At <= to.UnixMilli() {
				points = append(points, point)
			}
		}
	}
	sort.Slice(points, func(i, j int) bool { return points[i].At < points[j].At })
	return points, nil
}

// Heatmap aggregates points into square cells. It derives cells on demand; no second copy exists.
func (s *Service) Heatmap(device string, from, to time.Time, cellMM int) ([]Cell, error) {
	if cellMM < 50 {
		cellMM = 250
	}
	if cellMM > 1000 {
		cellMM = 1000
	}
	// Same lookback as Sessions/Points: a chunk that started before the window can still hold
	// points inside it. Without it the heatmap disagrees with the recordings and points it is
	// drawn beside.
	chunks, err := s.repo.Chunks(device, from.Add(-sourceIdleClose), to)
	if err != nil {
		return nil, err
	}
	counts := make(map[[2]int32]int)
	for _, chunk := range chunks {
		for _, point := range chunk.Points {
			if point.At < from.UnixMilli() || point.At > to.UnixMilli() {
				continue
			}
			key := [2]int32{quantizeCell(point.X, cellMM), quantizeCell(point.Y, cellMM)}
			counts[key]++
		}
	}
	cells := make([]Cell, 0, len(counts))
	for key, count := range counts {
		cells = append(cells, Cell{X: key[0], Y: key[1], Count: count})
	}
	sort.Slice(cells, func(i, j int) bool {
		if cells[i].X == cells[j].X {
			return cells[i].Y < cells[j].Y
		}
		return cells[i].X < cells[j].X
	})
	return cells, nil
}

func quantizeCell(value int32, cell int) int32 {
	size := int32(cell)
	// Floor division, including negative X.
	quotient := value / size
	if value%size < 0 {
		quotient--
	}
	return quotient * size
}
