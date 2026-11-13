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
    RADAR_RADIUS_FRACTION,
    autoRadarSpan,
    bearingForIdentity,
    distanceRSSI,
    classPresentation,
    estimateDistanceMetres,
    identityTitle,
    isEphemeralIdentity,
    radialRatio,
    ringsForSpan,
    type RadarSpan,
  } from '@/domain/observations/presentation';

  const props = withDefaults(
    defineProps<{
      identities: readonly ObservationIdentityView[];
      /** Currently isolated identity, so its blip can be marked. */
      selectedKey?: string | null;
      /** Transmit power at one metre in dBm; sources may override the default. */
      txPowerAt1m?: number;
      pathLossExponent?: number;
      /**
       * Fixed span in metres, or null to scale to the data. A fixed span keeps the scale still
       * while readings change, which is what makes small movements visible.
       */
      span?: RadarSpan | null;
    }>(),
    {
      txPowerAt1m: DEFAULT_TX_POWER_AT_1M,
      pathLossExponent: DEFAULT_PATH_LOSS_EXPONENT,
      span: null,
    },
  );

  const emit = defineEmits<{
    (event: 'select', key: string): void;
  }>();

  const titleId = useId();

  /*
   * Canvas geometry, derived from one number.
   *
   * The viewBox is a 100-unit square with the centre at (50, 50), so every coordinate in the
   * template used to repeat those figures as literals. They are named here instead: a blip drawn
   * with one centre and a ring drawn with another would silently disagree, and that is exactly the
   * kind of defect that looks like "the scale is broken" without being obvious to trace.
   */
  /** Centre of the canvas in view units. */
  const CENTRE = 50;
  /** Radius of the drawable area in view units. */
  const outerRadius = CENTRE;
  /** Padding outside the outermost ring, so its label is not clipped by the viewBox edge. */
  const VIEWBOX_PADDING = 8;
  /** Axis line extent, inset from the edge so the tick marks do not touch the border. */
  const AXIS_INSET = 4;
  /** Offset of a ring label from its circle, so the text does not sit on the stroke. */
  const LABEL_OFFSET = 1;

  /** Distance for every identity that has a usable reading. */
  const measured = computed(() =>
    props.identities.flatMap((identity) => {
      // Smoothed signal keeps the blip from jumping between rings on single-packet jitter.
      const metres = estimateDistanceMetres(distanceRSSI(identity), props.txPowerAt1m, props.pathLossExponent);
      return metres === null ? [] : [{ identity, metres }];
    }),
  );

  /**
   * Active span. Auto follows the furthest reading, so a nearby-only household fills the canvas
   * instead of clustering at the centre of a 30 m scale.
   */
  const activeSpan = computed<RadarSpan>(() =>
    props.span ?? autoRadarSpan(measured.value.map((entry) => entry.metres)),
  );

  /**
   * Placeable identities. An identity with no reading cannot be positioned, and is counted as
   * unplaced rather than drawn at the centre, which would read as "very close".
   */
  const placed = computed(() =>
    measured.value.map(({ identity, metres }) => {
      const bearing = (bearingForIdentity(identity.key) * Math.PI) / 180;
      const ratio = radialRatio(metres, activeSpan.value);
      return {
        identity,
        metres,
        // Beyond the span is drawn on the edge rather than hidden: it is still in range.
        beyondSpan: metres > activeSpan.value,
        x: CENTRE + Math.sin(bearing) * ratio * outerRadius * RADAR_RADIUS_FRACTION,
        y: CENTRE - Math.cos(bearing) * ratio * outerRadius * RADAR_RADIUS_FRACTION,
        ephemeral: isEphemeralIdentity(identity),
        // The class tone is resolved here rather than in the template so an absent classification
        // still produces a valid class name, which keeps the styling total.
        tone: classPresentation(identity.classification).tone,
        label: classPresentation(identity.classification).label,
      };
    }),
  );

  /**
   * Legend entries for the classes currently on the radar, in a fixed order so it does not
   * reshuffle as devices come and go.
   */
  const legend = computed(() => {
    const seen = new Map<string, string>();
    for (const blip of placed.value) {
      if (!seen.has(blip.tone)) seen.set(blip.tone, blip.label);
    }
    const order = ['installed', 'yours', 'outside', 'passing', 'learning'];
    return order.filter((tone) => seen.has(tone)).map((tone) => ({ tone, label: seen.get(tone)! }));
  });

  /** Number of identities drawn at the edge because they exceed the span. */
  const beyondSpan = computed(() => placed.value.filter((blip) => blip.beyondSpan).length);

  const unplaced = computed(() => props.identities.length - placed.value.length);

  /**
   * Ring circles with their labels, drawn outermost first so labels stay legible.
   *
   * Radius comes from the ring's own ratio, not from re-projecting its rounded label: rounding
   * 0.67 m to "0.7 m" would otherwise move the circle and reintroduce the uneven gaps the ratio
   * exists to prevent.
   */
  const rings = computed(() =>
    ringsForSpan(activeSpan.value)
      .map((ring) => ({
        metres: ring.metres,
        outer: ring.outer,
        radius: ring.ratio * outerRadius * RADAR_RADIUS_FRACTION,
      }))
      .reverse(),
  );

  /** Ring label. Sub-metre rings keep their decimal, or they would all read as `0 m`. */
  function radiusLabel(metres: number): string {
    return `${metres} m`;
  }
