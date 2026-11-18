package observations

// Classification tests. Every case states the behaviour in terms an operator would recognise,
// because the class is the only output they see and a wrong one is worse than no class at all.

import (
	"testing"
	"time"

	model "node-herder/models/observations"
)

// summaryOn builds a summary relative to a fixed reference day, so recency weighting is
// exercised without depending on when the test runs.
//
// ageDays is how many days before the reference day the summary belongs to.
func summaryOn(ageDays int, mutate func(*model.DailySummary)) model.DailySummary {
	s := model.DailySummary{
		Key:         "addr:aa:bb:cc:dd:ee:01",
		Day:         classifyReferenceDay.AddDate(0, 0, -ageDays).Format("2006-01-02"),
		Packets:     1000,
		ActiveHours: 6,
		HourMask:    hoursOf(9, 10, 11, 14, 15, 16),
		RSSIMean:    -65,
		Strongest:   -60,
		Weakest:     -70,
		Crossings:   0,
	}
	if mutate != nil {
		mutate(&s)
	}
	return s
}

// classifyReferenceDay is the "today" all classification tests use.
var classifyReferenceDay = time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)

// hoursOf builds an hour mask from hour numbers.
func hoursOf(hours ...int) uint32 {
	var mask uint32
	for _, hour := range hours {
		mask |= 1 << uint(hour)
	}
	return mask
}

// allDay returns a mask covering every hour.
func allDay() uint32 { return 0xFFFFFF }

func TestClassifyInfrastructure(t *testing.T) {
	// Always present, always the same signal, never moves: a fridge, a hub, a TV on standby.
	var days []model.DailySummary
	for age := 0; age < 7; age++ {
		days = append(days, summaryOn(age, func(s *model.DailySummary) {
			s.ActiveHours = 24
			s.HourMask = allDay()
			s.Packets = 20000
			s.Crossings = 0
		}))
	}

	got := classify(days, classifyReferenceDay)

	if got.Class != ClassInfrastructure {
		t.Errorf("class = %q, want infrastructure (evidence: %+v)", got.Class, got.Evidence)
	}
}

func TestClassifyPersonal(t *testing.T) {
	// Present most days but with a gap in the middle of the night, and it moves.
	var days []model.DailySummary
	for age := 0; age < 5; age++ {
		days = append(days, summaryOn(age, func(s *model.DailySummary) {
			s.ActiveHours = 16
			s.HourMask = hoursOf(7, 8, 9, 12, 13, 17, 18, 19, 20, 21, 22)
			s.Crossings = 8
		}))
	}

	got := classify(days, classifyReferenceDay)

	if got.Class != ClassPersonal {
		t.Errorf("class = %q, want personal (evidence: %+v)", got.Class, got.Evidence)
	}
}

func TestClassifyNeighbour(t *testing.T) {
	// Seen regularly but always weakly compared with the strongest device in the room.
	var days []model.DailySummary
	for age := 0; age < 5; age++ {
		days = append(days, summaryOn(age, func(s *model.DailySummary) {
			s.ActiveHours = 10
			s.HourMask = hoursOf(8, 9, 10, 19, 20, 21)
			s.RSSIMean = -88
			s.Strongest = -85
			s.Weakest = -92
		}))
	}

	got := classifyWithReference(days, classifyReferenceDay, -55)

	if got.Class != ClassNeighbour {
		t.Errorf("class = %q, want neighbour (evidence: %+v)", got.Class, got.Evidence)
	}
}

func TestClassifyTransient(t *testing.T) {
	// A single appearance, well under an hour: someone walked past.
	got := classify([]model.DailySummary{
		summaryOn(0, func(s *model.DailySummary) {
			s.Packets = 4
			s.ActiveHours = 1
			s.HourMask = hoursOf(15)
		}),
	}, classifyReferenceDay)

	if got.Class != ClassTransient {
		t.Errorf("class = %q, want transient (evidence: %+v)", got.Class, got.Evidence)
	}
}

