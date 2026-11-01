<script setup lang="ts">
  /* Small inline diagrams for the spatial help topics; purely illustrative. */
  import { SPATIAL_HELP_ILLUSTRATIONS, type SpatialHelpTopic } from '@/domain/spatial/view';

  defineProps<{ topic: SpatialHelpTopic }>();
</script>

<template>
  <figure class="nh-help-figure">
    <svg v-if="topic === 'zones'" viewBox="0 0 340 132" role="img" :aria-label="SPATIAL_HELP_ILLUSTRATIONS.zones">
      <rect x="8" y="8" width="324" height="116" rx="14" class="frame" />
      <rect x="34" y="34" width="98" height="62" rx="10" class="detection" />
      <rect x="154" y="52" width="66" height="44" rx="9" class="exclusion" />
      <path d="M246 92V42h64v50z" class="entry" />
      <circle cx="176" cy="104" r="6" class="sensor" />
      <text x="52" y="70" class="label">Detection</text>
      <text x="163" y="78" class="label small">Exclusion</text>
      <text x="258" y="70" class="label small">Entry</text>
    </svg>
    <svg v-else-if="topic === 'room'" viewBox="0 0 340 132" role="img" :aria-label="SPATIAL_HELP_ILLUSTRATIONS.room">
      <rect x="8" y="8" width="324" height="116" rx="14" class="frame" />
      <path d="M38 104V30h180v74" class="outline" />
      <path d="M132 104 112 56h40z" class="fov" />
      <circle cx="132" cy="104" r="6" class="sensor" />
      <rect x="214" y="58" width="56" height="26" rx="6" class="furniture" />
      <rect x="282" y="34" width="12" height="70" rx="4" class="wall" />
      <text x="48" y="50" class="label small">Room outline</text>
      <text x="206" y="100" class="label small">Furniture</text>
    </svg>
    <svg v-else-if="topic === 'entry'" viewBox="0 0 340 132" role="img" :aria-label="SPATIAL_HELP_ILLUSTRATIONS.entry">
      <rect x="8" y="8" width="324" height="116" rx="14" class="frame" />
      <path d="M96 112V20h34v92M210 112V20h34v92" class="wall" />
      <path d="M130 66h80" class="arrow" marker-end="url(#entry-arrow)" />
      <defs>
        <marker id="entry-arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M0 0 10 5 0 10z" class="arrow-head" />
        </marker>
      </defs>
      <circle cx="252" cy="66" r="24" class="hold" />
      <text x="238" y="70" class="label small">hold</text>
      <text x="42" y="116" class="label small">Doorway</text>
    </svg>
    <svg v-else viewBox="0 0 340 132" role="img" :aria-label="SPATIAL_HELP_ILLUSTRATIONS.heatmap">
      <rect x="8" y="8" width="324" height="116" rx="14" class="frame" />
      <defs>
        <linearGradient id="heat-scale" x1="0" x2="1" y1="0" y2="0">
          <stop offset="0%" stop-color="#3f6fb5" />
          <stop offset="45%" stop-color="#4f8bd8" />
          <stop offset="70%" stop-color="#e0b44a" />
          <stop offset="100%" stop-color="#d95a4e" />
        </linearGradient>
      </defs>
      <g class="heat">
        <rect x="38" y="34" width="38" height="38" rx="4" />
        <rect x="78" y="34" width="38" height="38" rx="4" class="warm" />
        <rect x="118" y="34" width="38" height="38" rx="4" class="hot" />
        <rect x="38" y="74" width="38" height="38" rx="4" class="cool" />
        <rect x="78" y="74" width="38" height="38" rx="4" />
        <rect x="118" y="74" width="38" height="38" rx="4" class="warm" />
      </g>
      <rect x="196" y="58" width="112" height="14" rx="7" fill="url(#heat-scale)" />
      <text x="198" y="48" class="label small">Low</text>
      <text x="278" y="48" class="label small">High</text>
      <text x="196" y="98" class="label small">Target-seconds</text>
    </svg>
    <figcaption>Illustration only. Overlays never change firmware or automations.</figcaption>
  </figure>
</template>

<style scoped>
  .nh-help-figure {
    margin: 0 0 0.85rem;
    max-width: 30rem;
  }
  svg {
    display: block;
    width: 100%;
    height: auto;
  }
  .frame {
    fill: var(--nh-surface-2);
    stroke: var(--nh-border);
  }
  .detection {
    fill: color-mix(in srgb, var(--nh-spatial-detection) 18%, transparent);
    stroke: var(--nh-spatial-detection);
    stroke-width: 2;
  }
  .exclusion {
    fill: color-mix(in srgb, var(--nh-spatial-exclusion) 16%, transparent);
    stroke: var(--nh-spatial-exclusion);
    stroke-dasharray: 5 3;
    stroke-width: 2;
  }
  .entry {
    fill: color-mix(in srgb, var(--nh-spatial-entry) 16%, transparent);
    stroke: var(--nh-spatial-entry);
    stroke-dasharray: 8 3;
    stroke-width: 2;
  }
  .sensor {
    fill: var(--nh-text);
  }
  .outline {
    fill: none;
    stroke: var(--nh-text-muted);
    stroke-width: 3;
  }
  .fov {
    fill: color-mix(in srgb, var(--nh-accent) 18%, transparent);
    stroke: var(--nh-accent);
    stroke-dasharray: 5 4;
  }
  .furniture {
    fill: color-mix(in srgb, var(--nh-border-strong) 65%, transparent);
    stroke: var(--nh-text-muted);
  }
  .wall {
    fill: var(--nh-text-muted);
  }
  .arrow {
    stroke: var(--nh-accent);
    stroke-width: 3;
  }
  .arrow-head {
    fill: var(--nh-accent);
  }
  .hold {
    fill: color-mix(in srgb, var(--nh-spatial-entry) 16%, transparent);
    stroke: var(--nh-spatial-entry);
    stroke-dasharray: 4 3;
    stroke-width: 2;
  }
  .heat rect {
    fill: #4f8bd8;
  }
  .heat .cool {
    fill: #3f6fb5;
  }
  .heat .warm {
    fill: #e0b44a;
  }
  .heat .hot {
    fill: #d95a4e;
  }
  text {
    fill: var(--nh-text);
    font-family: inherit;
    font-size: 13px;
  }
  .label.small {
    fill: var(--nh-text-muted);
    font-size: 11px;
  }
  figcaption {
    color: var(--nh-text-muted);
    font-size: 0.8rem;
    margin-top: 0.35rem;
  }
</style>
