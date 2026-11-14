<script setup lang="ts">
  /*
   * Detail panel for the isolated identity.
   *
   * This is no longer a screen. It used to replace the whole view with its own mini radar and its
   * own ring maths, which meant selecting a device threw away the radar the operator was reading —
   * and the two radars could disagree because they were separate implementations.
   *
   * Now it is the panel beside the shared radar, which stays visible and shows the selection.
   */
  import { computed } from 'vue';
  import type { ObservationDecision, ObservationIdentityView } from '@/types/observations.type';
  import { ObservationDecisions } from '@/types/observations.type';
  import {
    DISTANCE_BAND_LABELS,
    distanceBand,
    distanceRSSI,
    estimateDistanceMetres,
    identityTitle,
    isEphemeralIdentity,
    isProvisional,
    roundedDistanceMetres,
    trendPresentation,
  } from '@/domain/observations/presentation';
  import ClassBadge from './ClassBadge.vue';
  import EvidenceTimeline from './EvidenceTimeline.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    identity: ObservationIdentityView;
    pending: boolean;
  }>();

  const emit = defineEmits<{
    (event: 'decide', key: string, decision: ObservationDecision): void;
    (event: 'clear'): void;
  }>();

  const metres = computed(() => estimateDistanceMetres(distanceRSSI(props.identity)));
  const rounded = computed(() => roundedDistanceMetres(metres.value));
  const distanceText = computed(() => (rounded.value === null ? '—' : `${rounded.value} m`));
  const band = computed(() => distanceBand(metres.value));
  const trend = computed(() => trendPresentation(props.identity.trend));
  const provisional = computed(() => isProvisional(props.identity.classification));
  const isUntriaged = computed(() => props.identity.decision === ObservationDecisions.Unknown);

  /**
   * The class reasoning, shown once.
   *
   * It previously appeared twice in a row — once inside the badge component's own evidence line
   * and once as a paragraph — so the panel said the same sentence back to back.
   */
  const reason = computed(() => props.identity.classification?.reason ?? '');

  const sources = computed(() => props.identity.sources);

  function time(value: string): string {
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return '—';
    return parsed.toLocaleTimeString();
  }

  /** Absent readings must not render as 0 dBm, which would read as a strong signal. */
  function signal(rssi: number | null | undefined): string {
    if (typeof rssi !== 'number' || !Number.isFinite(rssi)) return '—';
    return `${rssi} dBm`;
  }

  /** Sparkline geometry, computed here rather than in the template. */
  const spark = computed(() => {
    const samples = props.identity.signal ?? [];
    if (samples.length < 2) return null;
    const min = Math.min(...samples);
    const max = Math.max(...samples);
    const span = max - min || 1;
    const points = samples
      .map((value, index) => {
        const x = (index / (samples.length - 1)) * 100;
        const y = 100 - ((value - min) / span) * 100;
        return `${x.toFixed(2)},${y.toFixed(2)}`;
      })
      .join(' ');
    return { points, min, max };
  });
</script>

<template>
  <aside class="observation-focus" aria-labelledby="observation-focus-title">
    <header class="observation-focus__header">
      <UiButton size="sm" variant="ghost" @click="emit('clear')">‹ All devices</UiButton>
      <span v-if="isEphemeralIdentity(props.identity)" class="observation-focus__tag">rotating</span>
    </header>

    <h2 id="observation-focus-title" class="observation-focus__title">
      {{ identityTitle(props.identity) }}
    </h2>
    <p class="observation-focus__key">{{ props.identity.key }}</p>

    <!-- Distance is the headline figure: it states numerically what the radar shows as a radius,
         so the operator does not have to estimate from the blip's position. -->
    <p class="observation-focus__distance">
      <span class="observation-focus__value">{{ distanceText }}</span>
      <span class="observation-focus__band">{{ DISTANCE_BAND_LABELS[band] }}</span>
    </p>
    <p class="observation-focus__trend" role="status">
      <span v-if="trend.arrow" class="observation-focus__arrow" aria-hidden="true">{{ trend.arrow }}</span>
      {{ trend.label }}
    </p>

    <!--
      The class and its reasoning. A bare badge would be useless here: this is the panel the
      operator opened to judge the classification, so it carries the evidence and the caveat.
    -->
    <section v-if="props.identity.classification" class="observation-focus__class">
      <ClassBadge :classification="props.identity.classification" compact />
      <p v-if="reason" class="observation-focus__reason">{{ reason }}</p>
      <p v-if="provisional" class="observation-focus__note">
        A guess from the first packet, not a verdict.
      </p>
      <p class="observation-focus__note">Advisory only — your own decision overrules it.</p>
    </section>

    <figure v-if="spark" class="observation-focus__spark">
      <svg viewBox="0 0 100 30" preserveAspectRatio="none" role="img" aria-label="Recent signal strength">
        <polyline class="observation-focus__spark-line" :points="spark.points" />
      </svg>
      <figcaption>
        Recent signal strength, {{ spark.min }} to {{ spark.max }} dBm. Estimated from signal
        strength only, accurate to a few metres indoors.
      </figcaption>
    </figure>

    <p v-else class="observation-focus__note">
      Signal history builds as advertisements arrive. A direction needs a few readings.
    </p>

    <!-- The measurements a card cannot show, all in one place. -->
    <dl class="observation-focus__facts">
      <div>
        <dt>Signal</dt>
        <dd>{{ signal(props.identity.rssi) }}</dd>
      </div>
      <div>
        <dt>Smoothed</dt>
        <dd>{{ signal(props.identity.rssiSmoothed) }}</dd>
      </div>
      <div>
        <dt>Packets</dt>
        <dd>{{ props.identity.packets }}</dd>
      </div>
      <div>
        <dt>Last seen</dt>
        <dd>{{ time(props.identity.lastSeen) }}</dd>
      </div>
      <div v-for="reading in sources" :key="reading.id">
        <dt>From {{ reading.id }}</dt>
        <dd>{{ signal(reading.rssi) }}</dd>
      </div>
    </dl>

    <EvidenceTimeline :identity-key="props.identity.key" />

    <div class="observation-focus__actions">
      <UiButton
        v-if="isUntriaged"
        size="sm"
        variant="primary"
        :disabled="props.pending"
        @click="emit('decide', props.identity.key, ObservationDecisions.Known)"
      >
        Mark known
      </UiButton>
      <UiButton
        v-if="props.identity.decision !== ObservationDecisions.Ignored"
        size="sm"
        variant="ghost"
        :disabled="props.pending"
        @click="emit('decide', props.identity.key, ObservationDecisions.Ignored)"
      >
        Ignore
      </UiButton>
      <span v-if="props.pending" class="observation-focus__pending" role="status">Saving…</span>
    </div>
  </aside>
