package observations

// This file turns stored evidence into an advisory class. It is a pure function of the summaries
// it is given and an injected "now", so every rule below is directly testable and nothing here
// reads a clock or a database.
//
// Three decisions shape the whole file:
//
//  1. The class is advisory. An operator's triage decision always outranks it, and the UI shows
//     both. This is what makes it safe to guess: a wrong guess is visibly a guess.
//  2. Classes are recomputed from evidence, never stored. Changing a threshold therefore
//     reclassifies history immediately, and a device whose behaviour changed is judged on the days
//     that actually happened rather than on a verdict recorded months ago.
//  3. Each day is scored separately and the scores are recency-weighted. Averaging the evidence
//     first and scoring once would let a single unusual day drag a long history with it, which is
//     exactly the "don't assume the 4am sighting is the fridge" requirement. Weighting per day
//     means a lone contrary day is one small voice, not a rewrite.

import (
	"math"
	"time"

	model "node-herder/models/observations"
)

// Class values. These are the only classes an operator ever sees, plus learning, which means
// "not enough evidence yet" and is rendered as a learning state rather than as a class.
const (
	ClassInfrastructure = "infrastructure"
	ClassPersonal       = "personal"
	ClassNeighbour      = "neighbour"
	ClassTransient      = "transient"
	ClassLearning       = "learning"
)

// Classification parameters. Each is a judgement recorded as a constant so a reader can find and
// argue with it, and so a test can pin the behaviour it produces.
const (
	// classHalfLifeDays is the recency half-life: evidence loses half its weight every two days.
	//
	// Two days, not five (maintainer decision, 2026-11-12): the operator had already spent weeks on
	// data they did not trust, so three days to a judgement beats a week. The measured consequence
	// is that sustained contrary behaviour wins after about two days — see plan §2.3.
	classHalfLifeDays = 2.0

	// retentionDays bounds how much history can be evidence. It matches the summary store's
	// retention, because evidence that has been pruned cannot contribute.
	retentionDays = 30

	// hoursRequiredForClassification is the evidence floor. Under 24 observed hours a duty cycle is
	// meaningless — one evening at home looks identical to a permanently installed device.
	hoursRequiredForClassification = 24

	// transientHours is the total observed hours at or below which a device is a passer-by.
	//
	// Only stored daily summaries contribute, and one is written only when a day passed the write
	// floor of two distinct hours and twenty packets. A contributing summary therefore always
	// represents real observation, so "one hour in total" means a device that was genuinely almost
	// never there rather than a session that had only just started.
	transientHours = 1

	// infrastructureMinDutyCycle is the fraction of the day below which a device cannot be called
	// always-present, however little it moves.
	infrastructureMinDutyCycle = 0.60

	// infrastructureFullDutyCycle is the duty cycle at which presence alone scores full marks.
	infrastructureFullDutyCycle = 0.90

	// infrastructureMaxCrossings is the movement allowed before a device stops looking installed.
	// A fridge does not move, so movement argues against infrastructure.
	infrastructureMaxCrossings = 2.0

	// neighbourDeltaDB is how far below the strongest device in the room a signal must stay to be
	// demoted to neighbour. Ten dB is roughly a factor of three in distance at indoor path loss, so
	// this separates "in the room" from "through a wall".
	neighbourDeltaDB = 10.0

	// personalBaselineScore is the floor for a device that is neither always-present nor clearly
	// outside. Everyone's phone lives here: present sometimes, and it moves.
	personalBaselineScore = 0.45

	// hysteresisMargin is how far a challenger must beat the incumbent before a change is
	// announced.
	//
	// This is a Schmitt trigger, not a vote count. An earlier draft required a candidate to win two
	// consecutive *evaluations*, which is meaningless: evaluations run per snapshot, seconds apart,
	// so two evaluations pass almost instantly and the guard protected nothing. A margin is
	// stateless and does the job the guard was for — a class that wins by a hair does not produce a
	// transition row, while a decisive change is announced at once.
	hysteresisMargin = 0.15

	// nightHoursStart and nightHoursEnd bound the quiet period used to recognise always-on
	// equipment. A device heard at 04:00 every night is almost certainly not a person.
	nightHoursStart = 1
	nightHoursEnd   = 5
)

