<script setup lang="ts">
  /*
   * Range-only polar display for wireless observations (spec 064, ADR-003).
   *
   * One source measures RSSI, which yields a distance and no bearing. Blips are therefore
   * placed on a stable spoke derived from the identity key and their radius is the estimated
   * range, so movement reads as radial drift. The legend states range-only explicitly: this
   * view must never imply a position it cannot compute.
   */
  import { computed, useId } from 'vue';
  import type { ObservationIdentityView } from '@/types/observations.type';
  import {
    DEFAULT_PATH_LOSS_EXPONENT,
    DEFAULT_TX_POWER_AT_1M,
    RANGE_RINGS_M,
    bearingForIdentity,
    distanceRSSI,
    estimateDistanceMetres,
    identityTitle,
    isEphemeralIdentity,
  } from '@/domain/observations/presentation';

  const props = withDefaults(
    defineProps<{
      identities: readonly ObservationIdentityView[];
      /** Currently isolated identity, so its blip can be marked. */
      selectedKey?: string | null;
      /** Transmit power at one metre in dBm; sources may override the default. */
      txPowerAt1m?: number;
      pathLossExponent?: number;
    }>(),
    {
      txPowerAt1m: DEFAULT_TX_POWER_AT_1M,
      pathLossExponent: DEFAULT_PATH_LOSS_EXPONENT,
    },
  );

  const emit = defineEmits<{
    (event: 'select', key: string): void;
  }>();

  const titleId = useId();

  /** Outer radius in view units; the largest ring sits at the edge. */
  const outerRadius = 100;
  const maxRing = RANGE_RINGS_M[RANGE_RINGS_M.length - 1];

  /**
   * Placeable identities only. An identity with no RSSI or an out-of-range estimate cannot be
   * positioned, and is listed as unplaced rather than drawn at the centre, which would read
   * as "very close".
   */
  const placed = computed(() =>
    props.identities.flatMap((identity) => {
      // Smoothed signal keeps the blip from jumping between rings on single-packet jitter.
      const metres = estimateDistanceMetres(distanceRSSI(identity), props.txPowerAt1m, props.pathLossExponent);
      if (metres === null) return [];
      const bearing = (bearingForIdentity(identity.key) * Math.PI) / 180;
      // Log scale so each ring is equally readable; clamp beyond the outer ring to the edge.
      const ratio = Math.min(Math.max(Math.log10(Math.max(metres, 0.1)) / Math.log10(maxRing), 0.04), 1);
      return [
        {
          identity,
          metres,
          x: 50 + Math.sin(bearing) * ratio * outerRadius * 0.9,
          y: 50 - Math.cos(bearing) * ratio * outerRadius * 0.9,
          ephemeral: isEphemeralIdentity(identity),
        },
      ];
    }),
  );

  const unplaced = computed(() => props.identities.length - placed.value.length);

  /** Ring circles with their labels, drawn outermost first so labels stay legible. */
  const rings = computed(() =>
    RANGE_RINGS_M.map((metres) => ({
      metres,
      radius: (Math.log10(metres) / Math.log10(maxRing)) * outerRadius * 0.9,
    })).reverse(),
  );

  function radiusLabel(metres: number): string {
    return metres < 1000 ? `${metres} m` : `${metres / 1000} km`;
  }
</script>