</template>

<style scoped>
  .observation-focus {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid var(--border-subtle, #2a3441);
    border-radius: 0.5rem;
    background: var(--surface-raised, rgba(255, 255, 255, 0.02));
    /* A panel, not a page: it never grows wider than its column. */
    min-width: 0;
    max-height: 70vh;
    overflow-y: auto;
  }

  .observation-focus__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .observation-focus__tag {
    padding: 0 0.35rem;
    border: 1px dashed var(--border-subtle, #2a3441);
    border-radius: 0.6rem;
    font-size: 0.65rem;
    color: var(--text-subtle, #868e96);
  }

  .observation-focus__title {
    margin: 0;
    font-size: 1.05rem;
    line-height: 1.25;
  }

  .observation-focus__key {
    margin: 0;
    font-size: 0.68rem;
    color: var(--text-subtle, #868e96);
    overflow-wrap: anywhere;
  }

  .observation-focus__distance {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    margin: 0.25rem 0 0;
  }

  .observation-focus__value {
    font-size: 1.75rem;
    font-weight: 700;
    line-height: 1;
    font-variant-numeric: tabular-nums;
  }

  .observation-focus__band {
    font-size: 0.75rem;
    color: var(--text-subtle, #868e96);
  }

  .observation-focus__trend {
    margin: 0;
    font-size: 0.78rem;
    color: var(--text-muted, #495057);
  }

  .observation-focus__class {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.25rem;
    padding-top: 0.4rem;
    border-top: 1px solid var(--border-subtle, #2a3441);
  }

  .observation-focus__reason {
    margin: 0;
    font-size: 0.78rem;
    line-height: 1.4;
    color: var(--text-muted, #495057);
  }

  .observation-focus__note {
    margin: 0;
    font-size: 0.7rem;
    line-height: 1.35;
    color: var(--text-subtle, #868e96);
  }

  .observation-focus__spark {
    margin: 0.25rem 0 0;
  }

  .observation-focus__spark svg {
    display: block;
    width: 100%;
    height: 2.5rem;
  }

  .observation-focus__spark-line {
    fill: none;
    stroke: var(--accent-primary, #4a9eff);
    stroke-width: 1.5;
    vector-effect: non-scaling-stroke;
  }

  .observation-focus__spark figcaption {
    margin-top: 0.2rem;
    font-size: 0.68rem;
    line-height: 1.35;
    color: var(--text-subtle, #868e96);
  }

  .observation-focus__facts {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.3rem 0.75rem;
    margin: 0.4rem 0 0;
    padding-top: 0.5rem;
    border-top: 1px solid var(--border-subtle, #2a3441);
    font-size: 0.72rem;
  }

  .observation-focus__facts dt {
    color: var(--text-subtle, #868e96);
  }

  .observation-focus__facts dd {
    margin: 0;
    font-variant-numeric: tabular-nums;
  }

  .observation-focus__actions {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    padding-top: 0.5rem;
    border-top: 1px solid var(--border-subtle, #2a3441);
  }

  .observation-focus__pending {
    font-size: 0.72rem;
    color: var(--text-subtle, #868e96);
  }
</style>