// Classified is the result of classifying one identity from its evidence.
//
// It carries the class, the candidate that might replace it, and the evidence behind both, so a
// consumer can show why rather than only what. A reloaded page, or an operator arguing with the
// result, needs the reasoning and not just the verdict.
type Classified struct {
	// Class is the advisory class, or ClassLearning when there is not enough evidence.
	Class string `json:"class"`
	// Provisional is true when the value came from the first-packet pre-filter rather than from the
	// classifier, and must be presented as unsettled.
	Provisional bool `json:"provisional,omitempty"`
	// Candidate is the class the evidence favours when hysteresis has held the incumbent back;
	// empty when no change is pending.
	Candidate string `json:"candidate,omitempty"`
	// ChangePending is true when Candidate differs from Class.
	ChangePending bool `json:"changePending,omitempty"`
	// Confidence is the margin over the runner-up, 0-1. Zero while learning.
	Confidence float64 `json:"confidence"`
	// HoursObserved is how much evidence exists, so the UI can say how far learning has to go.
	HoursObserved int `json:"hoursObserved"`
	// DaysConsidered is how many daily summaries contributed.
	DaysConsidered int `json:"daysConsidered"`
	// Evidence explains the class in terms an operator can check.
	Evidence EvidenceView `json:"evidence"`
}

// EvidenceView is the explanation behind a class.
//
// It is evidence rather than a score dump: an operator can read "present 22 of 24 hours, never
// moved" and decide whether they agree, which is the whole point of showing it.
type EvidenceView struct {
	// DutyCycle is the recency-weighted fraction of the day the device was observed.
	DutyCycle float64 `json:"dutyCycle"`
	// HoursObserved is the total observed hours across the contributing days.
	HoursObserved int `json:"hoursObserved"`
	// Crossings is the weighted mean movement signal per day.
	Crossings float64 `json:"crossings"`
	// Strongest is the strongest reading seen across the window, or 0 when unmeasured.
	Strongest int `json:"strongest"`
	// MeanRSSI is the weighted mean signal across the window, or 0 when unmeasured.
	MeanRSSI float64 `json:"meanRssi"`
	// NightPresence is true when the device was heard during the quiet period.
	NightPresence bool `json:"nightPresence"`
	// Reason is a short human-readable justification.
	Reason string `json:"reason"`
}

// classifiedDay is one summary prepared for scoring: its evidence plus the weight its age earns.
//
// Scoring works on this rather than on the raw summary so the rules never need to know about time,
// and the weighting is applied in exactly one place.
type classifiedDay struct {
	weight    float64
	summary   model.DailySummary
	dutyCycle float64
	crossings float64
	night     bool
}

// classify is the entry point for callers with no previous result.
func classify(summaries []model.DailySummary, now time.Time) Classified {
	return classifyInternal(summaries, now, 0, Classified{})
}

// classifyWithPrevious re-evaluates with the last announced class so hysteresis can hold an
// incumbent that a challenger has not yet clearly beaten.
func classifyWithPrevious(summaries []model.DailySummary, now time.Time, previous Classified) Classified {
	return classifyInternal(summaries, now, 0, previous)
}

// classifyWithReference classifies while treating a given strongest signal as the room's reference.
// It exists so tests can state "this is the strongest device in the room" explicitly rather than
// have the function invent a reference from the same data it is judging.
func classifyWithReference(summaries []model.DailySummary, now time.Time, referenceRSSI int) Classified {
	return classifyInternal(summaries, now, referenceRSSI, Classified{})
}

// classifyInternal implements the rules.
func classifyInternal(summaries []model.DailySummary, now time.Time, referenceRSSI int, previous Classified) Classified {
	days := weigh(summaries, now)

	if len(days) == 0 {
		return Classified{
			Class:    ClassLearning,
			Evidence: EvidenceView{Reason: "No evidence recorded yet."},
		}
	}

	view := render(days)

	// Transient is checked before the learning floor because "almost never heard" is provable from
	// little evidence. Requiring 24 hours first would mean treating a passer-by as a resident before
	// it could be called a passer-by.
	if view.HoursObserved <= transientHours {
		view.Reason = "Heard only briefly in total, so this is likely a passer-by."
		return Classified{
			Class:          ClassTransient,
			Confidence:     0.9,
			HoursObserved:  view.HoursObserved,
			DaysConsidered: len(days),
			Evidence:       view,
		}
	}

	if view.HoursObserved < hoursRequiredForClassification {
		remaining := hoursRequiredForClassification - view.HoursObserved
		view.Reason = plural(remaining, "more observed hour", "more observed hours") +
			" needed before a class is asserted."
		return Classified{
			Class:          ClassLearning,
			HoursObserved:  view.HoursObserved,
			DaysConsidered: len(days),
			Evidence:       view,
		}
	}

	scores := score(days, view, referenceRSSI)
	winner, winnerScore := scores.best()
	runnerUp := scores.runnerUp(winner)

	view.Reason = reasonFor(winner, view)

	result := Classified{
		Class:          winner,
		Confidence:     clamp01(winnerScore - runnerUp),
		HoursObserved:  view.HoursObserved,
		DaysConsidered: len(days),
		Evidence:       view,
	}

	// Hysteresis: hold the incumbent unless the challenger wins clearly. This stops a class that
	// wins by a hair from producing a transition row, while leaving a decisive change immediate.
	if incumbent := previous.Class; isAssertedClass(incumbent) && incumbent != winner {
		if winnerScore-scores[incumbent] < hysteresisMargin {
			result.Class = incumbent
			result.Candidate = winner
			result.ChangePending = true
			result.Evidence.Reason = reasonFor(incumbent, view) +
				" A change to " + winner + " is not yet decisive."
		}
	}

	return result
}

