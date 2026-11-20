<script setup lang="ts">
  /*
   * Filters for the wireless viewer: class presets, triage/visibility toggles, distance bands,
   * source facets and a free-text search (spec 066, extending spec 064).
   *
   * Every control is a labelled field so the label line, control height and baseline are shared;
   * the previous layout mixed a labelled fieldset with a bare button, which is why one control
   * floated out of line. On a phone the fields sit behind a disclosure to keep the noise down,
   * while search stays visible.
   *
   * Class presets are not a source facet: they come from the advisory classification every
   * identity carries, and the same ordered table also drives row order and List sections. Every
   * group is always rendered in the same order, with a fixed-width count, so a changing count
   * cannot make a chip appear, disappear or change width.
   */
  import { computed, ref } from 'vue';
  import type { ObservationIdentityView } from '@/types/observations.type';
  import { ObservationFacets } from '@/types/observations.type';
  import {
    ANY_DISTANCE_RANGE,
    DISTANCE_RANGES,
    isUsefulFacet,
    kindLabel,
    vendorHex,
  } from '@/domain/observations/presentation';
  import type { ClassGroup, ObservationClassFilter } from '@/domain/observations/classes';
  import { ANY_CLASS_GROUP } from '@/domain/observations/classes';

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
    /** Selected class preset; "any" applies no class restriction. */
    classFilter: ObservationClassFilter;
    /** Every class preset with its count, in fixed display order. */
    classOptions: readonly (ClassGroup & { count: number })[];
    /** Free-text query over title, key and vendor. */
    search: string;
    /** Whether demoted classes are revealed. */
    showDemoted: boolean;
    /**
     * Whether the reveal control is meaningful.
     *
     * The toggle is hidden when false. Before classification has run nothing is classed as
     * installed equipment, so the control changed nothing and read as broken — a filter that cannot
     * do anything is worse than no filter. It is also hidden while a class preset is active, since
     * that preset already overrides default hiding.
     */
    hasDemoted: boolean;
  }>();

  const emit = defineEmits<{
    (event: 'update:selected', value: Record<string, string[]>): void;
    (event: 'update:onlyNew', value: boolean): void;
    (event: 'update:distance', value: string): void;
    (event: 'update:classFilter', value: ObservationClassFilter): void;
    (event: 'update:search', value: string): void;
    (event: 'update:showDemoted', value: boolean): void;
  }>();

  /** Mobile disclosure state. Ignored on desktop, where the fields are always shown. */
  const filtersOpen = ref(false);

  /** How many narrowing filters are active, so a collapsed panel still says it is doing something. */
  const activeFilterCount = computed(() => {
    let count = 0;
    if (props.classFilter !== ANY_CLASS_GROUP) count += 1;
    if (props.onlyNew) count += 1;
    if (props.distance !== ANY_DISTANCE_RANGE.id) count += 1;
    for (const values of Object.values(props.selected)) {
      if (values.length > 0) count += 1;
    }
    return count;
  });

  /**
   * Facets this component knows how to render, intersected with what sources declare.
   *
   * A facet is then dropped when it has only one possible value, because such a filter cannot do
   * anything: with one option, both states match everything. In an ordinary household that is
   * exactly the Kind facet, since every device is `adv`. The declaration is still honoured — this
   * only decides whether a control is worth showing today.
   */
  const supported = computed(() =>
    props.facets.filter(
      (facet) =>
        (facet === ObservationFacets.Kind || facet === ObservationFacets.Vendor) &&
        isUsefulFacet(valuesFor(facet)),
    ),
  );

  /** Distinct values for one facet, taken from the current identities. */
  function valuesFor(facet: string): string[] {
    const values = new Set<string>();
    for (const identity of props.identities) {
      const value = facetValue(identity, facet);
      if (value) values.add(value);
    }
    return [...values].sort();
  }

  /** Distinct values per supported facet, memoised per render. */
  const options = computed(() => {
    const byFacet = new Map<string, string[]>();
    for (const facet of supported.value) {
      byFacet.set(facet, valuesFor(facet));
    }
    return byFacet;
  });

  function facetValue(identity: ObservationIdentityView, facet: string): string {
    if (facet === ObservationFacets.Kind) return identity.kind;
    // Vendors are numeric company identifiers; show the hex form people recognise.
    if (facet === ObservationFacets.Vendor) return vendorHex(identity.vendor);
    return '';
  }

  /** Readable label for a facet value: `adv` reads as "Bluetooth", not as jargon. */
  function facetValueLabel(facet: string, value: string): string {
    return facet === ObservationFacets.Kind ? kindLabel(value) : value;
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

  function onSearch(event: Event): void {
    emit('update:search', (event.target as HTMLInputElement).value);
  }

  function clearSearch(): void {
    emit('update:search', '');
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
    <div class="observation-filters__search">
      <label class="observation-filters__label" for="wireless-observation-search">Search</label>
      <div class="observation-filters__search-box">
        <input
          id="wireless-observation-search"
          type="search"
          class="observation-filters__search-input"
          :value="props.search"
          placeholder="Name, address or vendor"
          @input="onSearch"
        />
        <button
          v-if="props.search"
          type="button"
          class="observation-filters__clear"
          aria-label="Clear search"
          @click="clearSearch"
        >
          Clear
        </button>
      </div>
    </div>

    <!--
      On a phone the labelled fields collapse behind this control so the page is not a stack of
      unrelated groups. The count keeps an active filter discoverable while collapsed.
    -->
    <button
      type="button"
      class="observation-filters__disclosure"
      :class="{ 'is-active': filtersOpen }"
      :aria-expanded="filtersOpen"
      aria-controls="wireless-filter-fields"
      @click="filtersOpen = !filtersOpen"
    >
      Filters
      <span v-if="activeFilterCount > 0" class="observation-filters__badge">{{ activeFilterCount }}</span>
    </button>

    <div
      id="wireless-filter-fields"
      class="observation-filters__fields"
      :class="{ 'is-open': filtersOpen }"
    >
      <fieldset class="observation-filters__field">
        <legend>Class</legend>
        <div class="observation-filters__chips">
          <button
            type="button"
            class="observation-filters__chip"
            :class="{ 'is-active': props.classFilter === ANY_CLASS_GROUP }"
            :aria-pressed="props.classFilter === ANY_CLASS_GROUP"
            data-group="any"
            aria-label="Any class"
            @click="emit('update:classFilter', ANY_CLASS_GROUP)"
          >
            Any
          </button>
          <!--
            Every group stays in the same place with a fixed-width count. Omitting an empty group
            or letting a count gain a digit made the row shuffle while the operator was reading it.
          -->
          <button
            v-for="option in props.classOptions"
            :key="option.id"
            type="button"
            class="observation-filters__chip"
            :class="{ 'is-active': props.classFilter === option.id }"
            :disabled="option.count === 0 && props.classFilter !== option.id"
            :aria-pressed="props.classFilter === option.id"
            :data-group="option.id"
            :aria-label="`${option.label}, ${option.count} devices`"
            @click="emit('update:classFilter', option.id)"
          >
            {{ option.label }}
            <span class="observation-filters__count">{{ option.count }}</span>
          </button>
        </div>
      </fieldset>

      <fieldset class="observation-filters__field">
        <legend>Status</legend>
        <div class="observation-filters__chips">
          <button
            type="button"
            class="observation-filters__chip"
            :class="{ 'is-active': onlyNew }"
            :aria-pressed="onlyNew"
            @click="toggleOnlyNew"
          >
            Needs triage
          </button>
          <!--
            The complement of the default filter: the default hides installed equipment, and this
            reveals it so nothing is ever hidden without a way back. Only offered when something is
            actually demoted, because otherwise the control toggles between identical states.
          -->
          <button
            v-if="hasDemoted"
            type="button"
            class="observation-filters__chip"
            :class="{ 'is-active': showDemoted }"
            :aria-pressed="showDemoted"
            @click="toggleShowDemoted"
          >
            {{ showDemoted ? 'Hiding nothing' : 'Show all kinds' }}
          </button>
        </div>
      </fieldset>

      <!--
        Distance bands, not open-ended limits: "within 20 m" covers almost every device and so
        isolates nothing. Bands answer "which of these is in the next room".
      -->
      <fieldset class="observation-filters__field">
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

      <fieldset v-for="facet in supported" :key="facet" class="observation-filters__field">
        <legend>{{ facetLabel(facet) }}</legend>
        <div class="observation-filters__chips">
          <label
            v-for="value in options.get(facet) ?? []"
            :key="value"
            class="observation-filters__option"
          >
            <input
              type="checkbox"
              :checked="isSelected(facet, value)"
              @change="toggle(facet, value)"
            />
            <span>{{ facetValueLabel(facet, value) }}</span>
          </label>
        </div>
      </fieldset>
    </div>
  </div>
</template>

<style scoped>
  .observation-filters {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  /* One label rhythm for every field: the control below it lands on the same baseline. */
  .observation-filters__label,
  .observation-filters__field legend {
    margin: 0;
    padding: 0;
    color: var(--nh-text-muted);
    font-size: 0.7rem;
    line-height: 1.2;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .observation-filters__search {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    width: 100%;
  }

  .observation-filters__search-box {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    min-height: 2.25rem;
    padding: 0 0.6rem;
    border: 1px solid var(--nh-border);
    border-radius: 999px;
    background: var(--nh-surface);
  }

  .observation-filters__search-box:focus-within {
    outline: 2px solid var(--nh-accent);
    outline-offset: 1px;
  }

  .observation-filters__search-input {
    flex: 1;
    min-width: 0;
    background: transparent;
    border: 0;
    color: inherit;
    font: inherit;
    font-size: 0.8125rem;
    padding: 0.25rem 0;
  }

  .observation-filters__search-input:focus-visible {
    outline: none;
  }

  .observation-filters__clear {
    background: none;
    border: 0;
    color: var(--nh-accent);
    font: inherit;
    font-size: 0.75rem;
    padding: 0;
    cursor: pointer;
  }

  .observation-filters__clear:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: 2px;
    border-radius: 0.25rem;
  }

  .observation-filters__disclosure {
    display: inline-flex;
    align-items: center;
    align-self: flex-start;
    gap: 0.4rem;
    min-height: 2.25rem;
    padding: 0 0.9rem;
    border: 1px solid var(--nh-border);
    border-radius: 999px;
    background: var(--nh-surface);
    color: inherit;
    font: inherit;
    font-size: 0.8125rem;
    cursor: pointer;
  }

  .observation-filters__disclosure.is-active {
    border-color: var(--nh-accent);
    color: var(--nh-accent);
  }

  .observation-filters__disclosure:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: 2px;
  }

  .observation-filters__badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.25rem;
    min-height: 1.25rem;
    padding: 0 0.3rem;
    border-radius: 999px;
    background: var(--nh-accent);
    color: var(--nh-accent-contrast);
    font-size: 0.65rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }

  .observation-filters__fields {
    display: none;
    flex-direction: column;
    gap: 0.75rem;
  }

  .observation-filters__fields.is-open {
    display: flex;
  }

  .observation-filters__field {
    display: flex;
    flex-direction: column;
    gap: 0;
    border: 0;
    margin: 0;
    padding: 0;
    min-width: 0;
  }

  /*
   * A fieldset legend is not a flex item in Chromium, so `gap` does not space it from the control
   * and it sat 5.6px tighter than the search label. An explicit margin makes every field's
   * label-to-control distance identical.
   */
  .observation-filters__field legend {
    margin-bottom: 0.35rem;
  }

  .observation-filters__chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
    align-items: center;
  }

  .observation-filters__chip {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    min-height: 2.25rem;
    padding: 0 0.75rem;
    border: 1px solid var(--nh-border);
    border-radius: 999px;
    background: transparent;
    color: inherit;
    font: inherit;
    font-size: 0.8125rem;
    white-space: nowrap;
    cursor: pointer;
  }

  .observation-filters__chip.is-active {
    background: var(--nh-accent);
    border-color: var(--nh-accent);
    color: var(--nh-accent-contrast);
  }

  .observation-filters__chip:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .observation-filters__chip:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: 2px;
  }

  /*
   * Fixed slot for the count: tabular digits plus a reserved width mean 9 -> 10 does not resize
   * the chip, so the row cannot shift while it is being read.
   */
  .observation-filters__count {
    display: inline-block;
    min-width: 3ch;
    text-align: right;
    font-variant-numeric: tabular-nums;
    opacity: 0.75;
  }

  .observation-filters__option {
    display: inline-flex;
    gap: 0.3rem;
    align-items: center;
    min-height: 2.25rem;
    font-size: 0.8125rem;
    cursor: pointer;
  }

  /* A segmented group rather than checkboxes: these options are mutually exclusive. */
  .observation-filters__presets {
    display: inline-flex;
    border: 1px solid var(--nh-border);
    border-radius: 999px;
    overflow: hidden;
    max-width: 100%;
  }

  .observation-filters__preset {
    background: transparent;
    border: 0;
    color: inherit;
    font: inherit;
    font-size: 0.8125rem;
    padding: 0.35rem 0.75rem;
    cursor: pointer;
    white-space: nowrap;
  }

  .observation-filters__preset + .observation-filters__preset {
    border-left: 1px solid var(--nh-border);
  }

  .observation-filters__preset.is-active {
    background: var(--nh-accent);
    color: var(--nh-accent-contrast);
  }

  .observation-filters__preset:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: -2px;
  }

  /*
   * On a phone the joined segment is wider than the column, which clipped the far bands. Wrapping
   * them as individual pills keeps every option reachable and matches the Class chips.
   */
  @media (max-width: 899px) {
    .observation-filters__presets {
      flex-wrap: wrap;
      gap: 0.35rem;
      border: 0;
      border-radius: 0;
      overflow: visible;
    }

    .observation-filters__preset {
      min-height: 2.25rem;
      border: 1px solid var(--nh-border);
      border-radius: 999px;
    }

    .observation-filters__preset + .observation-filters__preset {
      border-left: 1px solid var(--nh-border);
    }
  }

  /*
   * From tablet up the fields are always visible and share one wrapping row with the search. The
   * disclosure has nothing to disclose there.
   */
  @media (min-width: 900px) {
    .observation-filters {
      flex-direction: row;
      flex-wrap: wrap;
      align-items: flex-start;
      gap: 0.75rem 1.25rem;
    }

    .observation-filters__search {
      flex: 0 1 22rem;
      width: auto;
    }

    .observation-filters__disclosure {
      display: none;
    }

    .observation-filters__fields {
      display: contents;
    }
  }
</style>
