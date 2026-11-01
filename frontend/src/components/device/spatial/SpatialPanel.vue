<script setup lang="ts">
  /*
   * Spatial tab shell. Draft state, persisted settings, workspace edits and
   * session analytics live in dedicated composables; this component binds them
   * to the tab navigation and presentation components.
   */
  import { computed, ref, watch } from 'vue';
  import type { Device } from '@/types/device';
  import { useSpatialView } from '@/composables/spatial/useSpatialView';
  import { useWorkspaceEditor } from '@/composables/spatial/useWorkspaceEditor';
  import { useSpatialSession } from '@/composables/spatial/useSpatialSession';
  import { formatSpatialDistance, radarGridSpacing, spatialTargetRows } from '@/domain/spatial/radar';
  import {
    DEFAULT_SPATIAL_VIEW,
    SPATIAL_CLEAR_PROMPT,
    SPATIAL_VIEWS,
    SpatialViewIds,
    connectionLabel,
    formatSessionStart,
    frameCountLabel,
    freshTargetLabel,
    isSpatialEditView,
    isSpatialLiveView,
    observationWindowLabel,
    sampleCountLabel,
    zoneOverlayLabel,
    zonePreviewRows,
    type SpatialViewId,
  } from '@/domain/spatial/view';
  import { OBSERVATION_WINDOW_MINUTES } from '@/domain/spatial/analytics';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import SpatialMappingEditor from './SpatialMappingEditor.vue';
  import SpatialDisplayOptions from './SpatialDisplayOptions.vue';
  import SpatialHelp from './SpatialHelp.vue';
  import SpatialContextHelp from './SpatialContextHelp.vue';
  import SpatialCanvas from './SpatialCanvas.vue';
  import SpatialZoneDevice from './SpatialZoneDevice.vue';
  import SpatialZoneEditor from './SpatialZoneEditor.vue';
  import SpatialRoomBuilder from './SpatialRoomBuilder.vue';
  import SpatialEntryPanel from './SpatialEntryPanel.vue';

  const props = defineProps<{ device: Device }>();
  const {
    mappings,
    mappingError,
    scale,
    orientation,
    online,
    configured,
    rowOptions,
    canSave,
    targets,
    targetStatuses,
    setMapping,
    selectedLabel,
    mappingWarning,
    saveDisplay,
    workspace,
    saveNotice,
    saveError,
    saving,
  } = useSpatialView(() => props.device);
  const {
    selectedId,
    selectedZone,
    selectedObject,
    undoStack,
    undo,
    addZone,
    addObject,
    updateZone,
    updateObject,
    updateSensor,
    mapEntry,
    mapZone,
    setOutline,
    removeSelected,
    moveSelected,
    beginGesture,
    endGesture,
    setVertex,
  } = useWorkspaceEditor(workspace);
  const {
    minutes,
    showTrails,
    cells,
    activity,
    trails,
    started,
    observedFrames,
    clearRequested,
    clear,
  } = useSpatialSession(
    () => `${props.device.id}:${JSON.stringify(mappings.value)}:${JSON.stringify(workspace.value.sensor)}`,
    () => targets.value,
    () => online.value
  );
  const mode = ref<SpatialViewId>(DEFAULT_SPATIAL_VIEW);
  const helpOpen = ref(false);
  const showGrid = ref(true);
  const showRoomContext = ref(true);
  const rows = computed(() => spatialTargetRows(targets.value));
  const spacing = computed(() => formatSpatialDistance(radarGridSpacing(scale.value)));
  const previews = computed(() => zonePreviewRows(props.device, workspace.value.zones, targets.value, online.value));
  const inactiveTargets = computed(() => targetStatuses.value.filter((status) => status.state !== 'active'));
  watch(mode, () => { selectedId.value = ''; });
</script>

