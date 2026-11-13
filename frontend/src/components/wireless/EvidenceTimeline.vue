<script setup lang="ts">
  /*
   * Evidence timeline for the isolated identity (spec 065).
   *
   * This is the pane that answers "why does it think that?". It shows each stored day as a bar for
   * the covered hours, so the shape of a device's life is visible at a glance: a solid bar every
   * day is installed equipment, a daytime-only bar is usually a person's device.
   *
   * It is deliberately not a chart library. One row per day with an hour strip carries more meaning
   * than any line plot of a mean, because the hour mask is the evidence and the mean is a
   * summary of it.
   */
  import { computed, ref, watch } from 'vue';
  import {
    loadObservationHistory,
    type ObservationDaySummary,
    type ObservationHistory,
    type ObservationTransition,
  } from '@/services/observation-history.service';
  import { classPresentation } from '@/domain/observations/presentation';

  const props = defineProps<{
    /** Identity whose evidence to show. */
    identityKey: string;
  }>();

  const history = ref<ObservationHistory | null>(null);
  const loading = ref(false);
  const error = ref('');

  /**
   * Load history whenever the identity changes.
   *
   * A failed load leaves any previous result in place rather than clearing it: showing stale
   * evidence is less confusing than a pane that empties itself on a transient network error.
   */
  watch(
    () => props.identityKey,
    async (key) => {
      if (!key) return;
      loading.value = true;
      error.value = '';
      try {
        history.value = await loadObservationHistory(key);
      } catch (failure) {
        error.value = failure instanceof Error ? failure.message : 'History could not be loaded.';
      } finally {
        loading.value = false;
      }
    },
    { immediate: true },
  );

  /** Days newest first, which is the order the backend already returns. */
  const days = computed<ObservationDaySummary[]>(() => history.value?.summaries ?? []);
  const transitions = computed<ObservationTransition[]>(() => history.value?.transitions ?? []);

  /** True when the store answered but there is genuinely nothing yet. */
  const empty = computed(() => !loading.value && !error.value && days.value.length === 0);

  /** The hours of the day covered, as 24 slots, for the strip. */
  function hoursOf(mask: number): boolean[] {
    return Array.from({ length: 24 }, (_, hour) => (mask & (1 << hour)) !== 0);
  }

  /** Short day label, e.g. "Thu 13". */
  function dayLabel(day: string): string {
    const parsed = new Date(`${day}T00:00:00Z`);
    if (Number.isNaN(parsed.getTime())) return day;
    return parsed.toLocaleDateString(undefined, { weekday: 'short', day: 'numeric', timeZone: 'UTC' });
  }

  /** Whether a day contains the quiet period, which is the strongest installed-equipment sign. */
  function coversNight(mask: number): boolean {
    for (let hour = 1; hour < 5; hour++) {
      if ((mask & (1 << hour)) !== 0) return true;
    }
    return false;
  }

  /** Plain-language description of one stored day. */
  function describeDay(day: ObservationDaySummary): string {
    const parts = [`Seen in ${day.activeHours} of 24 hours`];
    if (day.crossings > 0) parts.push(`${day.crossings} signal changes`);
    if (coversNight(day.hourMask)) parts.push('including overnight');
    return parts.join(', ') + '.';
  }

  /** Class label for a transition, in the same words the badge uses. */
  function transitionLabel(value: string): string {
    return classPresentation({ class: value, showByDefault: true, confidence: 0, hoursObserved: 0 }).label;
  }
</script>