// classScores holds one weighted score per asserted class.
//
// A score is the recency-weighted support for that class in [0, 1]. The classes are compared rather
// than tested in order, so a rule can be added or reweighted without silently changing which rule
// wins.
type classScores map[string]float64

// best returns the highest-scoring class and its score.
//
// Ties resolve by a fixed order, because Go map iteration is deliberately random and a class that
// flickered between two equally-supported answers would be worse than either.
func (s classScores) best() (string, float64) {
	order := []string{ClassTransient, ClassInfrastructure, ClassNeighbour, ClassPersonal}
	winner := ClassPersonal
	best := -1.0
	for _, class := range order {
		if score, ok := s[class]; ok && score > best {
			winner, best = class, score
		}
	}
	return winner, best
}

// runnerUp returns the second-highest score, which is what confidence is measured against.
func (s classScores) runnerUp(winner string) float64 {
	second := 0.0
	for class, score := range s {
		if class == winner {
			continue
		}
		if score > second {
			second = score
		}
	}
	return second
}

// score computes the weighted support for every class.
//
// Infrastructure and personal are properties of how a device behaves on a day, so they are scored
// per day and weighted. Neighbour is a property of the signal relative to the room, so it is scored
// once against the window's strongest reading.
func score(days []classifiedDay, view EvidenceView, referenceRSSI int) classScores {
	var weightTotal, infrastructure, personal float64
	for _, day := range days {
		weightTotal += day.weight
		infrastructure += day.weight * dayInfrastructureScore(day)
		personal += day.weight * dayPersonalScore(day)
	}

	scores := classScores{
		ClassTransient: 0,
		ClassNeighbour: scoreNeighbour(view, referenceRSSI),
		ClassPersonal:  personalBaselineScore,
	}
	if weightTotal > 0 {
		scores[ClassInfrastructure] = infrastructure / weightTotal
		scores[ClassPersonal] = personal / weightTotal
	}
	return scores
}

// dayInfrastructureScore is how much one day looks like installed equipment.
//
// Presence and stillness are multiplied, not added: "always there" and "never moved" are both
// required, because either alone describes something else — a phone is present and moves, a
// neighbour's meter is still but faint.
func dayInfrastructureScore(day classifiedDay) float64 {
	if day.dutyCycle < infrastructureMinDutyCycle {
		return 0
	}
	presence := (day.dutyCycle - infrastructureMinDutyCycle) /
		(infrastructureFullDutyCycle - infrastructureMinDutyCycle)
	presence = clamp01(presence)

	stillness := 1.0
	if day.crossings > 0 {
		stillness = clamp01(1 - day.crossings/(infrastructureMaxCrossings*4))
	}

	base := presence * stillness
	if day.night {
		// Heard overnight is the strongest sign of installed equipment rather than a person.
		base = clamp01(base + 0.1)
	}
	return base
}

// dayPersonalScore is how much one day looks like a device someone carries: present for part of the
// day, and moving. It rises with movement, because a device that stays put is better explained by
// infrastructure.
func dayPersonalScore(day classifiedDay) float64 {
	midDuty := clamp01(1 - math.Abs(day.dutyCycle-0.5)/0.5)
	movement := clamp01(day.crossings / 10)
	return clamp01(personalBaselineScore + 0.30*midDuty + 0.20*movement)
}

