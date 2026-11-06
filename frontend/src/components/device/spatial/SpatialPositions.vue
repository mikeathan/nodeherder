<script setup lang="ts">
  import { computed, onBeforeUnmount, ref, watch } from 'vue';
  import { copyWorkspace } from '@/domain/spatial/workspace';
  import { SPATIAL_DEFAULT_ORIENTATION, SPATIAL_DEFAULT_SCALE_MM } from '@/domain/spatial';
  import { SpatialViewIds } from '@/domain/spatial/view';
  import type { SpatialDisplayTrack } from '@/domain/spatial/tracks';
  import type { HeatCell } from '@/domain/spatial/analytics';
  import type { SpatialHeatCell, SpatialPoint, SpatialSession } from '@/services/spatial-history.service';
  import { spatialHistoryApi } from '@/services/spatial-history.service';
  import { useDeviceSettings } from '@/composables/useDeviceSettings';
  import SpatialCanvas from './SpatialCanvas.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{ deviceId: string }>();
  const { settings } = useDeviceSettings(() => props.deviceId);
  const workspace = computed(() => copyWorkspace(settings.value.spatial?.workspace));
  const preset = ref<'1h' | '24h' | '7d'>('24h');
  const view = ref<'player' | 'heatmap'>('player');
  const sessions = ref<SpatialSession[]>([]);
  const selected = ref('');
  const points = ref<SpatialPoint[]>([]);
  const cells = ref<SpatialHeatCell[]>([]);
  const loading = ref(false);
  const error = ref('');
  const time = ref(0);
  const playing = ref(false);
  const speed = ref(1);
  let timer: ReturnType<typeof setInterval> | null = null;
  const spans = { '1h': 3_600_000, '24h': 86_400_000, '7d': 604_800_000 } as const;
  const to = ref(Date.now());
  const from = computed(() => to.value - spans[preset.value]);
  const currentSession = computed(() => sessions.value.find((session) => session.id === selected.value));
  const cellCount = computed(() => cells.value.length);
  const heat = computed<HeatCell[]>(() => {
    const max = Math.max(1, ...cells.value.map((cell) => cell.count));
    return cells.value.map((cell) => ({ ...cell, intensity: cell.count / max }));
  });
  const tracks = computed<SpatialDisplayTrack[]>(() => {
    const at = time.value || (currentSession.value?.to ?? 0);
    const latest = new Map<number, SpatialPoint>();
    for (const point of points.value) {
      if (point.at > at) break;
      latest.set(point.slot, point);
    }
    return [...latest.entries()].map(([slot, point]) => ({
      id: slot, label: `Slot ${slot}`, slot, x: point.x, y: point.y, rawX: point.x, rawY: point.y,
      vx: 0, vy: 0, at: point.at, colorIndex: (slot - 1) % 4, state: 'live' as const,
    }));
  });
  const timeLabel = computed(() => (time.value ? new Date(time.value).toLocaleTimeString() : '—'));

  function stop() {
    playing.value = false;
    if (timer) clearInterval(timer);
    timer = null;
  }
  function step() {
    const last = currentSession.value?.to ?? 0;
    time.value += 200 * speed.value;
    if (time.value >= last) { time.value = last; stop(); }
  }
  function play() {
    if (playing.value || !points.value.length) return;
    if (time.value >= (currentSession.value?.to ?? 0)) time.value = currentSession.value?.from ?? 0;
    playing.value = true;
    timer = setInterval(step, 200);
  }
  function seek(value: number) { stop(); time.value = value; }
  async function loadSessions() {
    loading.value = true; error.value = '';
    try {
      sessions.value = await spatialHistoryApi.sessions(props.deviceId, from.value, to.value);
      selected.value = sessions.value[0]?.id ?? '';
      time.value = sessions.value[0]?.from ?? 0;
    } catch (cause) { error.value = (cause as Error).message; sessions.value = []; selected.value = ''; }
    finally { loading.value = false; }
  }
  async function loadPoints() {
    stop();
    points.value = [];
    if (!selected.value) return;
    loading.value = true; error.value = '';
    try {
      points.value = await spatialHistoryApi.points(props.deviceId, selected.value, from.value, to.value);
      time.value = currentSession.value?.from ?? 0;
    } catch (cause) { error.value = (cause as Error).message; }
    finally { loading.value = false; }
  }
  async function loadHeatmap() {
    loading.value = true; error.value = '';
    try { cells.value = await spatialHistoryApi.heatmap(props.deviceId, from.value, to.value); }
    catch (cause) { error.value = (cause as Error).message; cells.value = []; }
    finally { loading.value = false; }
  }
  function choosePreset(value: '1h' | '24h' | '7d') { preset.value = value; to.value = Date.now(); }
  watch([preset, () => props.deviceId], () => { to.value = Date.now(); void loadSessions(); void loadHeatmap(); }, { immediate: true });
  watch(selected, () => { void loadPoints(); });
  onBeforeUnmount(stop);
