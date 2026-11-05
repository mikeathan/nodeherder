<script setup lang="ts">
  /* Static spatial help topics and diagrams; no device data or configuration is changed. */
  import { computed } from 'vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import { SPATIAL_MAX_AGE_MS } from '@/domain/spatial';
  import { MILLISECONDS_PER_SECOND } from '@/domain/spatial/analytics';
  import { SPATIAL_HOLD_MS, SPATIAL_MERGE_MM, SPATIAL_SILENCE_MS } from '@/domain/spatial/tracks';

  defineProps<{ deviceId: string }>();
  const freshnessSeconds = computed(() => Math.round(SPATIAL_MAX_AGE_MS / MILLISECONDS_PER_SECOND));
  const holdSeconds = computed(() => Math.round(SPATIAL_HOLD_MS / MILLISECONDS_PER_SECOND));
  const silenceSeconds = computed(() => Math.round(SPATIAL_SILENCE_MS / MILLISECONDS_PER_SECOND));
  const mergeMetres = computed(() => (SPATIAL_MERGE_MM / 1000).toFixed(2));
</script>

<template>
  <div class="nh-spatial-help-layout">
    <nav class="nh-spatial-help-nav" aria-label="Spatial help topics">
      <a :href="`#spatial-mapping-${deviceId}`">Map a target</a>
      <a :href="`#spatial-plane-${deviceId}`">Read the plane</a>
      <a :href="`#spatial-tracks-${deviceId}`">Tracks, smoothing and limits</a>
      <a :href="`#spatial-troubleshooting-${deviceId}`">If no target appears</a>
    </nav>
    <div class="nh-spatial-help-topics">
      <UiCard :id="`spatial-mapping-${deviceId}`" title="Map a target" :heading-level="3" tabindex="-1">
        <ol>
          <li>Open <strong>Configure target mapping</strong> in the spatial view.</li>
          <li>
            When the device labels its fields by target number, each row shows the matching fields. Other naming styles
            can be mapped manually.
          </li>
          <li>
            For each target, select its matching X, Y and Active entities from the device. For numbered labels, choose
            the same number for all three and the matching row. Leave unused targets blank.
          </li>
          <li>
            Select <strong>Save spatial settings</strong>. These settings belong to NodeHerder; they do not change the
            device.
          </li>
        </ol>
        <figure>
          <svg
            viewBox="0 0 540 156"
            role="img"
            aria-label="Example: Target 1 X, Target 1 Y and Target 1 Active mapped to the matching fields">
            <rect x="4" y="4" width="532" height="148" rx="12" class="nh-diagram-frame" />
            <text x="20" y="32" class="nh-diagram-title">Target 1 · example mapping</text>
            <g class="nh-diagram-label">
              <text x="20" y="60">X</text>
              <text x="192" y="60">Y</text>
              <text x="364" y="60">Active</text>
            </g>
            <g class="nh-diagram-box">
              <rect x="20" y="70" width="156" height="47" rx="7" />
              <rect x="192" y="70" width="156" height="47" rx="7" />
              <rect x="364" y="70" width="156" height="47" rx="7" />
            </g>
            <g class="nh-diagram-text">
              <text x="30" y="99">Target 1 X</text>
              <text x="202" y="99">Target 1 Y</text>
              <text x="374" y="99">Target 1 Active</text>
            </g>
          </svg>
          <figcaption>Choose fields that describe the same target. Names depend on the device.</figcaption>
        </figure>
      </UiCard>
      <UiCard :id="`spatial-plane-${deviceId}`" title="Read the plane" :heading-level="3" tabindex="-1">
        <p>
          The Live dashboard combines room context, tracks and geometric zone previews. Recent trails show sampled
          movement while this view is open; they are not recorded history.
        </p>
        <p>
          The sensor dot marks its position from Room Builder. Each dot represents a track; the accompanying list
          shows its label, state and coordinates.
        </p>
        <figure>
          <svg
            viewBox="0 0 440 270"
            role="img"
            aria-label="Example coordinate plane with the sensor at the centre, positive X to the right and positive Y upwards">
            <rect x="20" y="15" width="240" height="240" rx="10" class="nh-diagram-frame" />
            <path d="M80 15v240 M140 15v240 M200 15v240 M20 75h240 M20 135h240 M20 195h240" class="nh-diagram-grid" />
            <path d="M35 135h210 M140 240V30" class="nh-diagram-axis" />
            <circle cx="140" cy="135" r="6" class="nh-diagram-sensor" />
            <circle cx="190" cy="83" r="9" class="nh-diagram-target" />
            <text x="205" y="86" class="nh-diagram-text">1</text>
            <g class="nh-diagram-text">
              <text x="116" y="37">+Y</text>
              <text x="225" y="125">+X</text>
              <text x="35" y="125">−X</text>
              <text x="275" y="100">1 · target</text>
              <text x="275" y="140">● · sensor</text>
            </g>
          </svg>
          <figcaption>
            Illustration only. Scale changes the displayed range; orientation rotates the view. Targets and overlays
            share the same sensor placement.
          </figcaption>
        </figure>
      </UiCard>
      <p>
        Use Zone editor to place detection, exclusion and entry overlays; Room Builder for walls, doors, furniture and
        sensor placement; Entry / exit for reported assumed presence; and Heatmap &amp; analytics for session
        observations. Each tool has its own help. A zone overlay does not configure your device.
      </p>
      <p>
        <a
          href="https://docs.everythingsmart.io/s/products/doc/695636da-dc42-490b-b5a1-281e96c73381"
          target="_blank"
          rel="noopener noreferrer"
          >Manufacturer Live Dashboard guide</a
        >
      </p>
      <UiCard :id="`spatial-tracks-${deviceId}`" title="Tracks, smoothing and limits" :heading-level="3" tabindex="-1">
        <p>
          The device reports up to three numbered <strong>slots</strong>, not people. A slot is reused when someone
          leaves, so the view keeps its own <strong>track</strong> label while a position stays near the previous one
          and never claims to identify a person.
        </p>
        <dl class="nh-spatial-states">
          <dt>Live</dt>
          <dd>A position the sensor just reported.</dd>
          <dt>Predicted</dt>
          <dd>A jump too large to be a person was rejected; the last known position is held briefly.</dd>
          <dt>Held</dt>
          <dd>A slot cleared; its last position stays visible for about {{ holdSeconds }} s, then disappears.</dd>
          <dt>Idle</dt>
          <dd>No new position for {{ silenceSeconds }} s. The device may simply be quiet because nothing moved.</dd>
        </dl>
        <p>
          <strong>Smoothing</strong> trades calm for lag: higher levels steady a standing dot but follow a fast walk
          less closely. The <strong>Raw</strong> toggle draws the unfiltered sensor reading from the same frames, and
          the jitter readout reports measured movement while stationary. Smoothing changes the drawing only; it does not
          improve what the sensor measured.
        </p>
        <p>
          Limits worth remembering: coordinates are best-effort and vary with mounting and room clutter; distance, angle
          and resolution are derived by the firmware, so they add no independent accuracy; duplicate suppression can hide
          a second person standing within about {{ mergeMetres }} m; and zone previews are overlays, not firmware occupancy
          unless you mapped a reported entity.
        </p>
      </UiCard>
      <UiCard :id="`spatial-troubleshooting-${deviceId}`" title="If no target appears" :heading-level="3" tabindex="-1">
        <ul>
          <li>
            Walk through the room and compare target positions with known furniture. If positions are offset, check
            sensor placement and direction in Room Builder.
          </li>
          <li>Check that the device and hub are online and that the mapping is saved.</li>
          <li>
            Check that X, Y and Active refer to the same target, and check the device's own tracking settings. A target
            is shown only when Active reports true.
          </li>
          <li>
            Positions come from synchronized frames that carry both axes together. When a server sends only sampled
            coordinates, readings older than {{ freshnessSeconds }} seconds are hidden rather than shown as live.
          </li>
          <li>Active uses the last reported state while connected. Some devices report it only when it changes.</li>
          <li>
            The status strip says which source is in use. Sampled updates can mix axes; frames cannot.
          </li>
        </ul>
      </UiCard>
    </div>
  </div>
