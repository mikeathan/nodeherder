package observations

// Provisional classification: a class decided from the first packet, before any history exists.
//
// Why this file exists: the default filter hides infrastructure, but classification needs 24 hours
// of evidence. Without a provisional class the first day is therefore the entire census — the exact
// complaint this work exists to fix. Two signals are available immediately and are safe to act on,
// so they carry the filter until the classifier has evidence.
//
// What it deliberately does not do: guess infrastructure. Plenty of personal devices use a static
// address, so hiding one the operator cares about would be a worse failure than showing one they do
// not. Provisionals are shown as provisional precisely because they are weak claims.

import (
	"time"

	model "node-herder/models/observations"
)

// Provisional class values. They are distinct from the classifier's classes so a consumer cannot
// mistake one for a verdict, and so a stored class can never be confused with a live guess.
const (
	// ProvisionalMobile is a device with a rotating address: someone's phone, watch or earbuds.
	ProvisionalMobile = "mobile-unidentifiable"
	// ProvisionalNeighbour is a device never heard strongly, so probably beyond the room.
	ProvisionalNeighbour = "likely-neighbour"
)

// neighbourWeakDB is how far below the strongest device seen in this session an identity must stay
// to be considered probably outside. It mirrors the classifier's neighbourDeltaDB so the provisional
// guess and the settled class do not disagree about what "far" means.
const neighbourWeakDB = 10.0

// Provisional is a class guessed from the first packet.
//
// It carries whether to show the device by default, because that decision is the whole point: the
// class alone would not stop the census from filling the screen.
type Provisional struct {
	// Class is one of the provisional values, or ClassLearning when neither signal applies.
	Class string
	// ShowByDefault reports whether the device should be visible without the operator revealing it.
	ShowByDefault bool
	// Provisional is always true for a guessed class, and lets a consumer render it as unsettled.
	Provisional bool
	// Reason explains the guess in the operator's terms.
	Reason string
}

// asClassified adapts a provisional guess to the shape a classification result takes.
//
// It exists so a provisional class can be carried in the same field as a real class and then be
// superseded by one. The provisional values are intentionally not real classes, so a consumer that
// only understands Classified still receives them as a class string with Provisional set, which is
// what makes "show the guess, but say it is a guess" possible without a second code path.
func (p Provisional) asClassified() Classified {
	return Classified{
		Class:       p.Class,
		Provisional: p.Provisional,
		Evidence:    EvidenceView{Reason: p.Reason},
	}
}

// provisionalClass decides the pre-filter result for one identity.
//
// strongestRSSI is the strongest reading seen in the current session, and zero when nothing has
// been measured yet. It is a session-scoped reference rather than a stored one because the point is
// to classify before any storage exists.
func provisionalClass(identity model.Identity, rssi, strongestRSSI int, _ time.Time) Provisional {
	// A rotating address is the strongest available signal, and it needs no reference: the device
	// rotates, therefore it belongs to a person, therefore it is not installed equipment.
	if !identity.Stable {
		return Provisional{
			Class:         ProvisionalMobile,
			ShowByDefault: true,
			Provisional:   true,
			Reason:        "Address rotates, so this is a person's device rather than household equipment.",
		}
	}

	// With no reference there is nothing to compare against, so no device can be called far away.
	// Returning learning keeps it visible, which is the safe failure.
	if strongestRSSI == 0 || rssi == 0 {
		return Provisional{
			Class:         ClassLearning,
			ShowByDefault: true,
			Reason:        "Not enough reading yet to say anything about this device.",
		}
	}

	if float64(strongestRSSI-rssi) > neighbourWeakDB {
		return Provisional{
			Class: ProvisionalNeighbour,
			// Shown, not hidden.
			//
			// Hiding a weak stable device was meant to reduce noise, but it inverted the intent:
			// rotating addresses are always shown (they are the `mobile-unidentifiable` class), and
			// a stable device is the only kind that can be identified and plotted. So the default
			// filter removed exactly the devices the operator could see, while keeping the
			// anonymous ones — a radar that showed one dot out of forty.
			//
			// The class still says "probably outside"; it just does not hide it.
			ShowByDefault: true,
			Provisional:   true,
			Reason:        "Always much weaker than devices in the room, so probably outside it.",
		}
	}

	// Stable and heard clearly: it may well be household equipment, so it stays visible. Calling it
	// infrastructure here would be a guess with no evidence behind it.
	return Provisional{
		Class:         ClassLearning,
		ShowByDefault: true,
		Reason:        "Heard clearly, but not enough history yet to say what it is.",
	}
}
