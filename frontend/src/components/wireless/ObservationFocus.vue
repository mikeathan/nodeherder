<script setup lang="ts">
  /*
   * Focus view: one isolated identity with a large proximity readout (spec 064).
   *
   * A single anchor yields distance only, so this shows concentric range rings and never
   * implies a bearing. The large figure plus band plus trend is what makes it readable at arm's
   * length on a phone, which a single blip on a big radar canvas was not.
   */
  import { computed } from 'vue';
  import type { ObservationIdentityView } from '@/types/observations.type';
  import { ObservationDecisions } from '@/types/observations.type';
  import type { ObservationDecision } from '@/types/observations.type';
  import {
    DEFAULT_PATH_LOSS_EXPONENT,
    DEFAULT_TX_POWER_AT_1M,
    DISTANCE_BAND_LABELS,
    distanceBand,
    distanceRSSI,
    estimateDistanceMetres,
    roundedDistanceMetres,
    identityTitle,
    isEphemeralIdentity,
    trendPresentation,
  } from '@/domain/observations/presentation';
  import { formatQuantity } from '@/utils/units';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = withDefaults(
    defineProps<{
      identity: ObservationIdentityView;
      /** Whether the last save for this identity is still in flight. */
      pending: boolean;
      txPowerAt1m?: number;
      pathLossExponent?: number;
    }>(),
    {
      txPowerAt1m: DEFAULT_TX_POWER_AT_1M,
      pathLossExponent: DEFAULT_PATH_LOSS_EXPONENT,
    },
  );

  const emit = defineEmits<{
    (event: 'decide', key: string, decision: ObservationDecision): void;
    (event: 'clear'): void;
  }>();

  /**
   * Estimated range, or null when the source reported no signal strength. Derived from the
   * smoothed median rather than the newest packet, so the figure does not swing with jitter.
   */
  const metres = computed(() =>
    estimateDistanceMetres(distanceRSSI(props.identity), props.txPowerAt1m, props.pathLossExponent),
  );

  /** Distance rounded to the precision the estimate supports; sub-metre figures are fiction. */
  const displayMetres = computed(() => roundedDistanceMetres(metres.value));

  const band = computed(() => distanceBand(metres.value));

  /** Distance through the shared formatter, so the missing value is never rendered as zero. */
  const distanceText = computed(() => formatQuantity(displayMetres.value, 'distance'));

  const trend = computed(() => trendPresentation(props.identity.trend));

  /** Sparkline geometry for the recent signal history. */
  const spark = computed(() => {
    const samples = props.identity.signal ?? [];
    if (samples.length < 2) return null;
    const min = Math.min(...samples);
    const max = Math.max(...samples);
    // A flat history has no range; drawing it mid-height avoids a divide by zero and reads as
    // "unchanged" rather than as a spike.
    const span = max - min || 1;
    const points = samples
      .map((sample, index) => {
        const x = (index / (samples.length - 1)) * 100;
        const y = 30 - ((sample - min) / span) * 26 - 2;
        return `${x.toFixed(1)},${y.toFixed(1)}`;
      })
      .join(' ');
    return { points, min, max };
  });

  const isUntriaged = computed(() => props.identity.decision === ObservationDecisions.Unknown);
</script>