</template>

<style scoped>
  .nh-spatial-states {
    margin: 0.5rem 0 0.75rem;
  }
  .nh-spatial-states dt {
    font-weight: 600;
  }
  .nh-spatial-states dd {
    margin: 0 0 0.4rem 1rem;
    color: var(--nh-text-muted);
  }
  .nh-spatial-help-layout {
    display: grid;
    grid-template-columns: minmax(10rem, 15rem) minmax(0, 1fr);
    gap: 1rem;
    margin-block: 1rem;
  }
  .nh-spatial-help-nav,
  .nh-spatial-help-topics {
    display: grid;
    align-content: start;
    gap: 1rem;
    min-width: 0;
  }
  .nh-spatial-help-nav a {
    display: flex;
    align-items: center;
    min-height: 44px;
  }
  .nh-spatial-help-topics :deep(.nh-card) {
    scroll-margin-top: 1rem;
  }
  ol,
  ul {
    padding-left: 1.25rem;
    line-height: 1.7;
  }
  figure {
    margin: 1rem 0 0;
    max-width: 34rem;
  }
  svg {
    display: block;
    width: 100%;
    height: auto;
    color: var(--nh-text);
  }
  figcaption {
    color: var(--nh-text-muted);
    font-size: 0.85rem;
    margin-top: 0.5rem;
  }
  .nh-diagram-frame {
    fill: var(--nh-surface-2);
    stroke: var(--nh-border);
  }
  .nh-diagram-grid {
    fill: none;
    stroke: var(--nh-border-strong);
  }
  .nh-diagram-axis {
    fill: none;
    stroke: var(--nh-text-muted);
    stroke-width: 1.5;
  }
  .nh-diagram-box {
    fill: var(--nh-surface);
    stroke: var(--nh-border-strong);
  }
  .nh-diagram-title,
  .nh-diagram-label,
  .nh-diagram-text {
    fill: currentColor;
    font-family: inherit;
    font-size: 15px;
  }
  .nh-diagram-title {
    font-weight: 600;
  }
  .nh-diagram-sensor {
    fill: var(--nh-text-muted);
  }
  .nh-diagram-target {
    fill: var(--nh-accent);
  }
  @media (max-width: 640px) {
    .nh-spatial-help-layout {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
