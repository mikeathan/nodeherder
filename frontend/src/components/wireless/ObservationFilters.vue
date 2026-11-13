<script setup lang="ts">
  /*
   * Facet-driven filters for the wireless viewer (spec 064).
   *
   * The component renders whatever facets the snapshot's sources declare, so a future medium
   * (WiFi probe, USB dongle) adds a filter dimension without touching this file. Facet values
   * are derived from the identities already present, never hardcoded per provider.
   */
  import { computed } from 'vue';
  import type { ObservationIdentityView } from '@/types/observations.type';
  import { ObservationFacets, ObservationDecisions } from '@/types/observations.type';
  import { DISTANCE_RANGES, vendorHex } from '@/domain/observations/presentation';

  const props = defineProps<{
    identities: readonly ObservationIdentityView[];
    /** Facets declared by at least one source, e.g. ["kind", "vendor"]. */
    facets: readonly string[];
    /** Selected facet values per facet name; an empty list means "all". */
    selected: Readonly<Record<string, readonly string[]>>;
    /** Whether to show only untriaged identities and unstable fingerprints. */
    onlyNew: boolean;
    /** Selected distance range id; "any" applies no filter. */
    distance: string;
    /** Whether demoted classes (installed, outside) are revealed. */
    showDemoted: boolean;
  }>();

  const emit = defineEmits<{
    (event: 'update:selected', value: Record<string, string[]>): void;
    (event: 'update:onlyNew', value: boolean): void;
    (event: 'update:distance', value: string): void;
    (event: 'update:showDemoted', value: boolean): void;
  }>();

  /** Facets this component knows how to render, intersected with what sources declare. */
  const supported = computed(() =>
    props.facets.filter((facet) => facet === ObservationFacets.Kind || facet === ObservationFacets.Vendor),
  );

  /** Distinct values per supported facet, taken from the current identities. */
  const options = computed(() => {
    const byFacet = new Map<string, string[]>();
    for (const facet of supported.value) {
      const values = new Set<string>();
      for (const identity of props.identities) {
        const value = facetValue(identity, facet);
        if (value) values.add(value);
      }
      byFacet.set(facet, [...values].sort());
    }
    return byFacet;
  });

  function facetValue(identity: ObservationIdentityView, facet: string): string {
    if (facet === ObservationFacets.Kind) return identity.kind;
    // Vendors are numeric company identifiers; show the hex form people recognise.
    if (facet === ObservationFacets.Vendor) return vendorHex(identity.vendor);
    return '';
  }

  function facetLabel(facet: string): string {
    if (facet === ObservationFacets.Kind) return 'Kind';
    if (facet === ObservationFacets.Vendor) return 'Vendor';
    return facet;
  }

  function isSelected(facet: string, value: string): boolean {
    return (props.selected[facet] ?? []).includes(value);
  }

  function toggle(facet: string, value: string): void {
    // Copy every facet list into a mutable array so the emitted map satisfies the prop type
    // even though the incoming selection is readonly.
    const next: Record<string, string[]> = {};
    for (const [name, values] of Object.entries(props.selected)) next[name] = [...values];
    const current = next[facet] ?? [];
    next[facet] = current.includes(value)
      ? current.filter((entry) => entry !== value)
      : [...current, value];
    emit('update:selected', next);
  }

  function toggleOnlyNew(): void {
    emit('update:onlyNew', !props.onlyNew);
  }

  function toggleShowDemoted(): void {
    emit('update:showDemoted', !props.showDemoted);
  }
</script>

<template>
  <div class="observation-filters">
    <button
      type="button"
      class="observation-filters__toggle"
      :class="{ 'is-active': onlyNew }"
      :aria-pressed="onlyNew"
      @click="toggleOnlyNew"
    >
      Only new
    </button>

    <!--
      The complement of the default filter. The default hides installed equipment and devices that
      look like neighbours, which is most of the census; this reveals them again so nothing is ever
      hidden without a way back.
    -->
    <button
      type="button"
      class="observation-filters__toggle"
      :class="{ 'is-active': showDemoted }"
      :aria-pressed="showDemoted"
      @click="toggleShowDemoted"
    >
      {{ showDemoted ? 'Hiding nothing' : 'Show all kinds' }}
    </button>

    <!--
      Distance bands, not open-ended limits: "within 20 m" covers almost every device and so
      isolates nothing. Bands answer "which of these is in the next room".
    -->
    <fieldset class="observation-filters__group">
      <legend>Distance</legend>
      <div class="observation-filters__presets">
        <button
          v-for="range in DISTANCE_RANGES"
          :key="range.id"
          type="button"
          class="observation-filters__preset"
          :class="{ 'is-active': props.distance === range.id }"
          :aria-pressed="props.distance === range.id"
          @click="emit('update:distance', range.id)"
        >
          {{ range.label }}
        </button>
      </div>
    </fieldset>

    <fieldset v-for="facet in supported" :key="facet" class="observation-filters__group">
      <legend>{{ facetLabel(facet) }}</legend>
      <label v-for="value in options.get(facet) ?? []" :key="value" class="observation-filters__option">
        <input
          type="checkbox"
          :checked="isSelected(facet, value)"
          @change="toggle(facet, value)"
        />
        <span>{{ value }}</span>
      </label>
    </fieldset>

    <p v-if="supported.length === 0" class="observation-filters__empty">
      No source declares filterable facets yet.
    </p>
    <p class="observation-filters__hint" aria-live="polite">
      {{ identities.length }} shown
      <span v-if="onlyNew">· {{ identities.filter((i) => i.decision === ObservationDecisions.Unknown).length }} new</span>
    </p>
  </div>
</template>

<style scoped>
  .observation-filters {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-items: flex-start;
  }

  .observation-filters__toggle {
    border: 1px solid var(--border-subtle, #2a3441);
    background: transparent;
    color: inherit;
    border-radius: 999px;
    padding: 0.25rem 0.875rem;
    cursor: pointer;
    font: inherit;
  }

  .observation-filters__toggle.is-active {
    border-color: var(--accent-primary, #4a9eff);
    color: var(--accent-primary, #4a9eff);
  }

  .observation-filters__toggle:focus-visible {
    outline: 2px solid var(--accent-primary, #4a9eff);
    outline-offset: 2px;
  }

  .observation-filters__group {
    border: 0;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
  }

  .observation-filters__group legend {
    color: var(--text-muted, #8b9bab);
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 0 0.25rem;
  }

  .observation-filters__option {
    display: inline-flex;
    gap: 0.25rem;
    align-items: center;
    font-size: 0.875rem;
    cursor: pointer;
  }

  /* A segmented group rather than checkboxes: these options are mutually exclusive. */
  .observation-filters__presets {
    display: inline-flex;
    border: 1px solid var(--border-subtle, #2a3441);
    border-radius: 999px;
    overflow: hidden;
  }

  .observation-filters__preset {
    background: transparent;
    border: 0;
    color: inherit;
    font: inherit;
    font-size: 0.8125rem;
    padding: 0.25rem 0.75rem;
    cursor: pointer;
    white-space: nowrap;
  }

  .observation-filters__preset + .observation-filters__preset {
    border-left: 1px solid var(--border-subtle, #2a3441);
  }

  .observation-filters__preset.is-active {
    background: var(--accent-primary, #4a9eff);
    color: #04121f;
  }

  .observation-filters__preset:focus-visible {
    outline: 2px solid var(--accent-primary, #4a9eff);
    outline-offset: -2px;
  }

  .observation-filters__hint,
  .observation-filters__empty {
    color: var(--text-muted, #8b9bab);
    font-size: 0.8125rem;
    margin: 0;
  }
</style>