<template>
  <section class="evidence-timeline" aria-labelledby="evidence-timeline-title">
    <h3 id="evidence-timeline-title" class="evidence-timeline__title">Evidence</h3>

    <p v-if="loading" class="evidence-timeline__note" role="status">Loading history…</p>

    <p v-else-if="error" class="evidence-timeline__error" role="alert">{{ error }}</p>

    <p v-else-if="empty" class="evidence-timeline__note">
      No stored days yet. A day is kept once a device has been heard in at least two different hours
      with at least twenty packets, so passing devices leave no record.
    </p>

    <template v-else>
      <!--
        One row per day. The hour strip is the evidence itself rather than a summary of it, so an
        operator can see that "present 20 hours" means "present at 3am" or "not".
      -->
      <ol class="evidence-timeline__days">
        <li v-for="day in days" :key="day.day" class="evidence-timeline__day">
          <span class="evidence-timeline__date">{{ dayLabel(day.day) }}</span>
          <span class="evidence-timeline__strip" role="img" :aria-label="describeDay(day)">
            <span
              v-for="(present, hour) in hoursOf(day.hourMask)"
              :key="hour"
              class="evidence-timeline__hour"
              :class="{ 'is-present': present, 'is-night': hour >= 1 && hour < 5 }"
            />
          </span>
          <span class="evidence-timeline__figure">{{ day.activeHours }}/24 h</span>
          <span class="evidence-timeline__signal">{{ Math.round(day.rssiMean) }} dBm</span>
        </li>
      </ol>

      <p class="evidence-timeline__legend">
        <span class="evidence-timeline__hour is-present" aria-hidden="true" /> heard ·
        <span class="evidence-timeline__hour is-night" aria-hidden="true" /> overnight window
      </p>

      <!--
        Transitions are shown with their reason because that reason is all that remains once the
        summaries behind it are pruned. A change with no explanation would be worse than no change.
      -->
      <div v-if="transitions.length > 0" class="evidence-timeline__transitions">
        <h4 class="evidence-timeline__subtitle">Changes</h4>
        <ul>
          <li v-for="transition in transitions" :key="`${transition.day}-${transition.at}`">
            <span class="evidence-timeline__change">
              {{ transitionLabel(transition.from) }} → {{ transitionLabel(transition.to) }}
            </span>
            <span class="evidence-timeline__when">{{ dayLabel(transition.day) }}</span>
            <span v-if="transition.reason" class="evidence-timeline__reason">{{ transition.reason }}</span>
          </li>
        </ul>
      </div>
    </template>
  </section>
</template>

<style scoped>
  .evidence-timeline {
    margin-top: 1rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--border-subtle, #dee2e6);
    max-width: 34rem;
    width: 100%;
  }

  .evidence-timeline__title {
    margin: 0 0 0.5rem;
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text-muted, #495057);
  }

  .evidence-timeline__note,
  .evidence-timeline__error {
    margin: 0;
    font-size: 0.75rem;
    line-height: 1.4;
    color: var(--text-subtle, #868e96);
  }

  .evidence-timeline__error {
    color: var(--danger, #c92a2a);
  }

  .evidence-timeline__days {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }

  .evidence-timeline__day {
    display: grid;
    grid-template-columns: 3.2rem 1fr 3.2rem 3.6rem;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.7rem;
  }

  .evidence-timeline__date {
    color: var(--text-muted, #495057);
    white-space: nowrap;
  }

  .evidence-timeline__strip {
    display: grid;
    grid-template-columns: repeat(24, 1fr);
    gap: 1px;
    height: 0.6rem;
  }

  .evidence-timeline__hour {
    background: var(--surface-raised, #f1f3f5);
    border-radius: 1px;
  }

  .evidence-timeline__hour.is-present {
    background: var(--accent-strong, #1864ab);
  }

  /* The night window is marked by outline, not a second colour, so it survives greyscale. */
  .evidence-timeline__hour.is-night {
    outline: 1px dotted var(--text-subtle, #868e96);
    outline-offset: -1px;
  }

  .evidence-timeline__figure,
  .evidence-timeline__signal {
    text-align: right;
    color: var(--text-subtle, #868e96);
    font-variant-numeric: tabular-nums;
  }

  .evidence-timeline__legend {
    margin: 0.4rem 0 0;
    font-size: 0.68rem;
    color: var(--text-subtle, #868e96);
  }

  .evidence-timeline__legend .evidence-timeline__hour {
    display: inline-block;
    width: 0.55rem;
    height: 0.55rem;
    vertical-align: middle;
  }

  .evidence-timeline__subtitle {
    margin: 0.75rem 0 0.3rem;
    font-size: 0.78rem;
    font-weight: 600;
    color: var(--text-muted, #495057);
  }

  .evidence-timeline__transitions ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .evidence-timeline__change {
    font-weight: 600;
    font-size: 0.72rem;
    margin-right: 0.35rem;
  }

  .evidence-timeline__when {
    font-size: 0.7rem;
    color: var(--text-subtle, #868e96);
  }

  .evidence-timeline__reason {
    display: block;
    font-size: 0.7rem;
    line-height: 1.35;
    color: var(--text-subtle, #868e96);
  }

  /*
   * On a phone the four-column grid squeezes the strip to nothing, so the figures drop below it
   * and the strip keeps the full width. The strip is the part that carries the meaning.
   */
  @media (max-width: 480px) {
    .evidence-timeline__day {
      grid-template-columns: 3rem 1fr;
      grid-template-areas:
        'date strip'
        'figure signal';
      row-gap: 0.1rem;
    }

    .evidence-timeline__date {
      grid-area: date;
    }

    .evidence-timeline__strip {
      grid-area: strip;
    }

    .evidence-timeline__figure {
      grid-area: figure;
      text-align: left;
    }

    .evidence-timeline__signal {
      grid-area: signal;
    }
  }
</style>