// scoreNeighbour is the only rule that needs a reference, because with one anchor distance is only
// ever a comparison. Without a reference it scores zero rather than guessing.
func scoreNeighbour(view EvidenceView, referenceRSSI int) float64 {
	if referenceRSSI == 0 || view.Strongest == 0 {
		return 0
	}
	delta := float64(referenceRSSI - view.Strongest)
	if delta <= neighbourDeltaDB {
		return 0
	}
	return clamp01(0.5 + (delta-neighbourDeltaDB)/30)
}

// weigh converts summaries into weighted days, discarding anything that cannot be evidence.
func weigh(summaries []model.DailySummary, now time.Time) []classifiedDay {
	today := startOfDay(now)
	days := make([]classifiedDay, 0, len(summaries))

	for _, summary := range summaries {
		day, err := time.Parse("2006-01-02", summary.Day)
		if err != nil {
			continue
		}
		age := int(today.Sub(day).Hours() / 24)
		if age < 0 {
			// A summary dated in the future is a clock problem, not evidence. Counting it would give
			// it more than full weight and let one bad timestamp dominate the class.
			continue
		}
		if age >= retentionDays {
			continue
		}

		days = append(days, classifiedDay{
			weight:    math.Pow(0.5, float64(age)/classHalfLifeDays),
			summary:   summary,
			dutyCycle: clamp01(float64(summary.ActiveHours) / 24),
			crossings: float64(summary.Crossings),
			night:     countBits(summary.HourMask, nightHoursStart, nightHoursEnd) > 0,
		})
	}

	return days
}

// render summarises the weighted days for display.
func render(days []classifiedDay) EvidenceView {
	var weightTotal, weightedDuty, weightedCrossings, weightedRSSI float64
	var hoursObserved int
	var strongest int
	measured := false

	for _, day := range days {
		weightTotal += day.weight
		weightedDuty += day.weight * day.dutyCycle
		weightedCrossings += day.weight * day.crossings
		weightedRSSI += day.weight * day.summary.RSSIMean
		hoursObserved += int(day.summary.ActiveHours)
		// RSSI 0 means unmeasured, not zero signal, so it must not become the strongest reading.
		if day.summary.Strongest != 0 && (!measured || day.summary.Strongest > strongest) {
			strongest = day.summary.Strongest
			measured = true
		}
	}

	view := EvidenceView{
		HoursObserved: hoursObserved,
		NightPresence: hasNightPresence(days),
	}
	if weightTotal > 0 {
		view.DutyCycle = clamp01(weightedDuty / weightTotal)
		view.Crossings = weightedCrossings / weightTotal
		view.MeanRSSI = weightedRSSI / weightTotal
	}
	if measured {
		view.Strongest = strongest
	}
	return view
}

// hasNightPresence reports whether any contributing day included the quiet period.
func hasNightPresence(days []classifiedDay) bool {
	for _, day := range days {
		if day.night {
			return true
		}
	}
	return false
}

// reasonFor renders the explanation for a class in terms an operator can check.
func reasonFor(class string, view EvidenceView) string {
	switch class {
	case ClassTransient:
		return "Heard only briefly in total, so this is likely a passer-by."
	case ClassInfrastructure:
		if view.NightPresence {
			return "Present almost all day, including overnight, and barely moved."
		}
		return "Present almost all day and barely moved, but not heard overnight."
	case ClassNeighbour:
		return "Always weaker than devices in the room, so likely beyond it rather than in it."
	default:
		if view.NightPresence {
			return "Present for part of the day, its signal changed, and it was heard overnight."
		}
		return "Present for part of the day with a changing signal, which suggests a device someone carries."
	}
}

// isAssertedClass reports whether a class is a real verdict rather than a placeholder.
func isAssertedClass(class string) bool {
	switch class {
	case ClassInfrastructure, ClassPersonal, ClassNeighbour, ClassTransient:
		return true
	default:
		return false
	}
}

// countBits returns how many hours in [start, end) are set in the mask.
func countBits(mask uint32, start, end int) int {
	count := 0
	for hour := start; hour < end; hour++ {
		if mask&(1<<uint(hour)) != 0 {
			count++
		}
	}
	return count
}

// startOfDay truncates to midnight UTC, so age arithmetic is in whole days.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// clamp01 bounds a value to [0, 1], which is what a score and a confidence must be.
func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

// plural renders "1 more observed hour" / "2 more observed hours".
func plural(count int, singular, pluralForm string) string {
	if count == 1 {
		return "1 " + singular
	}
	return itoa(count) + " " + pluralForm
}

// itoa renders a small non-negative integer without importing strconv for one call site.
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[position:])
}