<template>
  <section class="observation-focus" aria-labelledby="observation-focus-title">
    <header class="observation-focus__header">
      <UiButton size="sm" variant="ghost" @click="emit('clear')">‹ All identities</UiButton>
      <span v-if="isEphemeralIdentity(props.identity)" class="observation-focus__badge">
        Rotating address
      </span>
    </header>

    <h2 id="observation-focus-title" class="observation-focus__title">
      {{ identityTitle(props.identity) }}
    </h2>
    <p class="observation-focus__key">{{ props.identity.key }}</p>

    <!--
      Range rings, deliberately without a bearing. The inner blip sits at a radius proportional
      to the estimate so the graphic and the figure cannot disagree.
    -->
    <div class="observation-focus__radar" aria-hidden="true">
      <svg viewBox="0 0 100 100" role="presentation">
        <circle v-for="ring in [25, 50, 75]" :key="ring" class="observation-focus__ring" cx="50" cy="50" :r="ring" />
        <circle class="observation-focus__centre" cx="50" cy="50" r="2" />
        <circle
          v-if="metres !== null"
          class="observation-focus__target"
          cx="50"
          :cy="50 - Math.min(metres / 10, 1) * 34"
          r="3.5"
        />
      </svg>
    </div>

    <p class="observation-focus__distance">
      <span class="observation-focus__value">{{ distanceText }}</span>
      <span class="observation-focus__band">{{ DISTANCE_BAND_LABELS[band] }}</span>
    </p>

    <p class="observation-focus__trend" role="status">
      <span v-if="trend.arrow" class="observation-focus__arrow" aria-hidden="true">{{ trend.arrow }}</span>
      {{ trend.label }}
    </p>

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
      Signal history builds as advertisements arrive. Trend needs a few readings before it can
      show a direction.
    </p>

    <div class="observation-focus__actions">
      <UiButton
        v-if="isUntriaged"
        size="sm"
        variant="primary"
        :disabled="props.pending"
        @click="emit('decide', props.identity.key, 'known')"
      >
        Mark known
      </UiButton>
      <UiButton
        v-if="props.identity.decision !== ObservationDecisions.Ignored"
        size="sm"
        variant="ghost"
        :disabled="props.pending"
        @click="emit('decide', props.identity.key, 'ignored')"
      >
        Ignore
      </UiButton>
    </div>
  </section>
</template>

<style scoped>
  .observation-focus {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 1rem 0;
    text-align: center;
  }

  .observation-focus__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    gap: 0.5rem;
  }

  .observation-focus__badge {
    font-size: 0.75rem;
    color: var(--warning-text, #e0a33e);
    border: 1px dashed currentColor;
    border-radius: 999px;
    padding: 0 0.5rem;
  }

  .observation-focus__title {
    margin: 0;
    font-size: 1.25rem;
    word-break: break-word;
  }

  .observation-focus__key {
    margin: 0;
    color: var(--text-muted, #8b9bab);
    font-size: 0.75rem;
    word-break: break-all;
  }

  /* A square canvas keeps the rings circular at every width. */
  .observation-focus__radar {
    width: min(16rem, 60vw);
    aspect-ratio: 1;
  }

  .observation-focus__radar svg {
    width: 100%;
    height: 100%;
  }

  .observation-focus__ring {
    fill: none;
    stroke: var(--border-subtle, #2a3441);
    stroke-width: 0.5;
  }

  .observation-focus__centre {
    fill: var(--text-muted, #8b9bab);
  }

  .observation-focus__target {
    fill: var(--accent-primary, #4a9eff);
  }

  .observation-focus__distance {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    margin: 0;
  }

  .observation-focus__value {
    font-size: 2rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }

  .observation-focus__band {
    color: var(--text-muted, #8b9bab);
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .observation-focus__trend {
    margin: 0;
    font-size: 0.9375rem;
  }

  .observation-focus__arrow {
    font-weight: 700;
  }

  .observation-focus__spark {
    width: 100%;
    max-width: 20rem;
    margin: 0.25rem 0 0;
  }

  .observation-focus__spark svg {
    width: 100%;
    height: 3rem;
  }

  .observation-focus__spark-line {
    fill: none;
    stroke: var(--accent-primary, #4a9eff);
    stroke-width: 1.5;
    vector-effect: non-scaling-stroke;
  }

  .observation-focus__spark figcaption,
  .observation-focus__note {
    color: var(--text-muted, #8b9bab);
    font-size: 0.75rem;
    margin: 0.25rem 0 0;
  }

  .observation-focus__actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }
</style>