<template>
  <UiCard title="Spatial view" icon="signal">
    <nav class="nh-spatial-nav" aria-label="Spatial tools">
      <UiButton
        v-for="view in SPATIAL_VIEWS"
        :key="view.id"
        :pressed="mode === view.id"
        :variant="mode === view.id ? 'primary' : 'default'"
        @click="mode = view.id">
        {{ view.label }}
      </UiButton>
    </nav>
    <div class="nh-spatial-status" role="status">
      <strong>{{ connectionLabel(online) }}</strong>
      <span>{{ freshTargetLabel(targets) }}</span>
      <span>{{ zoneOverlayLabel(workspace.zones.length) }}</span>
    </div>
    <p v-if="!configured" role="status">
      No target mapping is saved. Open <strong>Configure target mapping</strong>, choose the matching X, Y and Active
      entities, then select <strong>Save spatial settings</strong>.
    </p>
    <template v-if="mode === SpatialViewIds.Live">
      <details class="nh-spatial-help" @toggle="helpOpen = ($event.target as HTMLDetailsElement).open">
        <summary>Help with the spatial view</summary>
        <SpatialHelp v-if="helpOpen" :device-id="device.id" />
      </details>
      <div v-if="!workspace.zones.length" class="nh-first-zone">
        <div>
          <strong>Creating your first zone</strong>
          <p>Check the live targets, then mark a desk, sofa or doorway on the canvas.</p>
        </div>
        <UiButton @click="mode = SpatialViewIds.Zones">Create your first zone</UiButton>
      </div>
    </template>
    <details>
      <summary>Configure target mapping</summary>
      <SpatialMappingEditor
        :mappings="mappings"
        :row-options="rowOptions"
        :warning="mappingWarning"
        :unavailable-label="selectedLabel"
        @update="setMapping" />
    </details>
    <SpatialDisplayOptions
      :scale="scale"
      :orientation="orientation"
      @update:scale="scale = $event"
      @update:orientation="orientation = $event" />
    <div class="nh-spatial-toggles">
      <label><input v-model="showGrid" type="checkbox" /> Show grid</label>
      <label v-if="mode === SpatialViewIds.Heatmap">
        <input v-model="showRoomContext" type="checkbox" /> Room context
      </label>
      <label v-else><input v-model="showTrails" type="checkbox" /> Show recent trails</label>
      <UiButton v-if="isSpatialEditView(mode)" :disabled="!undoStack.length" @click="undo">Undo last edit</UiButton>
    </div>
    <div class="nh-spatial-workspace" :class="{ 'has-editor': isSpatialEditView(mode) }">
      <div class="nh-spatial-scene">
        <SpatialCanvas
          :workspace="workspace"
          :targets="targets"
          :scale="scale"
          :orientation="orientation"
          :online="online"
          :mode="mode"
          :selected-id="selectedId"
          :trails="trails"
          :cells="mode === SpatialViewIds.Heatmap ? cells : []"
          :show-grid="showGrid"
          :show-room-context="showRoomContext"
          @select="selectedId = $event"
          @gesture-start="beginGesture"
          @gesture-end="endGesture"
          @move="moveSelected"
          @vertex="setVertex" />
        <p class="nh-muted">Grid spacing {{ spacing }}. Sensor position and direction come from Room Builder.</p>
        <p v-if="!online" class="nh-muted">Device or hub connection is offline.</p>
        <p v-else-if="!targets.length" class="nh-muted">
          No current targets. Check the mapping and tracking settings, or wait for fresh coordinate updates.
        </p>
        <ul v-else>
          <li v-for="row in rows" :key="row">{{ row }}</li>
        </ul>
        <ul v-if="inactiveTargets.length" class="nh-spatial-target-status" aria-label="Inactive target mapping status">
          <li v-for="status in inactiveTargets" :key="status.id" :class="`is-${status.state}`">
            <strong>Target {{ status.id }}</strong> — {{ status.label }}
          </li>
        </ul>
        <p class="nh-muted">
          X and Y arrive separately. The canvas uses the latest fresh readings, not a simultaneous frame.
        </p>
      </div>
      <SpatialZoneEditor
        v-if="mode === SpatialViewIds.Zones"
        :zones="workspace.zones"
        :selected="selectedZone"
        @select="selectedId = $event"
        @add="addZone"
        @update="updateZone"
        @remove="removeSelected" />
      <SpatialRoomBuilder
        v-if="mode === SpatialViewIds.Room"
        :workspace="workspace"
        :selected="selectedObject"
        @add="addObject"
        @update="updateObject"
        @sensor="updateSensor"
        @select="selectedId = $event"
        @remove="removeSelected"
        @outline="setOutline" />
    </div>
    <SpatialZoneDevice
      v-if="mode === SpatialViewIds.Zones && selectedZone"
      :device="device"
      :zone="selectedZone"
      :online="online"
      @map="mapZone"
      @occupancy="updateZone({ occupancyEntity: $event })" />
    <section v-if="isSpatialLiveView(mode)" aria-label="Zone occupancy previews">
      <h3>Zone previews</h3>
      <p class="nh-muted">
        Calculated from fresh target coordinates. These are overlays, not firmware occupancy reports.
      </p>
      <ul>
        <li v-for="preview in previews" :key="preview.zone.id">
          <strong>{{ preview.zone.name }}</strong> · {{ preview.zone.kind }} · {{ preview.preview }}
          <span v-if="preview.reported !== null">
            · device reported: {{ preview.reported }}{{ preview.stale ? ' (offline; may be stale)' : '' }}</span>
        </li>
      </ul>
    </section>
    <SpatialEntryPanel
      v-if="mode === SpatialViewIds.Entry"
      :device="device"
      :workspace="workspace"
      :online="online"
      @map="mapEntry" />
    <section v-if="mode === SpatialViewIds.Heatmap" aria-label="Heatmap and analytics">
      <h3>Heatmap & analytics</h3>
      <SpatialContextHelp topic="heatmap" />
      <label>
        Observation window
        <select v-model.number="minutes" class="nh-input">
          <option v-for="option in OBSERVATION_WINDOW_MINUTES" :key="option" :value="option">
            {{ observationWindowLabel(option) }}
          </option>
        </select>
      </label>
      <p>
        Session collection started {{ formatSessionStart(started) }} (browser local time).
        {{ frameCountLabel(observedFrames) }}.
      </p>
      <p v-if="!cells.length" role="status">
        No observed positions in this window. Keep this Spatial tab open while fresh target updates arrive.
      </p>
      <p class="nh-heat-legend">
        Low activity <span aria-hidden="true" /> High activity · relative target-second samples
      </p>
      <h4>Activity by hour (local time)</h4>
      <ul>
        <li v-for="hour in activity" :key="hour.label">
          {{ hour.label }}: {{ sampleCountLabel(hour.count) }}
        </li>
      </ul>
      <details v-if="cells.length">
        <summary>Heatmap cell counts (sensor coordinates)</summary>
        <ul>
          <li v-for="cell in cells" :key="`${cell.x}:${cell.y}`">
            X {{ formatSpatialDistance(cell.x) }}, Y {{ formatSpatialDistance(cell.y) }}: {{ sampleCountLabel(cell.count) }}
          </li>
        </ul>
      </details>
      <UiButton @click="clearRequested = true">Clear session observations</UiButton>
      <div v-if="clearRequested" role="group" aria-label="Confirm clearing session observations">
        <p>{{ SPATIAL_CLEAR_PROMPT }}</p>
        <UiButton variant="danger" @click="clear">Confirm clear observations</UiButton>
        <UiButton @click="clearRequested = false">Cancel</UiButton>
      </div>
    </section>
    <p v-if="mappingError" role="alert">{{ mappingError }}</p>
    <p v-if="saveError" role="alert">{{ saveError }}</p>
    <p v-if="saveNotice" role="status">{{ saveNotice }}</p>
    <div class="nh-spatial-save">
      <UiButton :disabled="!canSave" :loading="saving" @click="saveDisplay">Save spatial settings</UiButton>
      <span class="nh-muted">Saves mappings and room/zone overlays to NodeHerder.</span>
    </div>
  </UiCard>