</script>

<template>
  <figure class="observation-radar">
    <svg
      :aria-labelledby="titleId"
      class="observation-radar__canvas"
      role="img"
      :viewBox="`${-VIEWBOX_PADDING} ${-VIEWBOX_PADDING} ${2 * (CENTRE + VIEWBOX_PADDING)} ${2 * (CENTRE + VIEWBOX_PADDING)}`"
    >
      <title :id="titleId">Observed devices by distance, range only, no bearing</title>

      <circle
        v-for="ring in rings"
        :key="`ring-${ring.radius}`"
        class="observation-radar__ring"
        :cx="CENTRE"
        :cy="CENTRE"
        fill="none"
        :r="ring.radius"
      />
      <text
        v-for="ring in rings"
        :key="`label-${ring.radius}`"
        class="observation-radar__ring-label"
        :x="CENTRE + LABEL_OFFSET"
        :y="CENTRE - ring.radius - LABEL_OFFSET"
      >
        {{ radiusLabel(ring.metres) }}
      </text>

      <line
        class="observation-radar__axis"
        :x1="CENTRE"
        :x2="CENTRE"
        :y1="AXIS_INSET"
        :y2="2 * CENTRE - AXIS_INSET"
      />
      <line
        class="observation-radar__axis"
        :x1="AXIS_INSET"
        :x2="2 * CENTRE - AXIS_INSET"
        :y1="CENTRE"
        :y2="CENTRE"
      />

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
        <!--
          Shape encodes class as well as colour, so the radar stays readable without colour
          vision. A rotating address keeps its dashed ring from spec 064; class adds a tone.
        -->
        <circle
          class="observation-radar__blip"
          :class="[
            {
              'observation-radar__blip--ephemeral': blip.ephemeral,
              'observation-radar__blip--known': blip.identity.decision === 'known',
            },
            `observation-radar__blip--${blip.tone}`,
          ]"
          :cx="blip.x"
          :cy="blip.y"
          r="2.4"
        />
        <title>
          {{ blip.identity.label || blip.identity.key }} · {{ Math.round(blip.metres) }} m ·
          {{ blip.label }}
        </title>
      </g>
    </svg>

    <figcaption class="observation-radar__caption">
      <!--
        The legend names each class in words, so the radar is readable without colour vision and
        the operator learns what the tones mean. It lists only the classes actually drawn, which
        stops it becoming a permanent key to things that are not on screen.
      -->
      <span v-if="legend.length > 0" class="observation-radar__legend">
        <span v-for="entry in legend" :key="entry.tone" class="observation-radar__legend-item">
          <span class="observation-radar__legend-dot" :class="`observation-radar__blip--${entry.tone}`" aria-hidden="true" />
          {{ entry.label }}
        </span>
      </span>
      Range only · no bearing. Distances are estimated from signal strength and are accurate to
      a few metres indoors.
      <span v-if="unplaced > 0" class="observation-radar__unplaced">
        {{ unplaced }} without a signal reading not shown.
      </span>
      <span v-if="beyondSpan > 0" class="observation-radar__unplaced">
        {{ beyondSpan }} beyond {{ activeSpan }} m shown at the edge.
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
    min-width: 0;
    /* Centre the square canvas when the column is wider than the height cap. */
    align-items: center;
  }

  /*
   * The canvas fills its container. The old 28rem cap kept the global radar small even when the
   * layout had far more room, which is why it was hard to read.
   */
  .observation-radar__canvas {
    display: block;
    width: 100%;
    height: auto;
    aspect-ratio: 1;
    /*
     * Height grows with the viewport rather than the content column: the canvas is square, so a
     * modest height cap would silently shrink it. 82vh keeps the caption and toolbar on screen
     * at typical desktop heights while still letting the radar dominate.
     */
    max-height: 82vh;
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

  .observation-radar__blip-group:focus-visible .observation-radar__blip--installed {
  fill: var(--text-muted, #495057);
  opacity: 0.55;
}

.observation-radar__blip--yours {
  fill: var(--accent-strong, #1864ab);
}

.observation-radar__blip--outside {
  fill: var(--text-subtle, #868e96);
  opacity: 0.6;
}

.observation-radar__blip--passing {
  fill: var(--text-subtle, #868e96);
  opacity: 0.4;
}

.observation-radar__blip--learning {
  fill: var(--text-muted, #495057);
  opacity: 0.8;
}

.observation-radar__blip {
    stroke: var(--accent-primary, #4a9eff);
    stroke-width: 1.4;
  }

  .observation-radar__blip-group:hover .observation-radar__blip--installed {
  fill: var(--text-muted, #495057);
  opacity: 0.55;
}

.observation-radar__blip--yours {
  fill: var(--accent-strong, #1864ab);
}

.observation-radar__blip--outside {
  fill: var(--text-subtle, #868e96);
  opacity: 0.6;
}

.observation-radar__blip--passing {
  fill: var(--text-subtle, #868e96);
  opacity: 0.4;
}

.observation-radar__blip--learning {
  fill: var(--text-muted, #495057);
  opacity: 0.8;
}

.observation-radar__blip {
    stroke: var(--text-primary, #e6edf3);
    stroke-width: 1;
  }

  /* The isolated identity is highlighted so the radar and the Focus panel agree. */
  .observation-radar__blip-group.is-selected .observation-radar__blip--installed {
  fill: var(--text-muted, #495057);
  opacity: 0.55;
}

.observation-radar__blip--yours {
  fill: var(--accent-strong, #1864ab);
}

.observation-radar__blip--outside {
  fill: var(--text-subtle, #868e96);
  opacity: 0.6;
}

.observation-radar__blip--passing {
  fill: var(--text-subtle, #868e96);
  opacity: 0.4;
}

.observation-radar__blip--learning {
  fill: var(--text-muted, #495057);
  opacity: 0.8;
}

.observation-radar__blip {
    fill: var(--accent-primary, #4a9eff);
    stroke: var(--text-primary, #e6edf3);
    stroke-width: 1.2;
  }

  .observation-radar__blip--installed {
  fill: var(--text-muted, #495057);
  opacity: 0.55;
}

.observation-radar__blip--yours {
  fill: var(--accent-strong, #1864ab);
}

.observation-radar__blip--outside {
  fill: var(--text-subtle, #868e96);
  opacity: 0.6;
}

.observation-radar__blip--passing {
  fill: var(--text-subtle, #868e96);
  opacity: 0.4;
}

.observation-radar__blip--learning {
  fill: var(--text-muted, #495057);
  opacity: 0.8;
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

  .observation-radar__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 0.75rem;
  margin-bottom: 0.3rem;
}

.observation-radar__legend-item {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.observation-radar__legend-dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
}

.observation-radar__unplaced {
    display: block;
  }

  @media (prefers-reduced-motion: reduce) {
    .observation-radar__blip--installed {
  fill: var(--text-muted, #495057);
  opacity: 0.55;
}

.observation-radar__blip--yours {
  fill: var(--accent-strong, #1864ab);
}

.observation-radar__blip--outside {
  fill: var(--text-subtle, #868e96);
  opacity: 0.6;
}

.observation-radar__blip--passing {
  fill: var(--text-subtle, #868e96);
  opacity: 0.4;
}

.observation-radar__blip--learning {
  fill: var(--text-muted, #495057);
  opacity: 0.8;
}

.observation-radar__blip {
      transition: none;
    }
  }
</style>
