// Package summaries stores one row per identity per day: the evidence classification reads.
//
// It is deliberately narrow. It stores aggregates and nothing else — no observations, no packet
// contents, no payloads — so retaining a month of it cannot reconstruct what any device
// transmitted. That is what makes retention defensible, and it is asserted by the row shape rather
// than by review.
//
// The class is never stored. Classification is recomputed from these rows, so changing a rule
// reclassifies history immediately and a device whose behaviour changed is judged on the days that
// actually happened. Storing a verdict would freeze it.
//
// It follows internal/spatialhistory on purpose: own bbolt file, own Prune, worker-serialised
// writes, bounded queue that drops rather than blocking the observation pipeline. Two stores with
// the same lifecycle should not be built two ways.
package summaries

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	model "node-herder/models/observations"
	"node-herder/utils"
	"node-herder/utils/storage"

	bolt "go.etcd.io/bbolt"
)

const (
	filename = "observation-summaries.db"

	// Bucket names. Summaries and transitions live in separate buckets but share one retention
	// predicate, because a transition must never outlive the summaries that justified it.
	bucketSummaries   = "summaries"
	bucketTransitions = "transitions"

	// dayLayout is the key prefix for both buckets, so a single time comparison prunes both.
	dayLayout = "2006-01-02"

	// RetentionDays is the summary window. It is also the classification history limit: evidence
	// that has been pruned cannot contribute to a class.
	RetentionDays = 30

	// writeFloorHours and writeFloorPackets gate what is stored. A device heard once in one hour is
	// noise, and the maintainer's requirement was explicitly to store data without filling the
	// database with it. Enforced here, at the boundary, so no caller can forget it.
	writeFloorHours   = 2
	writeFloorPackets = 20

	// maxRowsPerRead bounds a response. A long history must not become an unbounded payload.
	maxRowsPerRead = 30
)

// Repository is the persistence this store provides.
//
// It is an interface so callers can be tested without a database, and so the recorder depends on a
// capability rather than on this concrete implementation.
type Repository interface {
	SaveSummary(summary model.DailySummary) error
	SaveTransition(transition model.Transition) error
	Summaries(key string, limit int) ([]model.DailySummary, error)
	Transitions(key string, limit int) ([]model.Transition, error)
	Prune(before time.Time) error
	Close() error
}

// Store is a bbolt-backed repository in its own file.
//
// It is safe for concurrent use because bbolt serialises transactions; the recorder further
// serialises writes through one worker, so contention is not a concern at these volumes.
type Store struct {
	kv storage.KeyValueDatabase
}

// Open opens the summaries database in the data directory.
func Open() (*Store, error) {
	return OpenAt(filepath.Join(utils.GetDataDir(), filename))
}

// OpenAt opens a store at an explicit path. Tests use this.
//
// No file is created until the database is opened, which is why callers only open it when there is
// something to write: an idle installation must not gain an empty database.
func OpenAt(path string) (*Store, error) {
	kv, err := storage.NewBoltKeyValueDatabase(path, bucketSummaries)
	if err != nil {
		return nil, err
	}
	return &Store{kv: kv}, nil
}

// Close releases the database file.
func (s *Store) Close() error { return s.kv.Close() }

// The store satisfies the recorder's needs; the assertion keeps the two from drifting apart.
var _ Repository = (*Store)(nil)

// SaveSummary stores one day of evidence for one identity.
//
// Writes below the floor are dropped silently rather than returned as an error: a day that does not
// qualify is not a failure, it is the normal case for a passing device. Returning an error would
// force every caller to handle a condition that is expected.
func (s *Store) SaveSummary(summary model.DailySummary) error {
	if summary.Key == "" || summary.Day == "" {
		return nil
	}
	if !meetsFloor(summary) {
		return nil
	}
	return s.kv.SetBatch(bucketSummaries, map[string]any{summary.Key: summary},
		func(key string, value any) ([]byte, []byte, error) {
			stored := value.(model.DailySummary)
			encoded, err := json.Marshal(stored)
			if err != nil {
				return nil, nil, err
			}
			return summaryKey(key, stored.Day), encoded, nil
		})
}

// SaveTransition stores one class change.
//
// The reason is stored with the row, not looked up later. A transition is a log entry: by the time
// an operator reads it, the summaries that produced it may have been pruned, so a pointer would
// often dangle.
func (s *Store) SaveTransition(transition model.Transition) error {
	if transition.Key == "" || transition.Day == "" {
		return nil
	}
	return s.kv.SetBatch(bucketTransitions, map[string]any{transition.Key: transition},
		func(key string, value any) ([]byte, []byte, error) {
			stored := value.(model.Transition)
			encoded, err := json.Marshal(stored)
			if err != nil {
				return nil, nil, err
			}
			return transitionKey(key, stored), encoded, nil
		})
}

