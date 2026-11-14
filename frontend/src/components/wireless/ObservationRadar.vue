<script setup lang="ts">
  /*
   * Range-only polar display for wireless observations (spec 064, ADR-003).
   *
   * One source measures RSSI, which yields a distance and no bearing. Blips are therefore
   * placed on a stable spoke derived from the identity key and their radius is the estimated
   * range, so movement reads as radial drift. The legend states range-only explicitly: this
   * view must never imply a position it cannot compute.
   */
  import { computed, useId, watch } from 'vue';
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
    labelableRings,
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
      /** Whether ring labels are drawn. The circles always draw; labels are an aid. */
      showLabels?: boolean;
    }>(),
    {
      txPowerAt1m: DEFAULT_TX_POWER_AT_1M,
      pathLossExponent: DEFAULT_PATH_LOSS_EXPONENT,
      span: null,
      showLabels: true,
    },
  );

  const emit = defineEmits<{
    (event: 'select', key: string): void;
    /**
     * The span in metres that this canvas actually drew.
     *
     * Emitted rather than recomputed by the caller: Auto excludes rotating addresses here, so a
     * second calculation elsewhere would disagree with what is on screen — which is exactly how a
     * readout came to say 5 m while the rings were drawn at 10 m.
     */
    (event: 'span-changed', metres: number): void;
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

  /** Distance for every identity that has a usable reading. */
  const measured = computed(() =>
    props.identities.flatMap((identity) => {
      // Smoothed signal keeps the blip from jumping between rings on single-packet jitter.
      const metres = estimateDistanceMetres(distanceRSSI(identity), props.txPowerAt1m, props.pathLossExponent);
      return metres === null ? [] : [{ identity, metres }];
    }),
  );

  /**
   * Active span, from the devices the operator can actually identify.
   *
   * Rotating addresses are excluded from the calculation on purpose. There are typically far more
   * of them than anything else, and they are the *least* useful to frame: a phone that re-keys
   * every few minutes cannot be tracked, so letting its distance set the scale repeats the outlier
   * mistake with a bigger crowd. If nothing identifiable is in range the whole set is used, so an
   * all-rotating house still gets a sensible scale.
   */
  const activeSpan = computed<RadarSpan>(() => {
    if (props.span) return props.span;
    const identifiable = measured.value.filter((entry) => !isEphemeralIdentity(entry.identity));
    const source = identifiable.length > 0 ? identifiable : measured.value;
    return autoRadarSpan(source.map((entry) => entry.metres));
  });

  // Report the settled span whenever it changes, so the toolbar readout and the zoom step agree
  // with the canvas by construction.
  watch(
    activeSpan,
    (span) => emit('span-changed', span),
    { immediate: true },
  );

  /** Rotating-address devices, which are drawn as a count rather than as individual blips. */
  const rotating = computed(() => measured.value.filter((entry) => isEphemeralIdentity(entry.identity)));

  /**
   * How many rotating devices sit at each ring band.
   *
   * Drawn as a count next to the band rather than as one hollow circle each. Thirty identical
   * hollow circles carry no information — you cannot tell them apart, cannot see them change, and
   * they bury the handful of devices you can actually identify. The count keeps the information
   * ("there are nine unidentifiable devices around 3 m") and drops the noise.
   */
  const rotatingBands = computed(() => {
    const bandCounts = new Map<string, { label: string; count: number }>();
    for (const entry of rotating.value) {
      const ratio = radialRatio(entry.metres, activeSpan.value);
      const band = ringBandFor(ratio, activeSpan.value);
      const existing = bandCounts.get(band.id);
      if (existing) existing.count++;
      else bandCounts.set(band.id, { label: band.label, count: 1 });
    }
    return [...bandCounts.values()];
  });

  /**
   * Placeable identities. An identity with no reading cannot be positioned, and is counted as
   * unplaced rather than drawn at the centre, which would read as "very close".
   */
  const placed = computed(() =>
    measured.value.filter(({ identity }) => !isEphemeralIdentity(identity)).map(({ identity, metres }) => {
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
        // Where this device was a few readings ago, when its trend says it is moving.
        //
        // A 3 dB change — the smallest the trend will report — moves a blip about 5% of the radius
        // at a wide span, which is invisible on a static screenshot. Drawing a faint tail from the
        // earlier position makes that motion legible as a direction, which is the thing the trend
        // is actually claiming. Only drawn for approaching/receding, so a steady device adds no
        // noise.
        tail: tailFor(identity, metres, activeSpan.value, bearing),
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

  /**
   * Devices with no usable reading at all.
   *
   * Rotating devices are counted separately, so they must not be double-reported here: the caption
   * said "28 without a signal reading not shown" about devices it had just described as counted.
   */
  const unplaced = computed(() =>
    props.identities.length - placed.value.length - rotating.value.length,
  );

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
        ratio: ring.ratio,
        metres: ring.metres,
        outer: ring.outer,
        radius: ring.ratio * outerRadius * RADAR_RADIUS_FRACTION,
      }))
      .reverse(),
  );

  /**
   * Rings that get a label.
   *
   * Pruned so labels never overlap when the scale is crowded, and empty when the operator turns
   * labels off. The circles are drawn from `rings`, not from this, so turning labels off never
   * changes the geometry.
   */
  const labelledRings = computed(() => {
    if (!props.showLabels) return [];
    // labelableRings sorts its own input, so it does not care that `rings` is outer-first here.
    const allowed = new Set(labelableRings(rings.value).map((ring) => ring.ratio));
    return rings.value.filter((ring) => allowed.has(ring.ratio));
  });

  /**
   * The previous position of a moving device, or null when it is not moving.
   *
   * Derived from the trend delta rather than from stored history: the trend says how many dB the
   * signal moved, and the projection turns that back into a canvas position. This keeps the radar
   * a pure function of the snapshot, so a reload cannot show a different tail.
   */
  function tailFor(
    identity: ObservationIdentityView,
    currentMetres: number,
    span: RadarSpan,
    bearing: number,
  ): { x: number; y: number } | null {
    const delta = identity.trendDeltaDb ?? 0;
    if (Math.abs(delta) < 1) return null;
    if (identity.trend !== 'approaching' && identity.trend !== 'receding') return null;

    // Undo the trend: a stronger signal means it was further away before.
    const previousRSSI = (distanceRSSI(identity) ?? 0) - delta;
    const previousMetres = estimateDistanceMetres(previousRSSI, props.txPowerAt1m, props.pathLossExponent);
    if (previousMetres === null) return null;

    const ratio = radialRatio(previousMetres, span);
    return {
      x: CENTRE + Math.sin(bearing) * ratio * outerRadius * RADAR_RADIUS_FRACTION,
      y: CENTRE - Math.cos(bearing) * ratio * outerRadius * RADAR_RADIUS_FRACTION,
    };
  }

  /**
   * Which ring band a radial ratio falls in, for grouping rotating devices.
   *
   * Bands come from the same `rings` the canvas draws, so a count always describes a region the
   * operator can see. Returns the innermost band below the first ring for anything closer.
   */
  function ringBandFor(ratio: number, span: RadarSpan): { id: string; label: string } {
    // Boundary ratios are fixed fractions of the radius, so they are produced from the span rather
    // than read back from the rendered rings. Reading the rendered set here ran before the canvas
    // had recomputed, which is how a 5 m span produced a band labelled "under 10 m".
    const boundaries = ringsForSpan(span)
      .map((ring) => ({ metres: ring.metres, ratio: ring.ratio }))
      .sort((a, b) => a.ratio - b.ratio);
    for (const boundary of boundaries) {
      if (ratio <= boundary.ratio) {
        return { id: `under-${boundary.metres}`, label: `under ${formatMetres(boundary.metres)}` };
      }
    }
    const outermost = boundaries[boundaries.length - 1];
    return { id: `over-${outermost?.metres ?? 0}`, label: `over ${formatMetres(outermost?.metres ?? 0)}` };
  }

  /** Distance for a band label, matching the ring label formatting. */
  function formatMetres(metres: number): string {
    return metres < 1 ? `${metres.toFixed(1)} m` : `${Math.round(metres)} m`;
  }

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

      <!-- Shared arrowhead for motion tails: one definition, so no per-blip element is needed. -->
      <defs>
        <marker
          id="observ-radar-arrow"
          viewBox="0 0 6 6"
          refX="5"
          refY="3"
          markerWidth="3"
          markerHeight="3"
          orient="auto"
        >
          <path d="M0,0 L6,3 L0,6 z" fill="currentColor" />
        </marker>
      </defs>

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
        v-for="ring in labelledRings"
        :key="`label-${ring.radius}`"
        class="observation-radar__ring-label"
        :x="CENTRE - ring.radius * 0.707 + 1.5"
        :y="CENTRE + ring.radius * 0.707 - 1.5"
        text-anchor="start"
      >
        {{ radiusLabel(ring.metres) }}
      </text>

      <!--
        Unidentifiable devices, as counts rather than blips. Each line names a band and how many
        rotating-address devices sit inside it, which is the information those hollow circles were
        carrying — without thirty identical marks burying everything that can be identified.
      -->
      <text
        v-for="(band, index) in rotatingBands"
        :key="band.label"
        class="observation-radar__rotating-count"
        :x="CENTRE + 3"
        :y="CENTRE + 12 + index * 4.2"
      >
        {{ band.count }} rotating {{ band.label }}
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
        :aria-label="`Isolate ${identityTitle(blip.identity)}, ${Math.round(blip.metres)} m`"
        :data-identity-key="blip.identity.key"
        @click="emit('select', blip.identity.key)"
        @keydown.enter.prevent="emit('select', blip.identity.key)"
        @keydown.space.prevent="emit('select', blip.identity.key)"
      >
        <!--
          Motion tail. Drawn beneath the blip so the current position stays the strongest mark,
          and as a line rather than a second dot so the direction reads at a glance.
        -->
        <line
          v-if="blip.tail"
          class="observation-radar__tail"
          :class="`observation-radar__tail--${blip.identity.trend}`"
          :x1="blip.tail.x"
          :y1="blip.tail.y"
          :x2="blip.x"
          :y2="blip.y"
        />
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
      <span v-if="rotating.length > 0" class="observation-radar__unplaced">
        {{ rotating.length }} with rotating addresses are counted by distance, not drawn
        individually — they cannot be told apart.
      </span>
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
    /*
     * Bounded by *width*, because the element is square: `width:100%` plus `aspect-ratio:1` makes
     * height follow width, so a max-height alone is ignored and the canvas overruns the fold. A
     * width cap is the only constraint that actually binds for a square element in a wide column.
     */
    width: min(100%, 22rem);
    height: auto;
    aspect-ratio: 1;
    /*
     * Sized by height budget rather than by the column. A square element with width:100% grows to
     * whatever the column allows, so in the radar's wide column it ran past the bottom of the
     * screen — the controls above it were reachable but the lower half of the radar, and the
     * caption under it, were not. Capping the height and centring fixes that without shrinking the
     * usable drawing area, because the viewBox scales uniformly.
     */
    max-height: 22rem;
    margin-inline: auto;
    background: var(--surface-sunken, #0f1419);
    border-radius: 50%;
  }

  .observation-radar__ring {
    stroke: var(--border-subtle, #2a3441);
    stroke-width: 0.4;
  }

  /*
   * Ring labels are sized in view units, which the viewBox then scales to the canvas. At 3 units
   * they rendered at about 9px on a 352px canvas — technically present, practically unreadable,
   * which is why the labels looked missing. 4.5 units lands near 14px at that size.
   *
   * The halo is a stroke of the canvas colour painted under the fill: labels sit on top of the
   * rings and the axis line, and without it a label crossing a circle became illegible.
   */
  /*
   * Motion tail. Faint enough not to compete with the blips, long enough to read as a direction.
   * The arrowhead is the marker, which needs no extra element and no per-blip maths.
   */
  .observation-radar__tail {
    stroke-width: 0.7;
    opacity: 0.55;
    marker-end: url(#observ-radar-arrow);
  }

  .observation-radar__tail--approaching {
    stroke: var(--accent-strong, #4a9eff);
  }

  .observation-radar__tail--receding {
    stroke: var(--text-subtle, #868e96);
    stroke-dasharray: 1.5 1;
  }

  /*
   * Rotating-device counts. Dimmer than a blip because they are context, not targets: they say how
   * much unidentifiable traffic is around without competing with the devices that have names.
   */
  .observation-radar__rotating-count {
    fill: var(--text-subtle, #7c8b9c);
    font-size: 3.2px;
    pointer-events: none;
    user-select: none;
  }

  .observation-radar__ring-label {
    fill: var(--text-body, #c9d6e2);
    font-size: 3.8px;
    font-weight: 600;
    paint-order: stroke;
    stroke: var(--surface-sunken, #0f1419);
    stroke-width: 1.1px;
    stroke-linejoin: round;
    /* Labels are decoration over a control-dense canvas; they must never eat a blip's click. */
    pointer-events: none;
    user-select: none;
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
