<script setup lang="ts">
  import { computed, onMounted, onUnmounted, ref } from 'vue';
  import { store } from '@/store';
  import type {
    ObservationDecision,
    ObservationSourceDescriptor,
    ObservationViewMode,
  } from '@/types/observations.type';
  import { ObservationViews } from '@/types/observations.type';
  import {
    RADAR_SPANS_M,
    DEFAULT_RADAR_SPAN,
    ANY_DISTANCE_RANGE,
    stepRadarSpan,
    snapshotSummary,
    sourceIsSilent,
    sourceStatus,
  } from '@/domain/observations/presentation';
  import { ANY_CLASS_GROUP } from '@/domain/observations/classes';
  import {
    OBSERVATION_VIEW_LABELS,
    useObservationViewer,
  } from '@/composables/useObservationViewer';
  import ObservationRadar from './ObservationRadar.vue';
  import IntegrationPageHeader from '@/components/integrations/shared/IntegrationPageHeader.vue';
  import ObservationFilters from './ObservationFilters.vue';
  import ObservationGroupList from './ObservationGroupList.vue';
  import ObservationFocus from './ObservationFocus.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import { OBSERVATION_PROVIDER_KEY } from '@/domain/integrations/observation';
  import { RouteName } from '@/types/router';

  const snapshot = computed(() => store.getters['observations/snapshot']);
  const loaded = computed(() => store.getters['observations/hasData']);
  // Derived from the canonical socket status rather than a parallel flag of our own, so the
  // warning cannot disagree with the connection indicator in the header.
  const stale = computed(() => store.getters['ws/getConnectionStatus'] !== 'connected');

  /** Keys with a save in flight, so the row reports progress rather than looking inert. */
  const pending = ref<Set<string>>(new Set());

  const {
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
    identities,
    now,
    paused,
    togglePause,
    facets,
    grouped,
    visible,
    selectedKey,
    selected,
    selectionLost,
    selectIdentity,
    clearSelection,
    applyDecisionToFrozen,
    scoped,
    availableViews,
  } = useObservationViewer(() => snapshot.value);

  /** Label for a view tab. */
  function viewLabel(mode: ObservationViewMode): string {
    return OBSERVATION_VIEW_LABELS[mode];
  }

  /**
   * The radar's actual span, reported upward.
   *
   * This is the single source of truth for the readout and the zoom step. The viewer used to
   * compute its own auto-span from every identity, while the radar computed one from the
   * identifiable identities only — so at Auto the readout said 5 m while the rings were drawn at
   * 10 m, and the zoom controls stepped from a number the canvas was not using. The radar emits
   * what it settled on instead, and everything else reads that.
   */
  const reportedSpan = ref<number | null>(null);
  const currentSpan = computed(() => reportedSpan.value ?? DEFAULT_RADAR_SPAN);

  /**
   * Zoom by one stop. The span is in metres, so zooming in means a *smaller* number.
   *
   * Stepping starts from the span actually drawn, so pressing "+" cannot jump to a range wider
   * than the screen shows.
   */
  function zoom(direction: -1 | 1): void {
    radarSpan.value = stepRadarSpan(currentSpan.value, direction);
  }

  const canZoomIn = computed(() => currentSpan.value > RADAR_SPANS_M[0]);
  const canZoomOut = computed(() => currentSpan.value < RADAR_SPANS_M[RADAR_SPANS_M.length - 1]);

  /** Current range, so the zoom direction is never ambiguous. */
  const rangeLabel = computed(() => `${currentSpan.value} m`);

  /**
   * Whether classification has enough evidence yet.
   *
   * Undefined means an older backend that does not report progress; the banner is omitted rather
   * than guessed at, because claiming "not ready" when the backend simply does not say would be a
   * lie in the other direction.
   */
  const progress = computed(() => snapshot.value?.classification);
  const waitingForEvidence = computed(() => progress.value !== undefined && !progress.value.ready);

  const summary = computed(() => snapshotSummary(snapshot.value));

  /**
   * What an empty list means this time.
   *
   * "Nothing to show" is wrong when a filter or query removed everything, and the operator needs
   * to tell an empty household from a filter they forgot about.
   */
  const emptyMessage = computed(() => {
    const query = search.value.trim();
    if (query) return `No devices match “${query}”.`;
    const narrowed =
      classFilter.value !== ANY_CLASS_GROUP ||
      distanceRange.value !== ANY_DISTANCE_RANGE.id ||
      onlyNew.value ||
      Object.values(facetSelection.value).some((values) => values.length > 0);
    return narrowed
      ? 'No devices match the current filters.'
      : 'Nothing to show. Advertisements appear here as sources report them.';
  });
  /**
   * Whether the integration is currently reading the radio.
   *
   * Undefined means an older backend that does not report it, which is treated as observing: those
   * builds had no gate flag and anything on this page was being read. A viewer that silently shows
   * nothing because the integration was disabled looks identical to one with a broken sensor, and
   * the two have completely different fixes.
   */
  const observing = computed(() => snapshot.value?.observing !== false);
  /** The integration's own page, where observing is turned on and off. */
  const settingsRoute = computed(() => {
    const instance = store.getters['integrations/instanceForProviderKey'](OBSERVATION_PROVIDER_KEY);
    // No instance at all means it was never added or has been removed; the catalog is where that is
    // fixed, so the link goes there rather than to a page that cannot exist yet.
    return instance
      ? { name: RouteName.Integration, params: { id: instance.id }, hash: '#setup' }
      : { name: RouteName.Integrations };
  });
  const truncated = computed(() => snapshot.value?.truncated === true);
  /** Sources that are connected but have delivered nothing, which usually means a config fault. */
  const sources = computed<ObservationSourceDescriptor[]>(() => snapshot.value?.sources ?? []);
  const silentSources = computed(() => sources.value.filter(sourceIsSilent));
  /** Source id to display name, so a row can say which receiver heard it. */
  const sourceNames = computed<ReadonlyMap<string, string>>(
    () => new Map(sources.value.map((source): [string, string] => [source.id, source.name])),
  );

  /**
   * Persist a triage decision. The row is marked pending until the request settles, and the
   * snapshot is not updated locally: the backend is the authority and broadcasts the result.
   */
  function decide(key: string, decision: ObservationDecision): void {
    pending.value = new Set(pending.value).add(key);
    applyDecisionToFrozen(key, decision);
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

  // Escape leaves isolation, the expected way out of a selected detail view. Listening on the
  // window keeps it working wherever focus sits inside the view.
  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape' && selectedKey.value !== null) clearSelection();
  }
  onMounted(() => {
    window.addEventListener('keydown', onKeydown);
    // This page can be opened directly, so the instance list may never have been read. Without it
    // the settings link has nothing to point at and would fall back to the catalog even though the
    // integration exists. A failure is deliberately ignored: the radar still works without the link,
    // and the link's fallback is still a valid destination.
    if (!store.getters['integrations/instanceForProviderKey'](OBSERVATION_PROVIDER_KEY)) {
      void store.dispatch('integrations/readInstances').catch(() => {});
    }
  });
  onUnmounted(() => window.removeEventListener('keydown', onKeydown));
