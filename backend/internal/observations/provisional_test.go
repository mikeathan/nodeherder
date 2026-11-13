package observations

// Provisional classification tests.
//
// The provisional class exists so the first day of observation is not the full census. It uses
// only signals present on the first packet, so it must be clearly weaker than the classifier and
// must never be presented as settled. These tests pin both properties: what it decides, and that
// it gets out of the way once real evidence exists.

import (
	"testing"
	"time"

	model "node-herder/models/observations"
)

// firstPacket builds evidence from a single advertisement.
func firstPacket(identity model.Identity, rssi int) model.Observation {
	return model.Observation{
		Source:     "esp-1",
		Medium:     model.MediumBLEAdvertisement,
		Identity:   identity,
		RSSI:       rssi,
		ObservedAt: classifyReferenceDay,
	}
}

func TestProvisionalRotatingAddressIsMobile(t *testing.T) {
	// A rotating address means the device belongs to a person, not to the building. It can never be
	// household infrastructure, so it is worth showing immediately even with no history.
	identity := model.Identity{Key: "fp:9a3c1f", Kind: model.KindAdvertisement, Stable: false}

	got := provisionalClass(identity, -70, -55, time.Time{})

	if got.Class != ProvisionalMobile {
		t.Errorf("class = %q, want %q for a rotating address", got.Class, ProvisionalMobile)
	}
	if !got.ShowByDefault {
		t.Errorf("ShowByDefault = false, want true: a person's device is worth seeing")
	}
	if !got.Provisional {
		t.Errorf("Provisional = false, want true")
	}
}

func TestProvisionalNeverStrongIsLikelyNeighbour(t *testing.T) {
	// A stable address that has never been heard strongly is probably beyond the room. Hiding it by
	// default is safe because the operator can reveal it, whereas hiding a device they care about
	// would look like a bug.
	identity := model.Identity{Key: "addr:aa:bb:cc:dd:ee:02", Kind: model.KindAdvertisement, Stable: true}

	got := provisionalClass(identity, -92, -55, time.Time{})

	if got.Class != ProvisionalNeighbour {
		t.Errorf("class = %q, want %q for a device never heard strongly", got.Class, ProvisionalNeighbour)
	}
	if got.ShowByDefault {
		t.Errorf("ShowByDefault = true, want false: an outside device should be hidden by default")
	}
}

func TestProvisionalStableAndStrongStaysVisible(t *testing.T) {
	// A stable address heard strongly may well be household equipment the operator cares about.
	// It stays visible and honestly marked as still learning, because guessing "infrastructure" from
	// a static address alone would hide personal devices that use one.
	identity := model.Identity{Key: "addr:aa:bb:cc:dd:ee:03", Kind: model.KindAdvertisement, Stable: true}

	got := provisionalClass(identity, -58, -55, time.Time{})

	if got.Class != ClassLearning {
		t.Errorf("class = %q, want %q: a stable strong device must not be called infrastructure yet",
			got.Class, ClassLearning)
	}
	if !got.ShowByDefault {
		t.Errorf("ShowByDefault = false, want true: never hide a device that might be yours")
	}
}

func TestProvisionalIsSupersededByRealEvidence(t *testing.T) {
	// Once a day of evidence exists the provisional class must give way, or the pre-filter would
	// become a permanent verdict that no amount of observation could overturn.
	identity := model.Identity{Key: "fp:9a3c1f", Kind: model.KindAdvertisement, Stable: false}
	var days []model.DailySummary
	for age := 0; age < 4; age++ {
		days = append(days, summaryOn(age, func(s *model.DailySummary) {
			s.ActiveHours = 22
			s.HourMask = allDay()
			s.Packets = 9000
		}))
	}

	got := classifyWithPrevious(days, classifyReferenceDay, provisionalClass(identity, -60, -60, time.Time{}).asClassified())

	if got.Provisional {
		t.Errorf("Provisional = true after four days of evidence, want the classifier's result")
	}
	if got.Class == ProvisionalMobile {
		t.Errorf("class = %q, want a real class to supersede the provisional one", got.Class)
	}
}

func TestProvisionalTruncatesRatherThanHidesUnknown(t *testing.T) {
	// Two signals are known to be safe to demote; anything else stays visible. A device the
	// pre-filter cannot judge must not be hidden, because a missing device looks like a fault.
	cases := []struct {
		name      string
		identity  model.Identity
		rssi      int
		strongest int
		want      string
	}{
		{
			name:      "rotating address",
			identity:  model.Identity{Key: "fp:1", Stable: false},
			rssi:      -80,
			strongest: -55,
			want:      ProvisionalMobile,
		},
		{
			name:      "stable but never strong",
			identity:  model.Identity{Key: "addr:aa:bb:cc:dd:ee:04", Stable: true},
			rssi:      -95,
			strongest: -55,
			want:      ProvisionalNeighbour,
		},
		{
			name:      "stable and strong",
			identity:  model.Identity{Key: "addr:aa:bb:cc:dd:ee:05", Stable: true},
			rssi:      -50,
			strongest: -50,
			want:      ClassLearning,
		},
		{
			// No reference yet: with nothing stronger seen, no device can be called far away.
			name:      "no reference available",
			identity:  model.Identity{Key: "addr:aa:bb:cc:dd:ee:06", Stable: true},
			rssi:      -95,
			strongest: 0,
			want:      ClassLearning,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := provisionalClass(test.identity, test.rssi, test.strongest, time.Time{}); got.Class != test.want {
				t.Errorf("class = %q, want %q", got.Class, test.want)
			}
		})
	}
}
