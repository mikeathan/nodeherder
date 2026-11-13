# ADR-002: Rule-based classification, not a learned model

Status: proposed | Date: 2026-11-12
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: none recorded |
Rules: NH-01, NH-04, NH-05

## Context and options

Four behaviour classes must be derived from evidence: infrastructure, personal, neighbour,
transient. Published work does apply supervised learning to BLE re-identification
([Learning to Link](https://arxiv.org/html/2609.26079v2), arXiv 2026), so a model is a
technically available option.

Options:

1. **Learned classifier.** Higher ceiling on separable classes, but requires labelled
   training data. There is none, and producing it means asking the operator to register and
   label their devices, which the maintainer has explicitly excluded. A model on unlabelled
   data would be guessing with extra steps.
2. **Published fingerprinting pipeline.** The literature targets re-identifying a device
   across address changes, not classifying behaviour into four coarse buckets. It solves a
   different problem and needs RF-level features (sequence numbers, information-element
   ordering) that the ESPHome advertisement message does not carry.
3. **Weighted rule scoring (chosen).** A bounded score per class from interpretable features,
   highest total wins, confidence from the margin.

## Decision and consequences

Choose **option 3**.

Consequences:

- **Explainable.** The UI can state why: "present 96% of the day, never moved". A model could
  not, and a class the operator cannot interrogate is one they will not trust.
- **Debuggable.** A wrong call is a constant to adjust, not a model to retrain.
- **No dependency.** NFR-03 holds.
- **Tunable without migration**, because the class is never persisted (ADR-001).
- Lower ceiling on genuinely ambiguous cases, accepted: a phone on a charger will read as
  infrastructure, and a television switched off at night will not read as 100% duty cycle.

Two safeguards are part of the decision, not tuning details:

- **Recency weighting** (5-day half-life) so a device that changes behaviour can change class.
- **Hysteresis:** a class change requires the candidate to win on two consecutive
  evaluations, so one unusual day cannot flip a long-standing class.

A fixed feature such as "seen at 04:00" is evidence with weight, never a verdict. The
maintainer rejected treating it as a label because a device that later moves must be able to
change class.

Verification: table-driven tests with synthetic evidence per class; specific tests for
recency weighting, hysteresis, and the learning threshold.

## Supersession

None. Reconsider if labelled data ever becomes available without device registration.