</script>

<template>
  <section class="wireless-viewer" aria-labelledby="wireless-title">
    <!--
      Integration-style header with a way back. This page is reached from an integration card or a
      Home shortcut, so it needs the same return affordance every other integration page has —
      arriving here with no way back left the operator stranded on a page that looks like a section
      of the application rather than a detail of an integration.
    -->
    <IntegrationPageHeader :title="'Wireless observation'" :subtitle="summary">
      <template #actions>
        <p v-if="stale" class="wireless-viewer__stale" role="status">
          Disconnected. Showing the last received data.
        </p>
        <!--
          The way in and out of this page. Without it, an operator who lands here from the board or a
          Home shortcut has no route to the controls that turn observation on and off, and would
          have to go back and hunt through Integrations.

          Labelled "Integration settings" rather than "Settings": the sidebar already has a Settings
          destination, so a bare "Settings" here reads as a second one and is ambiguous to a screen
          reader as well as to a person scanning for the app's settings.
        -->
        <UiButton size="sm" icon="settings" :to="settingsRoute">Integration settings</UiButton>
      </template>
    </IntegrationPageHeader>

    <!--
      Not observing is a state, not an absence. The radar renders exactly what a broken sensor
      renders — nothing — so it has to say which one this is and offer the fix inline.
    -->
    <p v-if="loaded && !observing" class="wireless-viewer__notice" role="status">
      Observation is turned off, so nothing is being recorded.
      <RouterLink :to="settingsRoute">Turn it on in the integration settings</RouterLink>.
    </p>

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
      <!--
        Source health is a diagnostic, not the main content. It sits behind a disclosure so it is
        available when something looks wrong without occupying the top of the page on every visit;
        a silent source still forces the notice below, so a fault cannot hide in here.
      -->
      <details v-if="sources.length > 0" class="wireless-viewer__sources-details">
        <summary>{{ sources.length }} {{ sources.length === 1 ? 'source' : 'sources' }}</summary>
      <ul class="wireless-viewer__sources">
        <li v-for="source in sources" :key="source.id" :class="{ 'is-silent': sourceIsSilent(source) }">
          <strong>{{ source.name }}</strong>
          <span>{{ sourceStatus(source) }}</span>
        </li>
      </ul>
      </details>

      <p v-if="silentSources.length > 0" class="wireless-viewer__notice" role="status">
        {{ silentSources.length === 1 ? 'One source is' : `${silentSources.length} sources are` }}
        connected but not sending advertisements. Confirm each device has
        <code>bluetooth_proxy</code> in its ESPHome configuration.
      </p>

      <!--
        Waiting state. A first day of observation and a broken classifier look identical — both
        show every row as learning — so this states which one it is. Without it the feature is
        indistinguishable from one that does not work.
      -->
      <p v-if="waitingForEvidence" class="wireless-viewer__waiting" role="status">
        <strong>Still gathering evidence.</strong>
        {{ progress?.note }}
        <span v-if="progress && progress.hoursObserved > 0" class="wireless-viewer__waiting-count">
          Best device so far: {{ progress.hoursObserved }} of 24 hours.
        </span>
      </p>

      <div class="wireless-viewer__toolbar">
        <UiButton :variant="paused ? 'primary' : 'default'" :aria-pressed="paused" @click="togglePause">
          {{ paused ? 'Resume updates' : 'Pause updates' }}
        </UiButton>
        <span v-if="paused" class="wireless-viewer__paused" role="status">Frozen</span>
        <span class="wireless-viewer__count">
          {{ visible.length }} of {{ identities.length }} shown
        </span>

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
        :distance="distanceRange"
        :class-filter="classFilter"
        :class-options="classOptions"
        :search="search"
        :show-demoted="showDemoted"
        :has-demoted="hasDemoted"
        @update:selected="facetSelection = $event"
        @update:only-new="onlyNew = $event"
        @update:distance="distanceRange = $event"
        @update:class-filter="classFilter = $event"
        @update:search="search = $event"
        @update:show-demoted="showDemoted = $event"
      />

      <!--
        Hidden-by-class notice. The count is always present when anything is hidden, because a
        device that disappears without explanation reads as a fault rather than as a filter.
      -->
      <!--
        One status line rather than two. "4 shown" and the hidden count were separate paragraphs,
        each costing a row of vertical space above the radar; together they read as one sentence
        about the current view and reclaim room for the canvas.
      -->
      <p v-if="demotedCount > 0" class="wireless-viewer__demoted" role="status">
        {{ demotedCount }} of {{ identities.length }} {{ identities.length === 1 ? 'device' : 'devices' }}
        hidden — installed equipment that never moves.
        <button type="button" class="wireless-viewer__link" @click="showDemoted = true">
          Show them
        </button>
      </p>

      <div
        v-if="viewMode !== ObservationViews.List"
        class="wireless-viewer__layout"
        :class="{ 'is-split': selected }"
      >
        <div class="wireless-viewer__radar-pane">
          <!--
            The scale bar sits above the canvas. It used to sit below it, which put zoom and the
            label toggle off the bottom of the screen once the radar was tall enough to be useful —
            the one place they are needed was the one place they could not be reached.
          -->
          <div class="wireless-viewer__scale" role="group" aria-label="Radar range">
            <button
              type="button"
              class="wireless-viewer__zoom"
              :disabled="!canZoomIn"
              aria-label="Zoom in"
              title="Zoom in"
              @click="zoom(-1)"
            >
              +
            </button>
            <span class="wireless-viewer__range" aria-live="polite">
              {{ rangeLabel }}
            </span>
            <button
              type="button"
              class="wireless-viewer__zoom"
              :disabled="!canZoomOut"
              aria-label="Zoom out"
              title="Zoom out"
              @click="zoom(1)"
            >
              −
            </button>
            <button
              type="button"
              class="wireless-viewer__scale-option"
              :class="{ 'is-active': radarSpan === null }"
              :aria-pressed="radarSpan === null"
              @click="radarSpan = null"
            >
              Auto
            </button>
            <!--
              Ring distances, not device names. The button used to read "Labels", which was taken
              to mean device labels — a different feature. "Rings" says exactly what it toggles.
            -->
            <button
              type="button"
              class="wireless-viewer__scale-option"
              :class="{ 'is-active': showRingLabels }"
              :aria-pressed="showRingLabels"
              @click="showRingLabels = !showRingLabels"
            >
              Rings
            </button>
            <!-- Device names are a separate control from ring distances: one labels the scale,
                 the other labels the devices. -->
            <button
              type="button"
              class="wireless-viewer__scale-option"
              :class="{ 'is-active': showDeviceNames }"
              :aria-pressed="showDeviceNames"
              @click="showDeviceNames = !showDeviceNames"
            >
              Names
            </button>
          </div>

          <ObservationRadar
            :identities="scoped"
            :selected-key="selectedKey"
            :span="radarSpan"
            :show-labels="showRingLabels"
            :show-names="showDeviceNames"
            @select="selectIdentity"
            @span-changed="reportedSpan = $event"
          />
        </div>
        <!--
          Selecting a device adds its detail beside the radar. The radar no longer shares its
          column with a row list: that rail was capped at 24rem, which truncated every name.
        -->
        <div v-if="selected" class="wireless-viewer__detail">
          <ObservationFocus
            :identity="selected"
            :pending="pending.has(selected.key)"
            @decide="decide"
            @clear="clearSelection"
          />
        </div>
      </div>

      <ObservationGroupList
        v-else
        :groups="grouped"
        :pending="pending"
        :selected-key="selectedKey"
        :now="now"
        :source-names="sourceNames"
        :empty-message="emptyMessage"
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
    /* Tighter than the previous 1.25rem: with the sources collapsed and the filters on one line,
       that gap left visible dead space between every band and pushed the content down. */
    gap: 0.75rem;
    padding: 1.25rem;
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
    color: var(--nh-text-muted);
    margin: 0.25rem 0 0;
    font-size: 0.875rem;
  }

  .wireless-viewer__stale {
    color: var(--nh-warn);
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
    border: 1px solid var(--nh-border);
    border-radius: 0.375rem;
    font-size: 0.8125rem;
  }

  /* A silent source is a likely misconfiguration, so it is marked by border and text, not
     colour alone. */
  .wireless-viewer__sources li.is-silent {
    border-style: dashed;
    border-color: var(--nh-warn);
  }

  .wireless-viewer__sources li span {
    color: var(--nh-text-muted);
  }

  .wireless-viewer__waiting {
    margin: 0;
    padding: 0.6rem 0.75rem;
    border-left: 3px solid var(--nh-accent);
    background: var(--nh-surface-2);
    border-radius: 0 0.25rem 0.25rem 0;
    font-size: 0.8rem;
    line-height: 1.45;
    color: var(--nh-text-muted);
  }

  .wireless-viewer__waiting-count {
    display: block;
    margin-top: 0.15rem;
    font-variant-numeric: tabular-nums;
    color: var(--nh-text-muted);
  }

  .wireless-viewer__demoted {
  margin: 0 0 0.75rem;
  font-size: 0.8rem;
  color: var(--nh-text-muted);
}

