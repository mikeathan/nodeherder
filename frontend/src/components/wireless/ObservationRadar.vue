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
      /** Whether ring distance labels are drawn. The circles always draw; labels are an aid. */
      showLabels?: boolean;
      /** Whether device names are drawn beside their blips. */
      showNames?: boolean;
    }>(),
    {
      txPowerAt1m: DEFAULT_TX_POWER_AT_1M,
      pathLossExponent: DEFAULT_PATH_LOSS_EXPONENT,
      span: null,
      showLabels: true,
      showNames: true,
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
  /** Device-name font size in view units, and the longest name that can fit beside a blip. */
  const LABEL_FONT = 3.2;
  const LABEL_MAX_CHARS = 18;

  /** Distance for every identity that has a usable reading. */
  const measured = computed(() =>
    props.identities.flatMap((identity) => {
      // Smoothed signal keeps the blip from jumping between rings on single-packet jitter.
      const metres = estimateDistanceMetres(distanceRSSI(identity), props.txPowerAt1m, props.pathLossExponent);
      return metres === null ? [] : [{ identity, metres }];
    }),
  );

  /**
   * Active span, from every device in range.
   *
   * Excluding rotating addresses from this calculation was a mistake, and a visible one: with a
   * single identifiable device at 0.3 m the scale collapsed to 1 m, so twelve devices out to 25 m
   * were all clamped to the canvas edge, stacked on top of each other, and reported as "under 1 m"
   * in the counts while their cards said 17 m. The scale has to describe what is drawn.
   *
   * Rotating devices are collapsed into counts rather than blips, but they are still the majority
   * of what is out there, and a radar that does not reach them is not describing the room.
   * `autoRadarSpan` already frames the median rather than the maximum, so a long tail does not drag
   * the scale wide.
   */
  const activeSpan = computed<RadarSpan>(() => {
    if (props.span) return props.span;
    return autoRadarSpan(measured.value.map((entry) => entry.metres));
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
    // Boundaries are computed once per span, not once per device. Rebuilding the ring set inside
    // the loop made this O(devices x rings) on every render for a value that is identical for all
    // of them — 28 devices meant 28 throwaway arrays.
    const boundaries = ringBandsForSpan(activeSpan.value);
    const bandCounts = new Map<string, { label: string; count: number }>();

    for (const entry of rotating.value) {
      // Anything beyond the span is bucketed by its *actual* distance, not by where it was drawn.
      // Clamping first put a device at 25 m into the outermost ring and labelled it "under 1 m",
      // contradicting the card beside it which correctly said 25 m.
      const band =
        entry.metres > activeSpan.value
          ? beyondSpanBand(entry.metres)
          : bandForRatio(boundaries, radialRatio(entry.metres, activeSpan.value));

      const existing = bandCounts.get(band.id);
      if (existing) existing.count++;
      else bandCounts.set(band.id, { label: band.label, count: 1 });
    }
    return [...bandCounts.values()];
  });

  /**
   * The bucket for a device beyond the span.
   *
   * Grouped on a 1-2-5 ladder so a handful of far devices become one or two readable lines rather
   * than one line each, and labelled with the distance it actually is.
   */
  function beyondSpanBand(metres: number): { id: string; label: string } {
    const steps = [2, 5, 10, 20, 50, 100, 200];
    const step = steps.find((value) => metres <= value) ?? Math.ceil(metres / 100) * 100;
    return { id: `beyond-${step}`, label: `beyond span, up to ${step} m` };
  }

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
        tail: tailFor(identity, activeSpan.value, bearing),
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
  /**
   * Devices drawn at the edge because they exceed the span.
   *
   * Counts the rotating devices too. They are no longer drawn as blips, but they are still
   * positioned beyond the span, so excluding them from this count meant the caption said nothing
   * while a dozen devices sat clamped on the same spot — which read as a single dot.
   */
  const beyondSpan = computed(
    () =>
      placed.value.filter((blip) => blip.beyondSpan).length +
      rotating.value.filter((entry) => entry.metres > activeSpan.value).length,
  );

  /**
   * Device names that can be drawn without colliding.
   *
   * This is the placement problem the feature turns on: a 20-character label is roughly half the
   * canvas width, so with a roomful of devices only a handful can ever fit. Rather than shrinking
   * the text or letting labels overlap, they are placed greedily by priority and the rest are
   * omitted — the blip and its tooltip remain, so nothing becomes unreachable.
   *
   * Priority is deliberate: the selected device always gets a name, then the nearest devices, since
   * those are what an operator is looking for. Text width is estimated rather than measured; exact
   * measurement needs a layout pass per render for a decision that only has to be about right.
   */
  const namedBlips = computed(() => {
    if (!props.showNames) return [] as typeof placed.value;
    const boxes: { x: number; y: number; width: number; height: number }[] = [];
    const shown: typeof placed.value = [];

    const ordered = [...placed.value].sort((a, b) => {
      const aSelected = a.identity.key === props.selectedKey ? 0 : 1;
      const bSelected = b.identity.key === props.selectedKey ? 0 : 1;
      if (aSelected !== bSelected) return aSelected - bSelected;
      return a.metres - b.metres;
    });

    for (const blip of ordered) {
      const box = labelBoxFor(blip);
      if (boxes.some((other) => overlaps(box, other))) continue;
      boxes.push(box);
      shown.push(blip);
    }
    return shown;
  });

  /** Where a blip's name sits, and the space it needs. */
  function labelBoxFor(blip: (typeof placed.value)[number]) {
    const text = shortLabel(blip);
    // Approximate advance width for the label font, plus padding.
    const width = text.length * 0.52 * LABEL_FONT + 1;
    const height = LABEL_FONT + 1.2;
    return { x: blip.x + 3.4, y: blip.y - 1, width, height };
  }

  /** Whether two label boxes overlap, with a margin so they never touch. */
  function overlaps(
    a: { x: number; y: number; width: number; height: number },
    b: { x: number; y: number; width: number; height: number },
  ): boolean {
    const margin = 0.8;
    return (
      a.x < b.x + b.width + margin &&
      b.x < a.x + a.width + margin &&
      a.y < b.y + b.height + margin &&
      b.y < a.y + a.height + margin
    );
  }

  /**
   * Short name for a blip.
   *
   * Truncated to a fixed budget because the label competes for a fixed amount of canvas. A MAC
   * address has to lose characters to fit, and losing them deliberately reads better than
   * overflowing into the next device.
   */
  function shortLabel(blip: (typeof placed.value)[number]): string {
    const name = blip.identity.label || blip.identity.key;
    return name.length > LABEL_MAX_CHARS ? `${name.slice(0, LABEL_MAX_CHARS - 1)}…` : name;
  }

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
  /**
   * Band boundaries for a span, innermost first.
   *
   * Produced from the span rather than read back from the rendered rings: reading the rendered set
   * ran before the canvas had recomputed, which is how a 5 m span produced a band labelled
   * "under 10 m".
   */
  function ringBandsForSpan(span: RadarSpan): { metres: number; ratio: number }[] {
    return ringsForSpan(span)
      .map((ring) => ({ metres: ring.metres, ratio: ring.ratio }))
      .sort((a, b) => a.ratio - b.ratio);
  }

  /** Which band a radial ratio falls in, given boundaries already computed for the span. */
  function bandForRatio(
    boundaries: readonly { metres: number; ratio: number }[],
    ratio: number,
  ): { id: string; label: string } {
    for (const boundary of boundaries) {
      if (ratio <= boundary.ratio) {
        return { id: `under-${boundary.metres}`, label: `under ${formatMetres(boundary.metres)}` };
      }
    }
    const outermost = boundaries[boundaries.length - 1];
    return { id: `over-${outermost?.metres ?? 0}`, label: `over ${formatMetres(outermost?.metres ?? 0)}` };
  }

  /**
   * Distance for a label, in metres.
   *
   * One formatter for ring labels and band counts. Two existed and disagreed: rings printed the
   * stored value while bands rounded, so the same boundary could read "1.4 m" on a circle and
   * "under 1 m" in the count beside it. Ring values are already rounded to two significant digits
   * where they are produced, so no further rounding is needed here.
   */
  function formatMetres(metres: number): string {
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
        {{ formatMetres(ring.metres) }}
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
      <!--
        Device names, drawn in their own layer above the blips rather than inside each interactive
        group. A <text> inside the group inflated its bounding box — the clickable centre moved off
        the dot and onto empty canvas — so labels are rendered separately and made untouchable.
      -->
      <g class="observation-radar__labels" aria-hidden="true">
        <text
          v-for="blip in namedBlips"
          :key="`name-${blip.identity.key}`"
          class="observation-radar__blip-label"
          :x="blip.x + 3.4"
          :y="blip.y + 1.4"
        >
          {{ shortLabel(blip) }}
        </text>
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
      <!--
        Unidentifiable devices, as counts rather than blips. Rendered as HTML below the canvas
        rather than as SVG text inside it: the labels were long enough to run past the viewBox edge
        and be clipped, and SVG text cannot wrap.
      -->
      <span v-if="rotatingBands.length > 0" class="observation-radar__buckets">
        <span class="observation-radar__buckets-title">
          {{ rotating.length }} with rotating addresses, by distance:
        </span>
        <span v-for="band in rotatingBands" :key="band.label" class="observation-radar__bucket">
          {{ band.count }} {{ band.label }}
        </span>
      </span>
      <span v-if="unplaced > 0" class="observation-radar__unplaced">
        {{ unplaced }} without a signal reading not shown.
      </span>
      <span v-if="beyondSpan > 0" class="observation-radar__unplaced">
        {{ beyondSpan }} beyond {{ activeSpan }} m — identified ones are drawn at the edge, the rest
        are in the counts above.
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
   * Rotating-device counts, below the canvas. Dimmer than the device list because they are
   * context: they say how much unidentifiable traffic is around without competing with the devices
   * that have names.
   */
  .observation-radar__buckets {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.25rem 0.75rem;
    margin-bottom: 0.3rem;
    font-size: 0.75rem;
    color: var(--text-subtle, #7c8b9c);
  }

  .observation-radar__buckets-title {
    font-weight: 600;
  }

  .observation-radar__bucket {
    font-variant-numeric: tabular-nums;
  }

  /*
   * Device names. Same treatment as ring labels — halo so the text survives crossing a circle or
   * another blip — but a brighter fill, because a name is the thing being looked for.
   */
  .observation-radar__blip-label {
    fill: var(--text-body, #e6edf5);
    font-size: 3.2px;
    font-weight: 600;
    paint-order: stroke;
    stroke: var(--surface-sunken, #0f1419);
    stroke-width: 1.1px;
    stroke-linejoin: round;
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
