<script setup lang="ts">
  /*
   * Device cards for the wireless viewer.
   *
   * Replaces the eight-column table as the default list. A table was the wrong shape for this data:
   * almost every row is the same kind, the same address type and the same vendor, so seven of the
   * eight columns carried no information — the operator described the result as "a big list of
   * mostly mac address", and a table makes that worse by giving each identical field a column of
   * its own.
   *
   * A card carries what actually differs between devices: what it is (class), what it is called,
   * how far away it is, and whether it is moving. The technical fields still matter, but they
   * belong in the isolated view where there is room to read them.
   */
  import { computed } from 'vue';
  import type { ObservationDecision, ObservationIdentityView } from '@/types/observations.type';
  import { ObservationDecisions } from '@/types/observations.type';
  import {
    classPresentation,
    distanceRSSI,
    estimateDistanceMetres,
    identityTitle,
    isEphemeralIdentity,
    isProvisional,
    roundedDistanceMetres,
    trendPresentation,
  } from '@/domain/observations/presentation';
  import ClassBadge from './ClassBadge.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    identities: readonly ObservationIdentityView[];
    /** Identity keys with a save in flight, so their controls report progress. */
    pending: ReadonlySet<string>;
    /** Currently isolated identity, so the card can be marked. */
    selectedKey?: string | null;
  }>();

  const emit = defineEmits<{
    (event: 'decide', key: string, decision: ObservationDecision): void;
    (event: 'select', key: string): void;
  }>();

  /** Row view models, computed once rather than per cell in the template. */
  const cards = computed(() =>
    props.identities.map((identity) => {
      const metres = estimateDistanceMetres(distanceRSSI(identity));
      const rounded = roundedDistanceMetres(metres);
      return {
        identity,
        title: identityTitle(identity),
        presentation: classPresentation(identity.classification),
        provisional: isProvisional(identity.classification),
        // A device with no usable reading gets a dash, never "0 m", which would claim it is on top
        // of the antenna.
        distance: rounded === null ? '—' : `${rounded} m`,
        trend: trendPresentation(identity.trend),
        ephemeral: isEphemeralIdentity(identity),
        untriaged: identity.decision === ObservationDecisions.Unknown,
      };
    }),
  );

  function isPending(key: string): boolean {
    return props.pending.has(key);
  }

  /** Relative time, because "3 min ago" answers the question a timestamp makes you compute. */
  function since(value: string): string {
    const parsed = Date.parse(value);
    if (Number.isNaN(parsed)) return '—';
    const seconds = Math.max(0, Math.round((Date.now() - parsed) / 1000));
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.round(seconds / 60);
    if (minutes < 60) return `${minutes} min ago`;
    const hours = Math.round(minutes / 60);
    return `${hours} h ago`;
  }
</script>

