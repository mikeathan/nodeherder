/**
 * Viewer state for the wireless pane: filtering, isolation, freezing and view selection.
 *
 * Extracted from the component so this logic is unit-testable without mounting a Vue component.
 * The component keeps only rendering and event wiring, matching the repository's separation of
 * pure view logic into `domain/` and shared behaviour into `composables/`.
 */
import { computed, getCurrentInstance, onUnmounted, ref, watch } from 'vue';
import type {
  ObservationIdentityView,
  ObservationSnapshot,
  ObservationViewMode,
} from '@/types/observations.type';
import {
  ObservationDecisions,
  ObservationFacets,
  ObservationViews,
} from '@/types/observations.type';
import {
  ANY_DISTANCE_RANGE,
  shownByDefault,
  distanceRangeById,
  isRecentArrival,
  matchesIdentityQuery,
  withinDistanceRange,
  vendorHex,
  type DistanceRange,
  type RadarSpan,
} from '@/domain/observations/presentation';
import {
  ANY_CLASS_GROUP,
  CLASS_GROUPS,
  classGroupOf,
  groupIdentitiesByClass,
  matchesClassGroup,
  orderIdentities,
  type ClassGroup,
  type ObservationClassFilter,
  type ObservationClassGroup,
} from '@/domain/observations/classes';

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
  /** Selected class preset; "any" applies no class restriction. */
  const classFilter = ref<ObservationClassFilter>(ANY_CLASS_GROUP);
  /** Free-text query over title, key and vendor. */
  const search = ref('');
  /** Radar span in metres, or null to scale automatically to the data. */
  const radarSpan = ref<RadarSpan | null>(null);
  /** Whether ring distance labels are drawn. On by default; the circles carry the scale. */
  const showRingLabels = ref(true);
  /**
   * Whether device names are drawn beside their blips.
   *
   * Separate from ring labels because they answer different questions — "how far is that circle"
   * versus "which device is that" — and because the name has to be dropped selectively when devices
   * crowd, while a ring label never does.
   */
  const showDeviceNames = ref(true);
  /**
   * The view the operator is reading. List is the primary surface: the names and metadata need
   * the page's width, and a 24rem rail truncated every title. Radar is a deliberate alternative
   * view rather than the default half of a split.
   */
  const viewMode = ref<ObservationViewMode>(ObservationViews.List);
  /** Last view that was not Focus, so leaving a detail returns where the operator came from. */
  const lastNonFocusView = ref<ObservationViewMode>(ObservationViews.List);

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
   * A coarse clock, so "only new" ages out without waiting for a snapshot.
   *
   * `Date.now()` inside a computed is not reactive: the computed only re-evaluates when one of its
   * dependencies changes, so with a steady stream of identical snapshots an identity that is no
   * longer new would stay in the filtered view indefinitely. A ticking ref makes elapsed time a
   * dependency. It is one second because the window is five minutes — anything finer is wasted
   * work, anything coarser lets a stale row linger visibly.
   */
  const now = ref(Date.now());
  const clock = setInterval(() => (now.value = Date.now()), 1000);
  /**
   * Stop the coarse clock.
   *
   * A component clears it on unmount. Tests call it directly, because outside a component instance
   * there is no unmount to hook: `onUnmounted` only warns, the interval survives, and Jest hangs at
   * the end of the run on the open handle. Exposing it is what makes the lifetime explicit instead
   * of relying on the caller happening to be a component.
   */
  function stop(): void {
    clearInterval(clock);
  }
  if (getCurrentInstance()) onUnmounted(stop);

  /**
   * Identities in display order.
   *
   * Ordering happens here, once, rather than in a component: every surface (cards, list, radar
   * scoping) then agrees, and ordering uses only stable fields so a snapshot update cannot move
   * a row.
   */
  const orderedIdentities = computed(() => orderIdentities(identities.value));

  /**
   * Whether the default hiding still applies.
   *
   * An explicit class selection overrides it for as long as it is active: selecting "Installed"
   * must show installed equipment, otherwise the preset would look broken. Returning to "Any"
   * restores the default.
   */
  const defaultHidingActive = computed(
    () => !showDemoted.value && classFilter.value === ANY_CLASS_GROUP,
  );

  /**
   * The filters shared by the visible set and the class-preset counts, so a chip's number is the
   * number of rows the operator would get. Class and default hiding are applied by the caller.
   */
  function matchesOtherFilters(
    identity: ObservationIdentityView,
    range: DistanceRange,
    selection: Readonly<Record<string, readonly string[]>>,
    query: string,
  ): boolean {
    if (onlyNew.value && identity.decision !== ObservationDecisions.Unknown) return false;
    if (!withinDistanceRange(identity, range)) return false;
    if (!matchesIdentityQuery(identity, query)) return false;
    return Object.entries(selection).every(([facet, values]) => {
      if (values.length === 0) return true;
      if (facet === ObservationFacets.Kind) return values.includes(identity.kind);
      if (facet === ObservationFacets.Vendor) return values.includes(vendorHex(identity.vendor));
      // An unknown facet cannot be filtered here; ignoring it keeps a newer backend usable.
      return true;
    });
  }

  /**
   * Apply the operator's filters.
   *
   * "Needs triage" means the decision is still unknown, so triaging a device removes it from the
   * filtered list. It previously meant "arrived in the last five minutes", which looked broken:
   * marking a device known left it on screen, because the filter never considered the decision.
   * The two ideas are kept apart now — `newCount` still counts recent arrivals, which is a
   * different question and a different word.
   */
  const visible = computed(() => {
    const selection = facetSelection.value;
    // Resolved once per evaluation, not once per device. Calling distanceRangeById inside the
    // predicate ran an array search for every identity on every filter pass.
    const range = distanceRangeById(distanceRange.value);
    const query = search.value;
    const activeClass = classFilter.value;
    const hideByDefault = defaultHidingActive.value;
    return orderedIdentities.value.filter((identity) => {
      if (hideByDefault && !shownByDefault(identity)) return false;
      if (!matchesClassGroup(identity, activeClass)) return false;
      return matchesOtherFilters(identity, range, selection, query);
    });
  });

  /**
   * Class presets to offer, with the count each would yield.
   *
   * Counts apply every filter except the class one, so a number is what the operator gets on
   * click. Every group is returned in fixed order, including zero counts: omitting a group or
   * letting its count widen the chip made the row shuffle while the operator was reading it. The
   * component renders a zero-count group as disabled.
   */
  const classOptions = computed<(ClassGroup & { count: number })[]>(() => {
    const range = distanceRangeById(distanceRange.value);
    const selection = facetSelection.value;
    const query = search.value;
    const counts = new Map<ObservationClassGroup, number>();
    for (const identity of orderedIdentities.value) {
      if (!matchesOtherFilters(identity, range, selection, query)) continue;
      const group = classGroupOf(identity).id;
      counts.set(group, (counts.get(group) ?? 0) + 1);
    }
    return CLASS_GROUPS.map((group) => ({ ...group, count: counts.get(group.id) ?? 0 }));
  });

  /** Grouped sections for the list view, in fixed class order. */
  const grouped = computed(() => groupIdentitiesByClass(visible.value));

  /**
   * How many visible identities arrived recently.
   *
   * Computed here rather than in the component so it shares the same ticking clock as the filter:
   * two independent `Date.now()` reads a second apart could disagree, leaving a header count that
   * contradicts the rows beneath it.
   */
  const newCount = computed(
    () => visible.value.filter((identity) => isRecentArrival(identity, now.value)).length,
  );

  /**
   * How many identities the default filter is hiding.
   *
   * Reported rather than silently dropped: a device that vanishes without explanation looks like a
   * fault, and the operator needs to know the count to decide whether to reveal them. When a class
   * preset is active the default hiding is overridden, so nothing is hidden by that rule and the
   * footer must not claim otherwise.
   */
  const demotedCount = computed(() => {
    if (!defaultHidingActive.value) return 0;
    return identities.value.filter((identity) => !shownByDefault(identity)).length;
  });

  /**
   * Whether the "show all kinds" control is meaningful.
   *
   * Separate from `demotedCount`, which is deliberately zero once revealed. While a class preset
   * is active the control would do nothing — the preset already overrides default hiding — so it
   * is not offered.
   */
  const hasDemoted = computed(
    () =>
      classFilter.value === ANY_CLASS_GROUP &&
      identities.value.some((identity) => !shownByDefault(identity)),
  );

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
    if (viewMode.value !== ObservationViews.Focus) lastNonFocusView.value = viewMode.value;
    selectedKey.value = key;
    viewMode.value = ObservationViews.Focus;
  }

  function clearSelection(): void {
    selectedKey.value = null;
    if (viewMode.value === ObservationViews.Focus) viewMode.value = lastNonFocusView.value;
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
      ? [ObservationViews.Focus, ObservationViews.List, ObservationViews.Radar]
      : [ObservationViews.List, ObservationViews.Radar],
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
    hasDemoted,
    distanceRange,
    classFilter,
    classOptions,
    search,
    radarSpan,
    showRingLabels,
    showDeviceNames,
    viewMode,
    liveIdentities,
    identities,
    now,
    paused,
    togglePause,
    facets,
    visible,
    grouped,
    newCount,
    selectedKey,
    selected,
    selectionLost,
    selectIdentity,
    clearSelection,
    applyDecisionToFrozen,
    scoped,
    availableViews,
    stop,
  };
}