.wireless-viewer__link {
    background: none;
    border: 0;
    color: var(--nh-accent);
    font: inherit;
    padding: 0;
    margin-left: 0.375rem;
    cursor: pointer;
    text-decoration: underline;
  }

  .wireless-viewer__link:focus-visible {
    outline: 2px solid var(--nh-accent);
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
    color: var(--nh-warn);
  }

  .wireless-viewer__views {
    display: inline-flex;
    border: 1px solid var(--nh-border);
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
    background: var(--nh-accent);
    color: #04121f;
  }

  .wireless-viewer__view:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: -2px;
  }

  /*
   * One column by default: the list view owns its own full-width layout, and the radar stands
   * alone here. Only an isolated device adds a second column for its detail.
   */
  .wireless-viewer__layout {
    display: grid;
    gap: 1.5rem;
    grid-template-columns: minmax(0, 1fr);
    align-items: start;
  }

  .wireless-viewer__layout.is-split {
    grid-template-columns: minmax(0, 1fr) minmax(18rem, 24rem);
  }

  .wireless-viewer__radar-pane {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-width: 0;
    /* Bound the radar by the column, so it can never push the detail panel out of view. */
    max-width: min(100%, 68vh);
  }

  .wireless-viewer__scale {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    justify-content: center;
    flex-wrap: wrap;
  }

  .wireless-viewer__zoom {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border: 1px solid var(--nh-border);
    border-radius: 50%;
    background: var(--nh-surface);
    color: var(--nh-text-muted);
    font-size: 1rem;
    font-weight: 600;
    line-height: 1;
    cursor: pointer;
  }

  .wireless-viewer__zoom:hover:not(:disabled) {
    border-color: var(--nh-accent);
    color: var(--nh-accent);
  }

  .wireless-viewer__zoom:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  /* The range readout is the anchor between the two buttons, so "in" and "out" are unambiguous. */
  .wireless-viewer__range {
    min-width: 3.5rem;
    text-align: center;
    font-size: 0.8rem;
    font-variant-numeric: tabular-nums;
    color: var(--nh-text-muted);
  }

  .wireless-viewer__scale-option {
    background: transparent;
    border: 1px solid var(--nh-border);
    color: inherit;
    font: inherit;
    font-size: 0.8125rem;
    padding: 0.2rem 0.6rem;
    border-radius: 999px;
    cursor: pointer;
  }

  .wireless-viewer__scale-option.is-active {
    background: var(--nh-accent);
    border-color: var(--nh-accent);
    color: #04121f;
  }

  .wireless-viewer__scale-option:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: 2px;
  }

  .wireless-viewer__detail {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    min-width: 0;
  }

  /* Stack the radar above the table on narrow screens rather than shrinking both. */
  /* Below 1200px the table would squeeze the radar, so stack with the radar first. */
  @media (max-width: 1199px) {
    .wireless-viewer__layout,
    .wireless-viewer__layout.is-split {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 899px) {
    .wireless-viewer {
      padding: 0.75rem;
      gap: 1rem;
    }

    .wireless-viewer__layout,
    .wireless-viewer__layout.is-split {
      grid-template-columns: 1fr;
      gap: 1rem;
    }

    /* The view switch owns its row on a phone instead of sitting stranded beside the count. */
    .wireless-viewer__views {
      width: 100%;
      margin-left: 0;
    }

    .wireless-viewer__view {
      flex: 1;
      text-align: center;
    }

    /* On a phone the radar is the whole point of the view, so it gets the full column. */
    .wireless-viewer__toolbar {
      gap: 0.5rem;
    }
  }
</style>