</script>

<template>
  <section class="spatial-positions" aria-label="Recorded positions">
    <div class="nh-positions-toolbar">
      <div role="group" aria-label="Positions view">
        <UiButton size="sm" :variant="view === 'player' ? 'primary' : 'ghost'" :pressed="view === 'player'" @click="view = 'player'">Player</UiButton>
        <UiButton size="sm" :variant="view === 'heatmap' ? 'primary' : 'ghost'" :pressed="view === 'heatmap'" @click="view = 'heatmap'; loadHeatmap()">Heatmap</UiButton>
      </div>
      <div role="group" aria-label="History range">
        <UiButton v-for="option in (['1h','24h','7d'] as const)" :key="option" size="sm"
          :variant="preset === option ? 'primary' : 'ghost'" :pressed="preset === option" @click="choosePreset(option)">
          {{ option }}
        </UiButton>
      </div>
      <UiButton size="sm" variant="ghost" @click="loadSessions(); loadHeatmap()">Refresh</UiButton>
    </div>
    <p v-if="error" role="alert">{{ error }}</p>
    <p v-else-if="loading" role="status">Loading recorded positions…</p>
    <template v-else>
      <SpatialCanvas :workspace="workspace" :tracks="view === 'player' ? tracks : []" :occupancy="new Map()"
        :scale="SPATIAL_DEFAULT_SCALE_MM" :orientation="SPATIAL_DEFAULT_ORIENTATION" :online="true"
        :mode="view === 'heatmap' ? SpatialViewIds.Heatmap : SpatialViewIds.Live" selected-id="" :trails="[]"
        :cells="view === 'heatmap' ? heat : []" :show-grid="true" :show-room-context="true" />
      <template v-if="view === 'player'">
        <ul class="nh-session-list" aria-label="Recorded sessions">
          <li v-for="session in sessions" :key="session.id">
            <button type="button" :class="{ 'is-selected': session.id === selected }" @click="selected = session.id">
              {{ new Date(session.from).toLocaleString() }} · {{ Math.round((session.to - session.from) / 60000) }} min ·
              {{ session.points }} points · slots {{ session.slots.join(',') }}
              <span v-if="session.capped"> · capped</span>
            </button>
          </li>
        </ul>
        <p v-if="!sessions.length" class="nh-muted">No recorded positions in this range. Enable “Record history” on the device and let a target move.</p>
        <template v-else-if="currentSession">
          <div class="nh-player-controls">
            <UiButton size="sm" :disabled="!points.length" @click="playing ? stop() : play()">{{ playing ? 'Pause' : 'Play' }}</UiButton>
            <label>Speed
              <select class="nh-input" :value="speed" @change="speed = Number(($event.target as HTMLSelectElement).value)">
                <option :value="1">1×</option><option :value="2">2×</option><option :value="4">4×</option>
              </select>
            </label>
            <span role="status">{{ timeLabel }}</span>
            <input class="nh-input" type="range" :min="currentSession.from" :max="currentSession.to" :value="time"
              step="200" @input="seek(Number(($event.target as HTMLInputElement).value))" />
          </div>
        </template>
      </template>
      <p v-else-if="!cellCount" class="nh-muted">No recorded positions in this range.</p>
    </template>
  </section>
</template>

<style scoped>
  .nh-positions-toolbar { display: flex; flex-wrap: wrap; gap: 0.5rem; align-items: center; justify-content: space-between; margin-block: 0.5rem; }
  .nh-session-list { list-style: none; padding: 0; display: grid; gap: 0.35rem; max-height: 14rem; overflow: auto; }
  .nh-session-list button { text-align: start; width: 100%; padding: 0.4rem 0.6rem; border: 1px solid var(--nh-border); border-radius: var(--nh-radius); background: transparent; color: inherit; }
  .nh-session-list button.is-selected { border-color: var(--nh-accent); background: var(--nh-surface-2); }
  .nh-player-controls { display: flex; flex-wrap: wrap; gap: 0.6rem; align-items: center; margin-block: 0.5rem; }
  .nh-player-controls label { display: inline-flex; gap: 0.3rem; align-items: center; }
</style>
