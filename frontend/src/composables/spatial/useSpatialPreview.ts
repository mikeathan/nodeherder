import { computed, ref, watch } from 'vue';
import { store } from '@/store';
import { findDevice, statusOf } from '@/composables/useHub';
import { useDeviceSettings } from '@/composables/useDeviceSettings';
import { copyWorkspace } from '@/domain/spatial/workspace';
import { frameSamples } from '@/domain/spatial/frames';
import { sensorProjectedPoint } from '@/domain/spatial/canvas';
import type { SpatialDisplayTrack } from '@/domain/spatial/tracks';

const PREVIEW_INTERVAL_MS = 500;

/**
 * Lightweight Home preview: reads the latest accepted frame directly and renders at most every
 * 500 ms. It deliberately skips smoothing, association and hold state so a card stays cheap.
 */
export function useSpatialPreview(deviceId: () => string, active: () => boolean) {
  const device = computed(() => findDevice(deviceId()));
  const { settings } = useDeviceSettings(deviceId);
  const frame = computed(() => store.getters['hub/spatialFrame'](deviceId()) as ReturnType<typeof store.getters['hub/spatialFrame']>);
  const workspace = computed(() => {
    const value = copyWorkspace(settings.value.spatial?.workspace);
    // A card preview shows the sensor and targets, not a full room-scene render.
    return { ...value, zones: [], outline: [], objects: [] };
  });
  const samples = ref<{ slot: number; x: number; y: number }[]>([]);
  const updatedAt = ref(0);
  let lastUpdate = 0;

  function capture(): void {
    if (!active() || !frame.value) return;
    const now = Date.now();
    if (updatedAt.value && now - lastUpdate < PREVIEW_INTERVAL_MS) return;
    lastUpdate = now;
    samples.value = frameSamples(frame.value)
      .filter((sample) => sample.active && sample.x !== null && sample.y !== null)
      .map((sample) => ({ slot: sample.slot, x: sample.x as number, y: sample.y as number }));
    updatedAt.value = now;
  }

  watch(frame, capture, { immediate: true });
  watch(active, (on) => { if (on) { lastUpdate = 0; capture(); } });

  const viewBox = computed(() => {
    const projected = samples.value.map((sample) => sensorProjectedPoint(sample, workspace.value.sensor, 0));
    const points = [{ x: 0, y: 0 }, ...projected];
    const xs = points.map((point) => point.x);
    const ys = points.map((point) => point.y);
    const minX = Math.min(...xs);
    const maxX = Math.max(...xs);
    const minY = Math.min(...ys);
    const maxY = Math.max(...ys);
    const pad = Math.max(750, Math.max(maxX - minX, maxY - minY) * 0.2);
    return `${minX - pad} ${minY - pad} ${Math.max(2000, maxX - minX + pad * 2)} ${Math.max(2000, maxY - minY + pad * 2)}`;
  });

  const tracks = computed<SpatialDisplayTrack[]>(() =>
    samples.value.map((sample) => ({
      id: sample.slot,
      label: `Slot ${sample.slot}`,
      slot: sample.slot,
      x: sample.x,
      y: sample.y,
      rawX: sample.x,
      rawY: sample.y,
      vx: 0,
      vy: 0,
      at: updatedAt.value,
      colorIndex: (sample.slot - 1) % 4,
      state: 'live' as const,
    }))
  );

  return {
    workspace,
    viewBox,
    tracks,
    updatedAt,
    online: computed(() => !!device.value && statusOf(device.value) === 'online'),
    hasFrame: computed(() => tracks.value.length > 0),
  };
}