// Summaries returns one identity's summaries, newest first, bounded to limit rows.
//
// Newest first because every consumer wants recent evidence: the classifier weights by age and the
// UI shows a timeline, so both would otherwise sort the same slice.
func (s *Store) Summaries(key string, limit int) ([]model.DailySummary, error) {
	if key == "" {
		return nil, nil
	}
	limit = boundLimit(limit)

	var rows []model.DailySummary
	if err := s.kv.ViewInRange(bucketSummaries, prefixFrom(key), prefixTo(key), func(_, value []byte) error {
		var summary model.DailySummary
		if err := json.Unmarshal(value, &summary); err != nil {
			// One unreadable row must not hide the rest of the history; skip it and continue.
			return nil
		}
		rows = append(rows, summary)
		return nil
	}); err != nil {
		if errors.Is(err, bolt.ErrBucketNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return newestFirst(rows, limit), nil
}

// Transitions returns one identity's class changes, newest first, bounded to limit rows.
func (s *Store) Transitions(key string, limit int) ([]model.Transition, error) {
	if key == "" {
		return nil, nil
	}
	limit = boundLimit(limit)

	var rows []model.Transition
	if err := s.kv.ViewInRange(bucketTransitions, prefixFrom(key), prefixTo(key), func(_, value []byte) error {
		var transition model.Transition
		if err := json.Unmarshal(value, &transition); err != nil {
			return nil
		}
		rows = append(rows, transition)
		return nil
	}); err != nil {
		if errors.Is(err, bolt.ErrBucketNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return newestFirstTransitions(rows, limit), nil
}

// Prune deletes rows older than the cutoff.
//
// Both buckets are pruned on the same cutoff by the same call, because a transition that outlived
// its evidence could be displayed but not explained.
func (s *Store) Prune(before time.Time) error {
	cutoff := before.UTC().Format(dayLayout)
	err := s.kv.Prune(func(key []byte) (bool, error) {
		day, ok := keyDay(key)
		if !ok {
			// An unparseable key cannot be aged, so it is left alone rather than deleted on a guess.
			return false, nil
		}
		return day < cutoff, nil
	})
	if errors.Is(err, bolt.ErrBucketNotFound) {
		return nil
	}
	return err
}

// meetsFloor reports whether a day is worth keeping.
func meetsFloor(summary model.DailySummary) bool {
	return summary.ActiveHours >= writeFloorHours && summary.Packets >= writeFloorPackets
}

// boundLimit clamps a requested limit into the range this store will serve.
func boundLimit(limit int) int {
	if limit <= 0 || limit > maxRowsPerRead {
		return maxRowsPerRead
	}
	return limit
}

// summaryKey encodes "<identity>|<day>".
//
// Identity leads because every read is "this identity's history", and a range scan is only possible
// when the identity is the prefix. Day-first was the obvious choice for pruning — one ordered scan —
// but it scatters a single identity's rows across the whole keyspace, which makes the read a full
// scan and lets a neighbouring identity's rows leak into the range. Pruning instead parses the day
// out of the key, which costs one cheap parse and buys a correct indexed read.
func summaryKey(key, day string) []byte {
	return []byte(key + "|" + day)
}

// transitionKey encodes "<identity>|<day>|<at>".
//
// Identity leads for the same reason as summaries. Including the time keeps same-day transitions
// ordered and distinct; two changes on one day are a real possibility when a device's evidence flips
// either side of the hysteresis margin.
func transitionKey(key string, transition model.Transition) []byte {
	return []byte(key + "|" + transition.Day + "|" + transition.At)
}

// keyDay extracts the day from a storage key of the form "<identity>|<day>[|...]".
func keyDay(key []byte) (string, bool) {
	parts := strings.SplitN(string(key), "|", 3)
	if len(parts) < 2 {
		return "", false
	}
	if _, err := time.Parse(dayLayout, parts[1]); err != nil {
		return "", false
	}
	return parts[1], true
}

// prefixFrom returns the first key belonging to an identity.
func prefixFrom(key string) []byte {
	return []byte(key + "|")
}

// prefixTo returns a bound above every key belonging to an identity.
//
// The separator is included on both ends so identity "a" cannot match identity "ab": a trailing "|"
// (0x7c) sorts above any day digit, and 0xff sorts above any trailing byte the suffix can contain.
func prefixTo(key string) []byte {
	return append([]byte(key+"|"), 0xff)
}

// newestFirst sorts summaries by day descending and truncates to limit.
func newestFirst(rows []model.DailySummary, limit int) []model.DailySummary {
	sortByDayDesc(len(rows), func(i, j int) bool { return rows[i].Day > rows[j].Day },
		func(i, j int) { rows[i], rows[j] = rows[j], rows[i] })
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}

// newestFirstTransitions sorts transitions by day then time descending and truncates to limit.
func newestFirstTransitions(rows []model.Transition, limit int) []model.Transition {
	sortByDayDesc(len(rows), func(i, j int) bool {
		if rows[i].Day != rows[j].Day {
			return rows[i].Day > rows[j].Day
		}
		return rows[i].At > rows[j].At
	}, func(i, j int) { rows[i], rows[j] = rows[j], rows[i] })
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}

// sortByDayDesc is an insertion sort.
//
// The slices here are at most a few dozen rows and are almost always already ordered by the key
// scan, so insertion sort is both faster and dependency-free compared with sort.Slice, and it
// keeps this package's imports to the minimum.
func sortByDayDesc(n int, less func(i, j int) bool, swap func(i, j int)) {
	for i := 1; i < n; i++ {
		for j := i; j > 0 && less(j, j-1); j-- {
			swap(j, j-1)
		}
	}
}
