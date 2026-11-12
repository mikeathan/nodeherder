<script setup lang="ts">
  /*
   * Wireless observation viewer (spec 064).
   *
   * A hub-level live instrument, not configuration: it shows whatever any enrolled observing
   * source reports. Nothing here names a provider, so a new source kind appears without a
   * component change.
   *
   * Live snapshots arrive several times a second, so the view offers a freeze control: a table
   * that reorders while you read it is unusable for triage.
   */
  import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
  // The app installs the store under a Symbol injection key, so vuex's bare useStore() cannot
  // find it. Components use the singleton, as the rest of the codebase does.
  import { store } from '@/store';
  import type { ObservationDecision, ObservationIdentityView, ObservationViewMode } from '@/types/observations.type';
  import { ObservationFacets, ObservationViews } from '@/types/observations.type';
  import {
    groupIdentities,
    isRecentArrival,
    snapshotSummary,
    sourceIsSilent,
    sourceStatus,
    vendorHex,
  } from '@/domain/observations/presentation';
  import ObservationRadar from './ObservationRadar.vue';
  import ObservationTable from './ObservationTable.vue';
  import ObservationFilters from './ObservationFilters.vue';
  import ObservationGroupList from './ObservationGroupList.vue';
  import ObservationFocus from './ObservationFocus.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const snapshot = computed(() => store.getters['observations/snapshot']);
  const loaded = computed(() => store.getters['observations/hasData']);
  // Derived from the canonical socket status rather than a parallel flag of our own, so the
  // warning cannot disagree with the connection indicator in the header.
  const stale = computed(() => store.getters['ws/getConnectionStatus'] !== 'connected');
  /** Live identities exactly as the hub last reported them. */
  const liveIdentities = computed<ObservationIdentityView[]>(() => snapshot.value?.identities ?? []);

  /** Facet filter selections, distinct from the isolated identity. */
  const facetSelection = ref<Record<string, string[]>>({});
  /** Show only identities that appeared within the recent window. */
  const onlyNew = ref(false);
  const viewMode = ref<ObservationViewMode>(ObservationViews.Radar);
  /** Keys with a save in flight, so the row reports progress rather than looking inert. */
  const pending = ref<Set<string>>(new Set());

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
    for (const source of snapshot.value?.sources ?? []) {
      for (const facet of source.facets) declared.add(facet);
    }
    return [...declared];
  });

  /** Apply the operator's facet selection, the recent-only toggle and the frozen view. */
  const visible = computed(() => {
    const selection = facetSelection.value;
    const now = Date.now();
    return identities.value.filter((identity) => {
      // "New" means recently appeared, never "untriaged": on a first run every identity is
      // untriaged, so filtering on the decision matches everything and looks broken.
      if (onlyNew.value && !isRecentArrival(identity, now)) return false;
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
   * Isolated identity, resolved from the live set by key.
   *
   * Held as a key rather than an object so a snapshot update cannot replace the selected
   * identity with a stale copy, and so an identity that ages out is detectable.
   */
  const selectedKey = ref<string | null>(null);
  const selected = computed(() => identities.value.find((identity) => identity.key === selectedKey.value) ?? null);
  /** True when a selection was made but the identity has since left the snapshot. */
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
   * Identities the radar and list render. Isolation is applied here rather than inside each
   * child so every view agrees on what is shown.
   */
  const scoped = computed(() => (selected.value ? [selected.value] : visible.value));

  /**
   * Views offered right now. Focus is only meaningful with a selection, so it is not offered
   * while nothing is isolated; it would otherwise render an empty panel.
   */
  const availableViews = computed<ObservationViewMode[]>(() =>
    selected.value
      ? [ObservationViews.Focus, ObservationViews.Radar, ObservationViews.List]
      : [ObservationViews.Radar, ObservationViews.List],
  );

  const VIEW_LABELS: Readonly<Record<ObservationViewMode, string>> = {
    radar: 'Radar',
    list: 'List',
    focus: 'Focus',
  };

  /** Label for a view tab. */
  function viewLabel(mode: ObservationViewMode): string {
    return VIEW_LABELS[mode];
  }

  const summary = computed(() => snapshotSummary(snapshot.value));
  const newCount = computed(() => visible.value.filter((identity) => isRecentArrival(identity)).length);
  const truncated = computed(() => snapshot.value?.truncated === true);
  /** Sources that are connected but have delivered nothing, which usually means a config fault. */
  const sources = computed(() => snapshot.value?.sources ?? []);
  const silentSources = computed(() => sources.value.filter(sourceIsSilent));

  /**
   * Persist a triage decision. The row is marked pending until the request settles, and the
   * snapshot is not updated locally: the backend is the authority and broadcasts the result.
   * While paused the frozen copy is updated so the operator sees their own change land.
   */
  function decide(key: string, decision: ObservationDecision): void {
    pending.value = new Set(pending.value).add(key);
    if (frozen.value) {
      frozen.value = frozen.value.map((identity) =>
        identity.key === key ? { ...identity, decision } : identity,
      );
    }
    void store
      .dispatch('ws/emit', {
        event: 'saveObservationIdentity',
        message: { key, decision },
      })
      .finally(() => {
        const next = new Set(pending.value);
        next.delete(key);
        pending.value = next;
      });
  }

  // While frozen, drop rows the live snapshot no longer contains. Otherwise resuming would show
  // identities that aged out minutes ago as though they were still present.
  watch(liveIdentities, (next) => {
    if (!frozen.value) return;
    const live = new Set(next.map((identity) => identity.key));
    frozen.value = frozen.value.filter((identity) => live.has(identity.key));
  });

  // Escape leaves isolation, the expected way out of a selected detail view. Listening on the
  // window keeps it working wherever focus sits inside the view.
  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape' && selectedKey.value !== null) clearSelection();
  }
  onMounted(() => window.addEventListener('keydown', onKeydown));
  onUnmounted(() => window.removeEventListener('keydown', onKeydown));
</script>

<template>
  <section class="wireless-viewer" aria-labelledby="wireless-title">
    <header class="wireless-viewer__header">
      <div>
        <h1 id="wireless-title">Wireless</h1>
        <p class="wireless-viewer__summary" aria-live="polite">
          {{ summary }}
          <span v-if="newCount > 0">· {{ newCount }} new</span>
        </p>
      </div>
      <p v-if="stale" class="wireless-viewer__stale" role="status">
        Disconnected. Showing the last received data.
      </p>
    </header>

    <p v-if="!loaded" class="wireless-viewer__empty">
      Waiting for the first observation. Enrol a device with Bluetooth proxy enabled to begin.
    </p>

    <template v-else>
      <p v-if="truncated" class="wireless-viewer__notice" role="status">
        Showing the most recently heard devices only; the snapshot reached its size limit.
      </p>

      <!--
        Silence has several indistinguishable causes, so each source states its own delivery
        status. Without this an empty viewer could mean an unreachable host, a device without
        bluetooth_proxy, or genuinely nothing in range.
      -->
      <ul v-if="sources.length > 0" class="wireless-viewer__sources">
        <li v-for="source in sources" :key="source.id" :class="{ 'is-silent': sourceIsSilent(source) }">
          <strong>{{ source.name }}</strong>
          <span>{{ sourceStatus(source) }}</span>
        </li>
      </ul>

      <p v-if="silentSources.length > 0" class="wireless-viewer__notice" role="status">
        {{ silentSources.length === 1 ? 'One source is' : `${silentSources.length} sources are` }}
        connected but not sending advertisements. Confirm each device has
        <code>bluetooth_proxy</code> in its ESPHome configuration.
      </p>

      <div class="wireless-viewer__toolbar">
        <UiButton :variant="paused ? 'primary' : 'default'" :aria-pressed="paused" @click="togglePause">
          {{ paused ? 'Resume updates' : 'Pause updates' }}
        </UiButton>
        <span v-if="paused" class="wireless-viewer__paused" role="status">Frozen</span>

        <div class="wireless-viewer__views" role="group" aria-label="View">
          <button
            v-for="mode in availableViews"
            :key="mode"
            type="button"
            class="wireless-viewer__view"
            :class="{ 'is-active': viewMode === mode }"
            :aria-pressed="viewMode === mode"
            @click="viewMode = mode"
          >
            {{ viewLabel(mode) }}
          </button>
        </div>
      </div>

      <!--
        A selection that aged out must say so. Silently reverting to the full list would look
        like the click never registered.
      -->
      <p v-if="selectionLost" class="wireless-viewer__notice" role="status">
        That identity has not been heard recently, so it left the list.
        <button type="button" class="wireless-viewer__link" @click="clearSelection">Show all</button>
      </p>

      <ObservationFilters
        :identities="identities"
        :facets="facets"
        :selected="facetSelection"
        :only-new="onlyNew"
        @update:selected="facetSelection = $event"
        @update:only-new="onlyNew = $event"
      />

      <ObservationFocus
        v-if="viewMode === ObservationViews.Focus && selected"
        :identity="selected"
        :pending="pending.has(selected.key)"
        @decide="decide"
        @clear="clearSelection"
      />

      <div v-else-if="viewMode === ObservationViews.Radar" class="wireless-viewer__layout">
        <ObservationRadar :identities="scoped" />
        <div class="wireless-viewer__detail">
          <ObservationTable
            :identities="scoped"
            :pending="pending"
            :selected-key="selectedKey"
            @decide="decide"
            @select="selectIdentity"
          />
        </div>
      </div>

      <ObservationGroupList
        v-else-if="viewMode === ObservationViews.List"
        :groups="grouped"
        :pending="pending"
        :selected-key="selectedKey"
        @decide="decide"
        @select="selectIdentity"
      />
    </template>
  </section>
</template>

<style scoped>
  .wireless-viewer {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    padding: 1.5rem;
  }

  .wireless-viewer__header {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    flex-wrap: wrap;
    align-items: baseline;
  }

  .wireless-viewer__header h1 {
    font-size: 1.375rem;
    margin: 0;
  }

  .wireless-viewer__summary,
  .wireless-viewer__empty,
  .wireless-viewer__notice {
    color: var(--text-muted, #8b9bab);
    margin: 0.25rem 0 0;
    font-size: 0.875rem;
  }

  .wireless-viewer__stale {
    color: var(--warning-text, #e0a33e);
    margin: 0;
    font-size: 0.875rem;
  }

  .wireless-viewer__sources {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  .wireless-viewer__sources li {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    padding: 0.375rem 0.625rem;
    border: 1px solid var(--border-subtle, #2a3441);
    border-radius: 0.375rem;
    font-size: 0.8125rem;
  }

  /* A silent source is a likely misconfiguration, so it is marked by border and text, not
     colour alone. */
  .wireless-viewer__sources li.is-silent {
    border-style: dashed;
    border-color: var(--warning-text, #e0a33e);
  }

  .wireless-viewer__sources li span {
    color: var(--text-muted, #8b9bab);
  }

  .wireless-viewer__link {
    background: none;
    border: 0;
    color: var(--accent-primary, #4a9eff);
    font: inherit;
    padding: 0;
    margin-left: 0.375rem;
    cursor: pointer;
    text-decoration: underline;
  }

  .wireless-viewer__link:focus-visible {
    outline: 2px solid var(--accent-primary, #4a9eff);
    outline-offset: 2px;
  }

  .wireless-viewer__toolbar {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .wireless-viewer__paused {
    font-size: 0.8125rem;
    color: var(--warning-text, #e0a33e);
  }

  .wireless-viewer__views {
    display: inline-flex;
    border: 1px solid var(--border-subtle, #2a3441);
    border-radius: 999px;
    overflow: hidden;
    margin-left: auto;
  }

  .wireless-viewer__view {
    background: transparent;
    border: 0;
    color: inherit;
    font: inherit;
    padding: 0.25rem 0.875rem;
    cursor: pointer;
  }

  .wireless-viewer__view.is-active {
    background: var(--accent-primary, #4a9eff);
    color: #04121f;
  }

  .wireless-viewer__view:focus-visible {
    outline: 2px solid var(--accent-primary, #4a9eff);
    outline-offset: -2px;
  }

  .wireless-viewer__layout {
    display: grid;
    gap: 1.5rem;
    grid-template-columns: minmax(16rem, 24rem) minmax(0, 1fr);
    align-items: start;
  }

  .wireless-viewer__detail {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    min-width: 0;
  }

  /* Stack the radar above the table on narrow screens rather than shrinking both. */
  @media (max-width: 899px) {
    .wireless-viewer {
      padding: 1rem;
    }

    .wireless-viewer__layout {
      grid-template-columns: 1fr;
    }

    .wireless-viewer__views {
      margin-left: 0;
    }
  }
</style>