func TestClassifyLearningBelowThreshold(t *testing.T) {
	// Under 24 hours of evidence no class may be asserted; a guess presented as a class is worse
	// than an honest "still learning", because the operator cannot tell the two apart.
	got := classify([]model.DailySummary{
		summaryOn(0, func(s *model.DailySummary) {
			s.ActiveHours = 3
			s.HourMask = hoursOf(9, 10, 11)
		}),
	}, classifyReferenceDay)

	if got.Class != ClassLearning {
		t.Errorf("class = %q, want learning", got.Class)
	}
	if got.HoursObserved >= hoursRequiredForClassification {
		t.Errorf("HoursObserved = %d, want below %d", got.HoursObserved, hoursRequiredForClassification)
	}
}

func TestClassifyIgnoresASingleContraryDay(t *testing.T) {
	// A long-standing infrastructure device that moved for one day must not be reclassified.
	// Otherwise a burglar, or a device being carried through the house, rewrites its history.
	var days []model.DailySummary
	for age := 1; age <= 14; age++ {
		days = append(days, summaryOn(age, func(s *model.DailySummary) {
			s.ActiveHours = 24
			s.HourMask = allDay()
		}))
	}
	days = append(days, summaryOn(0, func(s *model.DailySummary) {
		s.ActiveHours = 4
		s.HourMask = hoursOf(20, 21, 22, 23)
		s.Crossings = 20
	}))

	if got := classify(days, classifyReferenceDay); got.Class != ClassInfrastructure {
		t.Errorf("class = %q after one contrary day, want infrastructure", got.Class)
	}
}

func TestClassifyFollowsSustainedChange(t *testing.T) {
	// With a 2-day half-life, two days of consistent new behaviour must be enough. This is the
	// measured consequence recorded in plan §2.3 and AC-17; if the weighting changes, this test
	// is supposed to fail and force a deliberate decision.
	var days []model.DailySummary
	for age := 2; age <= 16; age++ {
		days = append(days, summaryOn(age, func(s *model.DailySummary) {
			s.ActiveHours = 24
			s.HourMask = allDay()
			s.Crossings = 0
		}))
	}
	for age := 0; age < 2; age++ {
		days = append(days, summaryOn(age, func(s *model.DailySummary) {
			s.ActiveHours = 14
			s.HourMask = hoursOf(8, 9, 10, 11, 12, 13, 18, 19, 20, 21)
			s.Crossings = 12
		}))
	}

	if got := classify(days, classifyReferenceDay); got.Class != ClassPersonal {
		t.Errorf("class = %q after two days of movement, want personal (evidence: %+v)",
			got.Class, got.Evidence)
	}
}

func TestClassifyHysteresisHoldsAMarginalResult(t *testing.T) {
	// A class that wins by a hair must not produce a transition. The incumbent is held and the
	// challenger reported as pending, so the UI can say "leaning towards X" without claiming a
	// change that the evidence does not clearly support.
	//
	// Deliberately ambiguous evidence: present all day, but moving more than an installed device
	// should. Measured scores are infrastructure 0.600 against personal 0.530 — a margin of 0.070,
	// below the 0.15 hysteresis threshold. These numbers are pinned because the fixture is only
	// marginal by that arithmetic; if a weighting changes, this test should be revisited rather
	// than quietly passing for a different reason.
	days := []model.DailySummary{
		summaryOn(0, func(s *model.DailySummary) {
			s.ActiveHours = 24
			s.HourMask = allDay()
			s.Crossings = 4
		}),
	}

	candidate := classify(days, classifyReferenceDay)
	if candidate.Class != ClassInfrastructure {
		t.Fatalf("class = %q, want the fixture to favour infrastructure", candidate.Class)
	}

	// The incumbent is the challenger's opposite, so the thin margin must hold the incumbent.
	held := classifyWithPrevious(days, classifyReferenceDay, Classified{Class: ClassPersonal})

	if held.Class != ClassPersonal {
		t.Errorf("class = %q, want the incumbent personal held while the margin is thin", held.Class)
	}
	if !held.ChangePending || held.Candidate != ClassInfrastructure {
		t.Errorf("ChangePending = %v Candidate = %q, want a pending change to infrastructure",
			held.ChangePending, held.Candidate)
	}
}

