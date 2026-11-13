/**
 * Viewer state for the wireless pane: filtering, isolation, freezing and view selection.
 *
 * Extracted from the component so this logic is unit-testable without mounting a Vue component.
 * The component keeps only rendering and event wiring, matching the repository's separation of
 * pure view logic into `domain/` and shared behaviour into `composables/`.
 */
import { computed, ref, watch } from 'vue';
import type {
  ObservationIdentityView,
  ObservationSnapshot,
  ObservationViewMode,
} from '@/types/observations.type';
import { ObservationFacets, ObservationViews } from '@/types/observations.type';
import {
  ANY_DISTANCE_RANGE,
  shownByDefault,
  distanceRangeById,
  groupIdentities,
  isRecentArrival,
  withinDistanceRange,
  vendorHex,
  type RadarSpan,
} from '@/domain/observations/presentation';

/** Labels for the view tabs, kept beside the union so a new view cannot be added silently. */
export const OBSERVATION_VIEW_LABELS: Readonly<Record<ObservationViewMode, string>> = {
  radar: 'Radar',
  list: 'List',
  focus: 'Focus',
};

/**
 * Build the viewer state.
 *
 * `snapshot` is the live hub state. Everything else is operator intent, which is deliberately
 * component-local: a persisted filter that survives a reload would leave the operator looking
 * at a silently narrowed view with no obvious cause.
 */
export function useObservationViewer(snapshot: () => ObservationSnapshot | null) {
  /** Facet filter selections, distinct from the isolated identity. */
  const facetSelection = ref<Record<string, string[]>>({});
  /** Show only identities that appeared within the recent window. */
  const onlyNew = ref(false);
  /**
   * Whether demoted classes (installed and outside) are revealed.
   *
   * Off by default, which is the whole point of the classification work: the operator's complaint
   * was a wall of MAC addresses, and the majority of that wall is equipment and neighbours. The
   * toggle exists so nothing is ever hidden without a way to see it.
   */
  const showDemoted = ref(false);
  /** Selected distance band id; the "any" range applies no filter. */
  const distanceRange = ref<string>(ANY_DISTANCE_RANGE.id);
  /** Radar span in metres, or null to scale automatically to the data. */
  const radarSpan = ref<RadarSpan | null>(null);
  const viewMode = ref<ObservationViewMode>(ObservationViews.Radar);

  /** Live identities exactly as the hub last reported them. */
  const liveIdentities = computed<ObservationIdentityView[]>(() => snapshot()?.identities ?? []);

  /**
   * Frozen identities, captured when paused. A copy is kept rather than updates being ignored,
   * so triage actions still land on what is on screen.
   */
  const frozen = ref<ObservationIdentityView[] | null>(null);
  const paused = computed(() => frozen.value !== null);

  /** Identities to render: the frozen copy while paused, otherwise the live snapshot. */
  const identities = computed(() => frozen.value ?? liveIdentities.value);

  function togglePause(): void {
    frozen.value = paused.value ? null : liveIdentities.value.map((identity) => ({ ...identity }));
  }

  /** Facets declared by any source; the filter component renders only these. */
  const facets = computed(() => {
    const declared = new Set<string>();
    for (const source of snapshot()?.sources ?? []) {
      for (const facet of source.facets) declared.add(facet);
    }
    return [...declared];
  });

  /**
   * Apply the operator's filters.
   *
   * "New" means recently appeared, never "untriaged": on a first run every identity is untriaged,
   * so filtering on the decision would match everything and look broken.
   */
  const visible = computed(() => {
    const selection = facetSelection.value;
    const now = Date.now();
    return identities.value.filter((identity) => {
      if (!showDemoted.value && !shownByDefault(identity)) return false;
      if (onlyNew.value && !isRecentArrival(identity, now)) return false;
      if (!withinDistanceRange(identity, distanceRangeById(distanceRange.value))) return false;
      return Object.entries(selection).every(([facet, values]) => {
        if (values.length === 0) return true;
        if (facet === ObservationFacets.Kind) return values.includes(identity.kind);
        if (facet === ObservationFacets.Vendor) return values.includes(vendorHex(identity.vendor));
        // An unknown facet cannot be filtered here; ignoring it keeps a newer backend usable.
        return true;
      });
    });
  });

  /** Grouped sections for the list view. */
  const grouped = computed(() => groupIdentities(visible.value));

  /**
   * How many identities the default filter is hiding.
   *
   * Reported rather than silently dropped: a device that vanishes without explanation looks like a
   * fault, and the operator needs to know the count to decide whether to reveal them.
   */
  const demotedCount = computed(() => {
    if (showDemoted.value) return 0;
    return identities.value.filter((identity) => !shownByDefault(identity)).length;
  });

  /**
   * Isolated identity, resolved by key rather than held as an object.
   *
   * A snapshot update must not replace the selection with a stale copy, and an identity that
   * ages out has to be detectable so the UI can say so instead of silently showing the list.
   */
  const selectedKey = ref<string | null>(null);
  const selected = computed(
    () => identities.value.find((identity) => identity.key === selectedKey.value) ?? null,
  );
  const selectionLost = computed(() => selectedKey.value !== null && selected.value === null);

  function selectIdentity(key: string): void {
    selectedKey.value = key;
    viewMode.value = ObservationViews.Focus;
  }

  function clearSelection(): void {
    selectedKey.value = null;
    if (viewMode.value === ObservationViews.Focus) viewMode.value = ObservationViews.Radar;
  }

  /**
   * Reflect a triage decision in the frozen copy.
   *
   * While paused the render source is the frozen list, so without this the operator's own
   * decision would not appear until they resumed. The backend remains the authority; this only
   * closes the gap for the frozen view.
   */
  function applyDecisionToFrozen(key: string, decision: ObservationIdentityView['decision']): void {
    if (!frozen.value) return;
    frozen.value = frozen.value.map((identity) =>
      identity.key === key ? { ...identity, decision } : identity,
    );
  }

  /** What the radar and list render. Isolation is applied once so every view agrees. */
  const scoped = computed(() => (selected.value ? [selected.value] : visible.value));

  /**
   * Views offered now. Focus is only meaningful with a selection; offering it without one would
   * render an empty panel.
   */
  const availableViews = computed<ObservationViewMode[]>(() =>
    selected.value
      ? [ObservationViews.Focus, ObservationViews.Radar, ObservationViews.List]
      : [ObservationViews.Radar, ObservationViews.List],
  );

  // While frozen, drop rows the live snapshot no longer contains. Otherwise resuming would show
  // identities that aged out minutes ago as though they were still present.
  watch(liveIdentities, (next) => {
    if (!frozen.value) return;
    const live = new Set(next.map((identity) => identity.key));
    frozen.value = frozen.value.filter((identity) => live.has(identity.key));
  });

  return {
    facetSelection,
    onlyNew,
    showDemoted,
    demotedCount,
    distanceRange,
    radarSpan,
    viewMode,
    liveIdentities,
    identities,
    paused,
    togglePause,
    facets,
    visible,
    grouped,
    selectedKey,
    selected,
    selectionLost,
    selectIdentity,
    clearSelection,
    applyDecisionToFrozen,
    scoped,
    availableViews,
  };
}
