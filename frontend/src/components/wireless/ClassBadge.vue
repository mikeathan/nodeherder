<script setup lang="ts">
  /*
   * Class badge for the wireless viewer (spec 065).
   *
   * Two rules make this honest:
   *
   *  1. A provisional class is worded and marked as a guess. A first-day pre-filter result that
   *     looked like a verdict would teach the operator to distrust every class.
   *  2. A learning class shows its progress instead of a label. "Learning — 6 of 24 hours" reads as
   *     work in progress, where a bare "Learning" reads as a feature that gave up.
   *
   * The class is always advisory, so the badge is never styled as an alert: it is evidence, and the
   * triage controls beside it are the operator's override.
   */
  import { computed } from 'vue';
  import type { ObservationClassification } from '@/types/observations.type';
  import {
    classPresentation,
    classificationEvidence,
    isProvisional,
    learningProgress,
  } from '@/domain/observations/presentation';

  const props = defineProps<{
    classification: ObservationClassification | undefined;
    /** Compact drops the hint, for dense rows where it would wrap. */
    compact?: boolean;
  }>();

  const presentation = computed(() => classPresentation(props.classification));
  const provisional = computed(() => isProvisional(props.classification));
  const progress = computed(() => learningProgress(props.classification));
  const evidence = computed(() => classificationEvidence(props.classification));

  /** Tooltip text, so the reasoning is reachable without expanding the row. */
  const title = computed(() => {
    const parts: string[] = [presentation.value.hint];
    if (provisional.value) parts.push('Provisional: this is a guess, not a verdict.');
    if (props.classification?.changePending) {
      parts.push(`Changing to ${props.classification.candidate ?? 'another class'} as evidence accumulates.`);
    }
    if (evidence.value) parts.push(evidence.value);
    return parts.join(' ');
  });
</script>

<template>
  <span
    class="class-badge"
    :class="[`class-badge--${presentation.tone}`, { 'class-badge--provisional': provisional }]"
    :title="title"
    :data-class="classification?.class ?? 'unknown'"
  >
    <span class="class-badge__label">{{ presentation.label }}</span>
    <span v-if="provisional" class="class-badge__marker" aria-hidden="true">?</span>
    <span v-if="provisional" class="sr-only">(provisional guess)</span>
    <span v-if="progress" class="class-badge__progress">{{ progress }}</span>
    <span v-if="classification?.changePending" class="class-badge__pending" aria-hidden="true">→</span>
    <span v-if="classification?.changePending" class="sr-only">
      Changing to {{ classification.candidate ?? 'another class' }}
    </span>
  </span>
  <p v-if="!compact && evidence" class="class-badge__evidence">{{ evidence }}</p>
</template>

<style scoped>
  .class-badge {
    display: inline-flex;
    align-items: baseline;
    gap: 0.25rem;
    padding: 0.1rem 0.45rem;
    border-radius: 0.75rem;
    font-size: 0.75rem;
    font-weight: 600;
    line-height: 1.4;
    white-space: nowrap;
    border: 1px solid transparent;
  }

  /*
   * Colour is never the only cue: every tone also differs in label text, and provisional adds a
   * marker. A colour-blind operator still sees "Installed" and "?".
   */
  .class-badge--installed {
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    border-color: var(--nh-border);
  }

  .class-badge--yours {
    background: var(--nh-accent-soft);
    color: var(--nh-accent);
    border-color: var(--nh-accent);
  }

  .class-badge--outside {
    background: transparent;
    color: var(--nh-text-muted);
    border-color: var(--nh-border);
    border-style: dashed;
  }

  .class-badge--passing {
    background: transparent;
    color: var(--nh-text-muted);
    border-color: transparent;
    font-style: italic;
  }

  .class-badge--learning {
    background: transparent;
    color: var(--nh-text-muted);
    border-color: var(--nh-border);
    font-weight: 500;
  }

  /* A provisional class is visually lighter: a guess should not carry a verdict's weight. */
  .class-badge--provisional {
    opacity: 0.85;
    border-style: dotted;
  }

  .class-badge__marker {
    font-size: 0.65rem;
    opacity: 0.7;
  }

  .class-badge__progress {
    font-weight: 400;
    opacity: 0.8;
  }

  .class-badge__pending {
    font-weight: 400;
    opacity: 0.7;
  }

  .class-badge__evidence {
    margin: 0.15rem 0 0;
    font-size: 0.7rem;
    line-height: 1.3;
    color: var(--nh-text-muted);
    white-space: normal;
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>