func TestClassifyHysteresisAllowsADecisiveChange(t *testing.T) {
	// A decisive challenger must not be held back, or a device that genuinely moved would keep its
	// old class for as long as the operator watched.
	days := []model.DailySummary{
		summaryOn(0, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
		summaryOn(1, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
		summaryOn(2, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
	}

	got := classifyWithPrevious(days, classifyReferenceDay, Classified{Class: ClassPersonal})

	if got.Class != ClassInfrastructure {
		t.Errorf("class = %q, want a decisive infrastructure result to be accepted", got.Class)
	}
	if got.ChangePending {
		t.Errorf("ChangePending = true, want no pending change once the change is accepted")
	}
}

func TestClassifyEvidenceDescribesWhy(t *testing.T) {
	days := []model.DailySummary{
		summaryOn(0, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
	}

	got := classify(days, classifyReferenceDay)

	if got.Evidence.DutyCycle <= 0 {
		t.Errorf("DutyCycle = %f, want the fraction of the day observed", got.Evidence.DutyCycle)
	}
	if got.Evidence.HoursObserved == 0 {
		t.Errorf("HoursObserved = 0, want the observed hours")
	}
	// Every class must be explainable; a bare label cannot be argued with or trusted.
	if got.Evidence.Reason == "" {
		t.Errorf("Reason is empty, want a human-readable justification")
	}
}

func TestClassifyHandlesNoEvidence(t *testing.T) {
	got := classify(nil, classifyReferenceDay)

	if got.Class != ClassLearning {
		t.Errorf("class = %q with no evidence, want learning", got.Class)
	}
	if got.Evidence.Reason == "" {
		t.Errorf("Reason is empty, want an explanation of why there is no class")
	}
}

func TestClassifyIgnoresSummariesOlderThanRetention(t *testing.T) {
	// Summaries beyond the retention window cannot be evidence, because they do not exist in
	// the store; a caller passing them must not extend the apparent history.
	var days []model.DailySummary
	for age := 0; age < retentionDays; age++ {
		days = append(days, summaryOn(age, func(s *model.DailySummary) {
			s.ActiveHours = 24
			s.HourMask = allDay()
		}))
	}
	days = append(days, summaryOn(retentionDays+5, func(s *model.DailySummary) {
		s.ActiveHours = 24
		s.HourMask = allDay()
	}))

	got := classify(days, classifyReferenceDay)

	if got.DaysConsidered > retentionDays {
		t.Errorf("DaysConsidered = %d, want at most %d", got.DaysConsidered, retentionDays)
	}
}

func TestClassifyNightPresenceDistinguishesAlwaysOn(t *testing.T) {
	// "Seen at 04:00" is the strongest single discriminator between infrastructure and a person,
	// so it is evidence with weight rather than a verdict on its own.
	alwaysOn := []model.DailySummary{summaryOn(0, func(s *model.DailySummary) {
		s.ActiveHours = 24
		s.HourMask = allDay()
	})}
	if got := classify(alwaysOn, classifyReferenceDay); !got.Evidence.NightPresence {
		t.Errorf("NightPresence = false for an all-day mask, want true")
	}

	daytime := []model.DailySummary{summaryOn(0, func(s *model.DailySummary) {
		s.ActiveHours = 8
		s.HourMask = hoursOf(9, 10, 11, 12, 13, 14, 15, 16)
	})}
	if got := classify(daytime, classifyReferenceDay); got.Evidence.NightPresence {
		t.Errorf("NightPresence = true for a daytime-only mask, want false")
	}
}