<template>
  <div class="observ-cards">
    <ul v-if="cards.length > 0" class="device-cards__list">
      <li
        v-for="card in cards"
        :key="card.identity.key"
        class="observ-card"
        :class="{ 'is-selected': props.selectedKey === card.identity.key }"
      >
        <!--
          The whole card is the isolate control. A separate link inside the card would compete with
          the triage buttons for taps, and on a phone the card is the natural target.
        -->
        <button
          type="button"
          class="observ-card__main"
          :aria-pressed="props.selectedKey === card.identity.key"
          :aria-label="`Isolate ${card.title}, ${card.distance}`"
          @click="emit('select', card.identity.key)"
        >
          <span class="observ-card__line">
            <span class="observ-card__name">{{ card.title }}</span>
            <span class="observ-card__distance">{{ card.distance }}</span>
          </span>

          <span class="observ-card__line observ-card__line--meta">
            <ClassBadge :classification="card.identity.classification" compact />
            <span v-if="card.ephemeral" class="observ-card__tag">rotating</span>
            <span class="observ-card__trend">
              <span v-if="card.trend.arrow" aria-hidden="true">{{ card.trend.arrow }}</span>
              {{ card.trend.label }}
            </span>
            <span class="observ-card__when">{{ since(card.identity.lastSeen) }}</span>
          </span>
        </button>

        <!--
          Triage is only offered while it can still change something. An already-known device shows
          no "Mark known" button: a control whose action is a no-op teaches the operator that the
          controls do not matter.
        -->
        <div
          v-if="card.untriaged || card.identity.decision !== ObservationDecisions.Ignored"
          class="observ-card__actions"
          role="group"
          :aria-label="`Triage for ${card.title}`"
        >
          <UiButton
            v-if="card.identity.decision !== ObservationDecisions.Known"
            :disabled="isPending(card.identity.key)"
            size="sm"
            variant="default"
            @click="emit('decide', card.identity.key, ObservationDecisions.Known)"
          >
            Known
          </UiButton>
          <UiButton
            v-if="card.identity.decision !== ObservationDecisions.Ignored"
            :disabled="isPending(card.identity.key)"
            size="sm"
            variant="ghost"
            @click="emit('decide', card.identity.key, ObservationDecisions.Ignored)"
          >
            Ignore
          </UiButton>
          <span v-if="isPending(card.identity.key)" class="observ-card__pending" role="status">Saving…</span>
        </div>
      </li>
    </ul>

    <p v-else class="device-cards__empty">
      Nothing to show. Advertisements appear here as sources report them.
    </p>
  </div>
</template>

<style scoped>
  .observ-cards__list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  .observ-card {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    padding: 0.5rem 0.625rem;
    border: 1px solid var(--border-subtle, #2a3441);
    border-radius: 0.5rem;
    background: var(--surface, transparent);
    /* The left edge carries the class tone, so a column of cards can be scanned without reading. */
    border-left-width: 3px;
  }

  .observ-card.is-selected {
    background: color-mix(in srgb, var(--accent-primary, #4a9eff) 12%, transparent);
    border-color: var(--accent-primary, #4a9eff);
  }

  .observ-card__main {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    width: 100%;
    background: none;
    border: 0;
    color: inherit;
    font: inherit;
    padding: 0;
    text-align: left;
    cursor: pointer;
  }

  .observ-card__main:focus-visible {
    outline: 2px solid var(--accent-primary, #4a9eff);
    outline-offset: 2px;
    border-radius: 0.25rem;
  }

  .observ-card__line {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    min-width: 0;
  }

  .observ-card__line--meta {
    flex-wrap: wrap;
    gap: 0.35rem 0.5rem;
    font-size: 0.72rem;
    color: var(--text-subtle, #868e96);
  }

  .observ-card__name {
    font-weight: 600;
    font-size: 0.85rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Distance is the number the operator came for, so it is fixed-width and right-aligned. */
  .observ-card__distance {
    margin-left: auto;
    font-variant-numeric: tabular-nums;
    font-size: 0.8rem;
    color: var(--text-muted, #495057);
    white-space: nowrap;
  }

  .observ-card__tag {
    padding: 0 0.3rem;
    border: 1px dashed var(--border-subtle, #2a3441);
    border-radius: 0.6rem;
    font-size: 0.65rem;
  }

  .observ-card__when {
    margin-left: auto;
    font-variant-numeric: tabular-nums;
  }

  .observ-card__actions {
    display: flex;
    align-items: center;
    gap: 0.35rem;
  }

  .observ-card__pending {
    font-size: 0.72rem;
    color: var(--text-subtle, #868e96);
  }

  .observ-cards__empty {
    margin: 0;
    padding: 1rem 0.75rem;
    color: var(--text-muted, #8b9bab);
  }

  @media (max-width: 600px) {
    /* Touch targets: the card body and each button need a comfortable height. */
    .observ-card {
      padding: 0.625rem 0.75rem;
    }

    .observ-card__main {
      gap: 0.3rem;
    }
  }
</style>