<template>
  <figure class="observation-radar">
    <svg
      :aria-labelledby="titleId"
      class="observation-radar__canvas"
      role="img"
      viewBox="-8 -8 116 116"
    >
      <title :id="titleId">Observed devices by distance, range only, no bearing</title>

      <circle
        v-for="ring in rings"
        :key="ring.metres"
        class="observation-radar__ring"
        cx="50"
        cy="50"
        fill="none"
        :r="ring.radius"
      />
      <text
        v-for="ring in rings"
        :key="`label-${ring.metres}`"
        class="observation-radar__ring-label"
        :x="50 + 1"
        :y="50 - ring.radius - 1"
      >
        {{ radiusLabel(ring.metres) }}
      </text>

      <line class="observation-radar__axis" x1="50" x2="50" y1="4" y2="96" />
      <line class="observation-radar__axis" x1="4" x2="96" y1="50" y2="50" />

      <!--
        Each blip is a real control, not a decorative dot: clicking isolates that identity.
        SVG groups cannot receive focus or announce a role, so each blip carries an explicit
        tabindex, role and label. A wider transparent hit area sits under the visible dot,
        because a 2.4-unit radius is far below a comfortable touch target.
      -->
      <g
        v-for="blip in placed"
        :key="blip.identity.key"
        class="observation-radar__blip-group"
        :class="{ 'is-selected': props.selectedKey === blip.identity.key }"
        role="button"
        tabindex="0"
        :aria-pressed="props.selectedKey === blip.identity.key"
        :aria-label="`Isolate ${identityTitle(blip.identity)}, ${Math.round(blip.metres)} metres`"
        :data-identity-key="blip.identity.key"
        @click="emit('select', blip.identity.key)"
        @keydown.enter.prevent="emit('select', blip.identity.key)"
        @keydown.space.prevent="emit('select', blip.identity.key)"
      >
        <circle
          class="observation-radar__hit"
          :cx="blip.x"
          :cy="blip.y"
          r="6"
        />
        <circle
          class="observation-radar__blip"
          :class="{
            'observation-radar__blip--ephemeral': blip.ephemeral,
            'observation-radar__blip--known': blip.identity.decision === 'known',
          }"
          :cx="blip.x"
          :cy="blip.y"
          r="2.4"
        />
        <title>{{ blip.identity.label || blip.identity.key }} · {{ Math.round(blip.metres) }} m</title>
      </g>
    </svg>

    <figcaption class="observation-radar__caption">
      Range only · no bearing. Distances are estimated from signal strength and are accurate to
      a few metres indoors.
      <span v-if="unplaced > 0" class="observation-radar__unplaced">
        {{ unplaced }} without a signal reading not shown.
      </span>
    </figcaption>
  </figure>
</template>

<style scoped>
  .observation-radar {
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .observation-radar__canvas {
    width: 100%;
    max-width: 28rem;
    height: auto;
    aspect-ratio: 1;
    background: var(--surface-sunken, #0f1419);
    border-radius: 50%;
  }

  .observation-radar__ring {
    stroke: var(--border-subtle, #2a3441);
    stroke-width: 0.4;
  }

  .observation-radar__ring-label {
    fill: var(--text-muted, #8b9bab);
    font-size: 3px;
  }

  .observation-radar__axis {
    stroke: var(--border-subtle, #2a3441);
    stroke-width: 0.3;
    stroke-dasharray: 1 1.5;
  }

  /* Invisible but clickable: the visible dot is far smaller than a usable touch target. */
  .observation-radar__hit {
    fill: transparent;
    cursor: pointer;
  }

  .observation-radar__blip-group {
    cursor: pointer;
    outline: none;
  }

  .observation-radar__blip-group:focus-visible .observation-radar__blip {
    stroke: var(--accent-primary, #4a9eff);
    stroke-width: 1.4;
  }

  .observation-radar__blip-group:hover .observation-radar__blip {
    stroke: var(--text-primary, #e6edf3);
    stroke-width: 1;
  }

  /* The isolated identity is highlighted so the radar and the Focus panel agree. */
  .observation-radar__blip-group.is-selected .observation-radar__blip {
    fill: var(--accent-primary, #4a9eff);
    stroke: var(--text-primary, #e6edf3);
    stroke-width: 1.2;
  }

  .observation-radar__blip {
    fill: var(--accent-primary, #4a9eff);
    stroke: var(--surface-sunken, #0f1419);
    stroke-width: 0.3;
  }

  /* Known devices are de-emphasised so new arrivals stand out. */
  .observation-radar__blip--known {
    fill: var(--accent-muted, #4b6b8a);
  }

  /* A rotating address is explicitly hollow: it is not a stable device. */
  .observation-radar__blip--ephemeral {
    fill: none;
    stroke: var(--text-muted, #8b9bab);
    stroke-width: 0.7;
    stroke-dasharray: 1 0.8;
  }

  .observation-radar__caption {
    color: var(--text-muted, #8b9bab);
    font-size: 0.8125rem;
  }

  .observation-radar__unplaced {
    display: block;
  }

  @media (prefers-reduced-motion: reduce) {
    .observation-radar__blip {
      transition: none;
    }
  }
</style>