</template>
<style scoped>
  .nh-spatial-nav,
  .nh-spatial-status,
  .nh-spatial-toggles,
  .nh-spatial-save {
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem;
    align-items: center;
    margin-block: 0.75rem;
  }
  .nh-spatial-status {
    justify-content: space-between;
    padding: 0.75rem;
    background: var(--nh-surface-2);
    border-radius: var(--nh-radius-m);
  }
  .nh-spatial-workspace {
    display: grid;
    gap: 1rem;
    min-width: 0;
  }
  .has-editor {
    grid-template-columns: minmax(0, 1.7fr) minmax(16rem, 1fr);
  }
  .nh-spatial-scene {
    min-width: 0;
  }
  .nh-spatial-help {
    margin-block: 0.75rem;
  }
  .nh-spatial-target-status {
    margin: 0.5rem 0 0;
    padding-left: 1.25rem;
    color: var(--nh-text-muted);
    font-size: 0.85rem;
  }
  .nh-first-zone {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
  }
  .nh-heat-legend {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
  }
  .nh-heat-legend span {
    width: 8rem;
    height: 0.7rem;
    background: linear-gradient(to right, blue, cyan, green, yellow, red);
  }
  summary {
    cursor: pointer;
    padding-block: 0.5rem;
  }
  select {
    max-width: 100%;
  }
  ul {
    padding-left: 1.25rem;
  }
  @media (max-width: 900px) {
    .has-editor {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
