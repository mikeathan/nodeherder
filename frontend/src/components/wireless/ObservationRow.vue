<script setup lang="ts">
  /*
   * One wireless identity, rendered identically in the radar side panel and the List view.
   *
   * A single row component exists because the two surfaces used to invent their own anatomy for
   * the same data: a three-line card on one side and a dense vendor-grouped row on the other. The
   * operator should not have to relearn a device when they switch view, and one component cannot
   * drift from itself.
   *
   * Presentational only: distance, trend, seen time and the title come from the observations
   * domain, not from computations in this template.
   */
  import { computed } from 'vue';
  import type { ObservationDecision, ObservationIdentityView } from '@/types/observations.type';
  import { ObservationDecisions } from '@/types/observations.type';
  import {
    distanceRSSI,
    estimateDistanceMetres,
    identityTitle,
    isEphemeralIdentity,
    roundedDistanceMetres,
    seenLabel,
    trendPresentation,
  } from '@/domain/observations/presentation';
  import ClassBadge from './ClassBadge.vue';
  import SignalSparkline from './SignalSparkline.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    identity: ObservationIdentityView;
    /** Whether a triage save is in flight for this identity. */
    pending: boolean;
    /** Whether this identity is currently isolated. */
    selected: boolean;
    /** Current time in milliseconds, so the relative label is reactive and testable. */
    now: number;
    /** Source id to display name, so the row can say which receiver heard it. */
    sourceNames: ReadonlyMap<string, string>;
  }>();

  const emit = defineEmits<{
    (event: 'decide', key: string, decision: ObservationDecision): void;
    (event: 'select', key: string): void;
  }>();

  const view = computed(() => {
    const metres = estimateDistanceMetres(distanceRSSI(props.identity));
    const rounded = roundedDistanceMetres(metres);
    return {
      title: identityTitle(props.identity),
      // A device with no usable reading gets a dash, never "0 m", which would claim it is on top
      // of the antenna.
      distance: rounded === null ? '—' : `${rounded} m`,
      trend: trendPresentation(props.identity.trend),
      ephemeral: isEphemeralIdentity(props.identity),
      seen: seenLabel(props.identity.lastSeen, props.now),
    };
  });

  /** A sparkline needs at least two readings; otherwise the column would be empty chrome. */
  const showSpark = computed(() => (props.identity.signal?.length ?? 0) > 1);

  /** Which receiving sources heard this identity, by name. */
  const via = computed(() => {
    const names = new Set<string>();
    for (const reading of props.identity.sources ?? []) {
      const name = props.sourceNames.get(reading.id);
      if (name) names.add(name);
    }
    return [...names].join(' · ');
  });

  /**
   * Only name the source when there is a choice to make. With a single receiver every row would
   * repeat the same "via …" line, which is noise rather than information.
   */
  const showVia = computed(() => props.sourceNames.size > 1 && via.value !== '');
</script>

<template>
  <li class="observation-row" :class="{ 'is-selected': props.selected }">
    <!--
      The row body is the isolate control. A separate link inside the row would compete with the
      triage buttons for taps, and on a phone the row is the natural target.
    -->
    <button
      type="button"
      class="observation-row__main"
      :aria-pressed="props.selected"
      :aria-label="`Isolate ${view.title}, ${view.distance}`"
      @click="emit('select', props.identity.key)"
    >
      <span class="observation-row__identity">
        <span class="observation-row__name">{{ view.title }}</span>
        <ClassBadge :classification="props.identity.classification" compact />
        <span v-if="view.ephemeral" class="observation-row__tag">rotating</span>
      </span>
      <span class="observation-row__meta">
        <span class="observation-row__distance">{{ view.distance }}</span>
        <span class="observation-row__trend">
          <span v-if="view.trend.arrow" aria-hidden="true">{{ view.trend.arrow }}</span>
          {{ view.trend.label }}
        </span>
        <span v-if="showVia" class="observation-row__via">via {{ via }}</span>
        <span class="observation-row__seen">{{ view.seen }}</span>
      </span>
    </button>

    <span v-if="showSpark" class="observation-row__spark">
      <SignalSparkline :signal="props.identity.signal" />
    </span>

    <!--
      Triage is only offered while it can still change something. A control whose action is a
      no-op teaches the operator that the controls do not matter.
    -->
    <span class="observation-row__actions" role="group" :aria-label="`Triage for ${view.title}`">
      <UiButton
        v-if="props.identity.decision !== ObservationDecisions.Known"
        :disabled="props.pending"
        size="sm"
        variant="default"
        @click="emit('decide', props.identity.key, ObservationDecisions.Known)"
      >
        Known
      </UiButton>
      <UiButton
        v-if="props.identity.decision !== ObservationDecisions.Ignored"
        :disabled="props.pending"
        size="sm"
        variant="ghost"
        @click="emit('decide', props.identity.key, ObservationDecisions.Ignored)"
      >
        Ignore
      </UiButton>
      <span v-if="props.pending" class="observation-row__pending" role="status">Saving…</span>
    </span>
  </li>
</template>

<style scoped>
  .observation-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto auto;
    align-items: center;
    gap: 0.75rem;
    padding: 0.45rem 0.6rem;
    border-bottom: 1px solid color-mix(in srgb, var(--nh-border) 55%, transparent);
  }

  .observation-row.is-selected {
    background: color-mix(in srgb, var(--nh-accent) 12%, transparent);
  }

  .observation-row__main {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    min-width: 0;
    background: none;
    border: 0;
    color: inherit;
    font: inherit;
    padding: 0;
    text-align: left;
    cursor: pointer;
  }

  .observation-row__main:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: 2px;
    border-radius: 0.25rem;
  }

  .observation-row__identity {
    display: flex;
    align-items: baseline;
    gap: 0.4rem;
    min-width: 0;
  }

  .observation-row__name {
    font-weight: 600;
    font-size: 0.85rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .observation-row__tag {
    padding: 0 0.3rem;
    border: 1px dashed var(--nh-border);
    border-radius: 0.6rem;
    font-size: 0.65rem;
  }

  .observation-row__meta {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.15rem 0.6rem;
    font-size: 0.7rem;
    color: var(--nh-text-muted);
  }

  /* Distance is the number the operator came for, so it leads the metadata line. */
  .observation-row__distance {
    font-variant-numeric: tabular-nums;
    color: var(--nh-text-muted);
  }

  .observation-row__via {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .observation-row__seen {
    margin-left: auto;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .observation-row__spark {
    display: inline-flex;
    align-items: center;
  }

  .observation-row__actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.3rem;
  }

  .observation-row__pending {
    font-size: 0.7rem;
    color: var(--nh-text-muted);
  }

  /* On a phone the two columns cannot share 390px, so the actions move under the identity. */
  @media (max-width: 599px) {
    .observation-row {
      grid-template-columns: 1fr;
      gap: 0.25rem;
      padding: 0.5rem 0.6rem;
    }

    .observation-row__spark {
      display: none;
    }

    .observation-row__actions {
      justify-content: flex-start;
    }
  }
</style>
