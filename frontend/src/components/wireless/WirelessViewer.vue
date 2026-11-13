<script setup lang="ts">
  import { computed, onMounted, onUnmounted, ref } from 'vue';
  import { store } from '@/store';
  import type { ObservationDecision, ObservationViewMode } from '@/types/observations.type';
  import { ObservationViews } from '@/types/observations.type';
  import {
    RADAR_SPANS_M,
    isRecentArrival,
    snapshotSummary,
    sourceIsSilent,
    sourceStatus,
  } from '@/domain/observations/presentation';
  import {
    OBSERVATION_VIEW_LABELS,
    useObservationViewer,
  } from '@/composables/useObservationViewer';
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

  /** Keys with a save in flight, so the row reports progress rather than looking inert. */
  const pending = ref<Set<string>>(new Set());

  const {
    facetSelection,
    onlyNew,
    showDemoted,
    demotedCount,
    distanceRange,
    radarSpan,
    viewMode,
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
  } = useObservationViewer(() => snapshot.value);

  /** Label for a view tab. */
  function viewLabel(mode: ObservationViewMode): string {
    return OBSERVATION_VIEW_LABELS[mode];
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
        :distance="distanceRange"
        :show-demoted="showDemoted"
        @update:selected="facetSelection = $event"
        @update:only-new="onlyNew = $event"
        @update:distance="distanceRange = $event"
        @update:show-demoted="showDemoted = $event"
      />

      <!--
        Hidden-by-class notice. The count is always present when anything is hidden, because a
        device that disappears without explanation reads as a fault rather than as a filter.
      -->
      <p v-if="demotedCount > 0" class="wireless-viewer__demoted" role="status">
        {{ demotedCount }} {{ demotedCount === 1 ? 'device' : 'devices' }} hidden — installed
        equipment and devices that look like neighbours.
        <button type="button" class="wireless-viewer__link" @click="showDemoted = true">
          Show them
        </button>
      </p>

      <ObservationFocus
        v-if="viewMode === ObservationViews.Focus && selected"
        :identity="selected"
        :pending="pending.has(selected.key)"
        @decide="decide"
        @clear="clearSelection"
      />

      <div v-else-if="viewMode === ObservationViews.Radar" class="wireless-viewer__layout">
        <div class="wireless-viewer__radar-pane">
          <ObservationRadar
            :identities="scoped"
            :selected-key="selectedKey"
            :span="radarSpan"
            @select="selectIdentity"
          />
          <!--
            Scale control. Auto follows the data so a nearby-only household fills the canvas;
            a fixed span keeps the scale still, which is what makes small movements visible.
          -->
          <div class="wireless-viewer__scale" role="group" aria-label="Radar scale">
            <span class="wireless-viewer__scale-label">Scale</span>
            <button
              type="button"
              class="wireless-viewer__scale-option"
              :class="{ 'is-active': radarSpan === null }"
              :aria-pressed="radarSpan === null"
              @click="radarSpan = null"
            >
              Auto
            </button>
            <button
              v-for="span in RADAR_SPANS_M"
              :key="span"
              type="button"
              class="wireless-viewer__scale-option"
              :class="{ 'is-active': radarSpan === span }"
              :aria-pressed="radarSpan === span"
              @click="radarSpan = span"
            >
              {{ span }} m
            </button>
          </div>
        </div>
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

  .wireless-viewer__demoted {
  margin: 0 0 0.75rem;
  font-size: 0.8rem;
  color: var(--text-muted, #495057);
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

  /*
   * The radar is the primary surface: it takes the remaining width and the table gets a bounded
   * column. The previous split capped the radar at 24rem, which made the global view small even
   * on a wide screen.
   */
  .wireless-viewer__layout {
    display: grid;
    gap: 1.5rem;
    grid-template-columns: minmax(0, 1fr) minmax(17rem, 24rem);
    align-items: start;
  }

  .wireless-viewer__radar-pane {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-width: 0;
  }

  .wireless-viewer__scale {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    justify-content: center;
    flex-wrap: wrap;
  }

  .wireless-viewer__scale-label {
    color: var(--text-muted, #8b9bab);
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    margin-right: 0.25rem;
  }

  .wireless-viewer__scale-option {
    background: transparent;
    border: 1px solid var(--border-subtle, #2a3441);
    color: inherit;
    font: inherit;
    font-size: 0.8125rem;
    padding: 0.2rem 0.6rem;
    border-radius: 999px;
    cursor: pointer;
  }

  .wireless-viewer__scale-option.is-active {
    background: var(--accent-primary, #4a9eff);
    border-color: var(--accent-primary, #4a9eff);
    color: #04121f;
  }

  .wireless-viewer__scale-option:focus-visible {
    outline: 2px solid var(--accent-primary, #4a9eff);
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
    .wireless-viewer__layout {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 899px) {
    .wireless-viewer {
      padding: 0.75rem;
      gap: 1rem;
    }

    .wireless-viewer__layout {
      grid-template-columns: 1fr;
      gap: 1rem;
    }

    .wireless-viewer__views {
      margin-left: 0;
    }

    /* On a phone the radar is the whole point of the view, so it gets the full column. */
    .wireless-viewer__toolbar {
      gap: 0.5rem;
    }
  }
</style>
