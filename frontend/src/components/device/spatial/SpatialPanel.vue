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
  import { useDerivedPresence } from '@/composables/spatial/useDerivedPresence';
  import { formatSpatialDistance, radarGridSpacing, radarViewBox, spatialTrackRows } from '@/domain/spatial/radar';
  import { SPATIAL_ILLUSTRATIVE_FOV } from '@/domain/spatial';
  import { SPATIAL_SMOOTHING_OPTIONS } from '@/domain/spatial/smoothing';
  import {
    DEFAULT_SPATIAL_VIEW,
    SpatialViewIds,
    connectionLabel,
    isSpatialEditView,
    isSpatialLiveView,
    sourceLabel,
    zoneOverlayLabel,
    zonePreviewRows,
    type SpatialViewId,
  } from '@/domain/spatial/view';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import SpatialMappingEditor from './SpatialMappingEditor.vue';
  import SpatialDisplayOptions from './SpatialDisplayOptions.vue';
  import SpatialHelp from './SpatialHelp.vue';
  import SpatialCanvas from './SpatialCanvas.vue';
  import SpatialViewer from './SpatialViewer.vue';
  import SpatialZoneDevice from './SpatialZoneDevice.vue';
  import SpatialZoneEditor from './SpatialZoneEditor.vue';
  import SpatialRoomBuilder from './SpatialRoomBuilder.vue';
  import SpatialEntryPanel from './SpatialEntryPanel.vue';
  import SpatialFirmwarePanel from './SpatialFirmwarePanel.vue';
  import SpatialCalibrationPanel from './SpatialCalibrationPanel.vue';

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
    tracks,
    occupancy,
    summary,
    source,
    hasData,
    jitterMm,
    duplicates,
    recentTargets,
    showRaw,
    smoothing,
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
    setSensor,
    mapEntry,
    mapZone,
    mapFirmware,
    setOutline,
    removeSelected,
    moveSelected,
    beginGesture,
    endGesture,
    setVertex,
  } = useWorkspaceEditor(workspace);
  const mode = ref<SpatialViewId>(DEFAULT_SPATIAL_VIEW);
  const helpOpen = ref(false);
  const showGrid = ref(true);
  const showFov = ref(false);
  const zoom = ref(1);
  const showRoomContext = ref(true);
  const displayFov = computed(() =>
    workspace.value.sensor.fov > 0 ? workspace.value.sensor.fov : SPATIAL_ILLUSTRATIVE_FOV
  );
  const rows = computed(() => spatialTrackRows(tracks.value));
  const derived = useDerivedPresence(() => props.device);
  // The hub is authoritative for zone occupancy once it derives it; the client map remains
  // the fallback so the canvas still shows something with an older hub.
  const canvasOccupancy = computed(() => {
    const signals = derived.zones.value;
    if (!derived.online.value || signals.size === 0) return occupancy.value;
    const merged = new Map<string, { occupied: boolean }>();
    for (const zone of workspace.value.zones) {
      const signal = signals.get(zone.id);
      const fallback = occupancy.value.get(zone.id)?.occupied ?? false;
      merged.set(zone.id, { occupied: signal?.occupied ?? fallback });
    }
    return merged;
  });
  const primaryMode = computed<'live' | 'draw' | 'setup'>(() => {
    const drawViews: SpatialViewId[] = [SpatialViewIds.Zones, SpatialViewIds.Room, SpatialViewIds.Entry];
    const setupViews: SpatialViewId[] = [SpatialViewIds.Firmware, SpatialViewIds.Calibrate];
    if (drawViews.includes(mode.value)) return 'draw';
    if (setupViews.includes(mode.value)) return 'setup';
    return 'live';
  });
  const PRIMARY_MODES = [
    { id: 'live' as const, label: 'Live' },
    { id: 'draw' as const, label: 'Draw' },
    { id: 'setup' as const, label: 'Setup' },
  ];
  const DRAW_VIEWS = [
    { id: SpatialViewIds.Zones, label: 'Zones' },
    { id: SpatialViewIds.Room, label: 'Room' },
    { id: SpatialViewIds.Entry, label: 'Entry / exit' },
  ];
  const SETUP_VIEWS = [
    { id: SpatialViewIds.Firmware, label: 'Firmware' },
    { id: SpatialViewIds.Calibrate, label: 'Calibration' },
  ];
  function selectPrimary(id: 'live' | 'draw' | 'setup') {
    mode.value = id === 'live' ? SpatialViewIds.Live : id === 'draw' ? SpatialViewIds.Zones : SpatialViewIds.Firmware;
  }
  // Setup panels are forms, not a scene; live and drawing keep the canvas.
  const showCanvas = computed(() => primaryMode.value !== 'setup');
  const serverPresence = computed(() => {
    if (!derived.derived.value) return 'Server presence: not derived yet.';
    if (!derived.online.value) return 'Server presence: derived device offline.';
    const values = derived.values.value;
    const count = values.targetCount ?? 0;
    const empty = values.emptySeconds ?? 0;
    const last = values.lastZone ?? 'none yet';
    return `Server presence: ${values.occupied ? 'occupied' : 'clear'} · ${count} ${count === 1 ? 'target' : 'targets'} · last zone ${last} · empty ${empty}s`;
  });
  const spacing = computed(() => formatSpatialDistance(radarGridSpacing(scale.value)));
  const previews = computed(() =>
    zonePreviewRows(props.device, workspace.value.zones, occupancy.value, online.value, hasData.value, derived.zones.value, derived.online.value)
  );
  const inactiveTargets = computed(() => targetStatuses.value.filter((status) => status.state !== 'active'));
  watch(mode, () => { selectedId.value = ''; });
</script>
<template>
  <UiCard title="Presence" icon="signal">
    <div class="nh-presence-head">
      <div class="nh-spatial-status" role="status">
        <strong>{{ connectionLabel(online) }}</strong>
        <span>{{ summary.label }}</span>
        <span>{{ zoneOverlayLabel(workspace.zones.length) }}</span>
      </div>
      <UiButton :disabled="!canSave" :loading="saving" @click="saveDisplay">Save spatial settings</UiButton>
    </div>
    <p v-if="mappingError" role="alert">{{ mappingError }}</p>
    <p v-if="saveError" role="alert">{{ saveError }}</p>
    <p v-if="saveNotice" role="status">{{ saveNotice }}</p>
    <p v-if="!configured" role="status">
      No target mapping is saved. Open <strong>Target mapping</strong> at the end of this tab, choose the matching X,
      Y and Active entities, then select <strong>Save spatial settings</strong>.
    </p>

    <nav class="nh-spatial-nav" aria-label="Presence mode">
      <UiButton v-for="primary in PRIMARY_MODES" :key="primary.id" :pressed="primaryMode === primary.id"
        :variant="primaryMode === primary.id ? 'primary' : 'default'" @click="selectPrimary(primary.id)">
        {{ primary.label }}
      </UiButton>
    </nav>
    <nav v-if="primaryMode === 'draw'" class="nh-spatial-subnav" aria-label="Drawing view">
      <UiButton v-for="view in DRAW_VIEWS" :key="view.id" size="sm" :pressed="mode === view.id"
        :variant="mode === view.id ? 'primary' : 'ghost'" @click="mode = view.id">{{ view.label }}</UiButton>
    </nav>
    <nav v-else-if="primaryMode === 'setup'" class="nh-spatial-subnav" aria-label="Setup view">
      <UiButton v-for="view in SETUP_VIEWS" :key="view.id" size="sm" :pressed="mode === view.id"
        :variant="mode === view.id ? 'primary' : 'ghost'" @click="mode = view.id">{{ view.label }}</UiButton>
    </nav>

    <div v-if="isSpatialEditView(mode)" class="nh-edit-actions">
      <UiButton size="sm" variant="ghost" :disabled="!undoStack.length" @click="undo">Undo last edit</UiButton>
    </div>

    <div class="nh-spatial-workspace" :class="{ 'has-editor': isSpatialEditView(mode) }">
      <div v-if="showCanvas" class="nh-spatial-scene">
        <SpatialViewer
          v-model:zoom="zoom"
          v-model:show-grid="showGrid"
          v-model:show-fov="showFov"
          :base-view-box="radarViewBox(scale)"
          :fov-angle="displayFov"
          :edit-mode="isSpatialEditView(mode)"
          :reset-key="mode">
          <template #default="{ viewBox, panMode }">
            <SpatialCanvas
              :workspace="workspace"
              :tracks="tracks"
              :occupancy="canvasOccupancy"
              :show-raw="showRaw"
              :scale="scale"
              :orientation="orientation"
              :online="online"
              :mode="mode"
              :selected-id="selectedId"
              :trails="[]"
              :cells="[]"
              :view-box="viewBox"
              :pan-mode="panMode"
              :show-grid="showGrid"
              :show-fov="showFov"
              :fov-angle="displayFov"
              :ui-scale="zoom"
              :show-room-context="showRoomContext"
              @select="selectedId = $event"
              @gesture-start="beginGesture"
              @gesture-end="endGesture"
              @move="moveSelected"
              @vertex="setVertex" />
          </template>
          <template #extra>
            <label class="nh-field">
              <span class="nh-label">Smoothing</span>
              <select v-model="smoothing" class="nh-input">
                <option v-for="option in SPATIAL_SMOOTHING_OPTIONS" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </label>
            <SpatialDisplayOptions :scale="scale" :orientation="orientation" @update:scale="scale = $event" @update:orientation="orientation = $event" />
            <label class="nh-presence-toggle"><input v-model="showRaw" type="checkbox" /> Raw</label>
          </template>
        </SpatialViewer>
        <div class="nh-presence-readout">
          <p class="nh-muted">{{ serverPresence }}</p>
          <p class="nh-muted">
            {{ sourceLabel(source) }} · Jitter {{ formatSpatialDistance(jitterMm) }} ·
            {{ duplicates }} duplicate {{ duplicates === 1 ? 'track' : 'tracks' }} suppressed · Grid
            {{ spacing }}
          </p>
          <p v-if="!online" class="nh-muted">Device or hub connection is offline.</p>
          <p v-else-if="!tracks.length" class="nh-muted">
            No current tracks. Check the mapping and tracking settings, or wait for fresh coordinate updates.
          </p>
          <ul v-else aria-label="Visible tracks">
            <li v-for="row in rows" :key="row">{{ row }}</li>
          </ul>
          <ul v-if="inactiveTargets.length" class="nh-spatial-target-status" aria-label="Inactive target mapping status">
            <li v-for="status in inactiveTargets" :key="status.id" :class="`is-${status.state}`">
              <strong>Target {{ status.id }}</strong> — {{ status.label }}
            </li>
          </ul>
          <p class="nh-muted">
            {{
              source === 'frames'
                ? 'Positions come from synchronized target frames. Sensor position and direction come from Room builder.'
                : 'Positions come from sampled coordinate updates; X and Y may not be simultaneous. Sensor position and direction come from Room builder.'
            }}
          </p>
        </div>
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

    <template v-if="mode === SpatialViewIds.Live">
      <div v-if="!workspace.zones.length" class="nh-first-zone">
        <div>
          <strong>Creating your first zone</strong>
          <p>Check the live tracks, then mark a desk, sofa or doorway on the canvas.</p>
        </div>
        <UiButton @click="mode = SpatialViewIds.Zones">Create your first zone</UiButton>
      </div>
      <details class="nh-spatial-help" @toggle="helpOpen = ($event.target as HTMLDetailsElement).open">
        <summary>Help with the presence view</summary>
        <SpatialHelp v-if="helpOpen" :device-id="device.id" />
      </details>
    </template>

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
        Calculated from filtered tracks with a short exit grace. These are overlays, not firmware occupancy
        reports.
      </p>
      <ul>
        <li v-for="preview in previews" :key="preview.zone.id">
          <strong>{{ preview.zone.name }}</strong> · {{ preview.zone.kind }} · {{ preview.preview }}
          <span v-if="preview.reported !== null">
            · device reported: {{ preview.reported }}{{ preview.stale ? ' (offline; may be stale)' : '' }}</span>
        </li>
      </ul>
    </section>

    <SpatialCalibrationPanel
      v-if="mode === SpatialViewIds.Calibrate"
      :device="device"
      :workspace="workspace"
      :targets="recentTargets"
      :slot-count="mappings.length"
      :online="online"
      @apply="setSensor" />

    <SpatialFirmwarePanel
      v-if="mode === SpatialViewIds.Firmware"
      :device="device"
      :workspace="workspace"
      :online="online"
      @map="mapFirmware" />

    <SpatialEntryPanel
      v-if="mode === SpatialViewIds.Entry"
      :device="device"
      :workspace="workspace"
      :online="online"
      @map="mapEntry" />

    <details class="nh-presence-setup">
      <summary>Target mapping</summary>
      <SpatialMappingEditor
        :mappings="mappings"
        :row-options="rowOptions"
        :warning="mappingWarning"
        :unavailable-label="selectedLabel"
        @update="setMapping" />
    </details>
  </UiCard>
</template>

<style scoped>
  .nh-presence-head {
    position: sticky;
    top: 0;
    z-index: 1;
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem;
    align-items: center;
    justify-content: space-between;
    padding-block: 0.35rem;
    background: var(--nh-surface);
  }
  .nh-presence-head .nh-spatial-status {
    flex: 1 1 16rem;
  }
  .nh-replay-banner {
    margin-block: 0.5rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--nh-accent);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
  }
  .nh-spatial-nav,
  .nh-spatial-status {
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
  /* Checkboxes sit on the same baseline as the select controls without a stacked label. */
  .nh-presence-toggle {
    display: inline-flex;
    gap: 0.35rem;
    align-items: center;
    height: var(--nh-ctl-h);
    font-size: 0.9rem;
    white-space: nowrap;
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
  .nh-presence-readout {
    margin-block-start: 0.5rem;
  }
  .nh-presence-readout ul {
    margin: 0.25rem 0 0;
    padding-left: 1.25rem;
  }
  .nh-spatial-help,
  .nh-presence-setup {
    margin-block: 0.75rem;
  }
  .nh-spatial-target-status {
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
    margin-block-start: 0.75rem;
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
