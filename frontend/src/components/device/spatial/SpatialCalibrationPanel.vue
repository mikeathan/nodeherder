<script setup lang="ts">
  /*
   * Paired-point calibration: a live slot's raw measured position plus its known room
   * position determines the rigid sensor pose. Solver rules live in domain/spatial/calibration.ts.
   */
  import { computed, reactive, ref } from 'vue';
  import type { Device } from '@/types/device';
  import type { SpatialWorkspace } from '@/types/spatial';
  import type { SpatialDisplayTrack } from '@/domain/spatial/tracks';
  import type { CalibrationPair } from '@/domain/spatial/calibration';
  import { CALIBRATION_MAX_PAIRS, solveCalibration } from '@/domain/spatial/calibration';
  import { exposeDisplayLabel } from '@/domain/exposes';
  import { useDeviceCommand } from '@/composables/useDeviceCommand';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    device: Device;
    workspace: SpatialWorkspace;
    tracks: readonly SpatialDisplayTrack[];
    online: boolean;
  }>();
  const emit = defineEmits<{ apply: [patch: Partial<SpatialWorkspace['sensor']>] }>();
  const { send, isPending, canControl } = useDeviceCommand();
  const notice = ref('');
  const rows = reactive(
    Array.from({ length: 2 }, () => ({
      slot: null as number | null,
      measuredX: null as number | null,
      measuredY: null as number | null,
      capturedAt: 0,
      roomX: '',
      roomY: '',
    }))
  );

  // Options keep every reported slot, including aged ones, so a selection cannot be silently
  // dropped by the 250 ms ageing tick; staleness is reported separately and blocks applying.
  const options = computed(() => [...props.tracks].sort((left, right) => left.slot - right.slot));
  const isLive = (slot: number | null) =>
    slot !== null && props.tracks.some((track) => track.slot === slot && track.state === 'live');
  const angleEntity = computed(() => {
    const entity = props.device.exposes[props.workspace.firmwareEntities?.installationAngle ?? ''];
    return entity && entity.type === 'numeric' && entity.access_mode === 'readwrite' ? entity : undefined;
  });

  // The measured position is captured when the slot is chosen; later ageing must not rewrite it.
  const pairs = computed<CalibrationPair[]>(() =>
    rows
      .filter((row) => row.slot !== null)
      .map((row) => ({
        slot: row.slot,
        measuredX: row.measuredX ?? Number.NaN,
        measuredY: row.measuredY ?? Number.NaN,
        roomX: Number(row.roomX),
        roomY: Number(row.roomY),
      }))
  );
  const solved = computed(() => solveCalibration(pairs.value));
  const staleRows = computed(() =>
    rows
      .map((row, index) => ({ row, index }))
      .filter(({ row }) => row.slot !== null && !isLive(row.slot))
      .map(({ index }) => index + 1)
  );
  const busy = computed(() => (angleEntity.value ? isPending(props.device.id, angleEntity.value.name) : false));

  function addRow() {
    if (rows.length < CALIBRATION_MAX_PAIRS) {
      rows.push({ slot: null, measuredX: null, measuredY: null, capturedAt: 0, roomX: '', roomY: '' });
    }
  }
  /** Captures the slot's raw position once; re-selecting the same slot refreshes the capture. */
  function capture(index: number) {
    const row = rows[index];
    const track = props.tracks.find((item) => item.slot === row.slot);
    row.measuredX = track?.rawX ?? null;
    row.measuredY = track?.rawY ?? null;
    row.capturedAt = track ? Date.now() : 0;
    notice.value = '';
  }
  function removeRow(index: number) {
    if (rows.length > 2) rows.splice(index, 1);
  }
  function reset() {
    const empty = () => ({ slot: null, measuredX: null, measuredY: null, capturedAt: 0, roomX: '', roomY: '' });
    rows.splice(0, rows.length, empty(), empty());
    notice.value = '';
  }
  function apply() {
    if (solved.value.error || !solved.value.solution || staleRows.value.length) return;
    emit('apply', {
      x: Math.round(solved.value.solution.x),
      y: Math.round(solved.value.solution.y),
      rotation: Math.round(solved.value.solution.rotation * 10) / 10,
    });
    notice.value = 'Sensor pose updated in the draft. Review it, then save the workspace.';
  }
  function writeAngle() {
    if (!angleEntity.value || !solved.value.solution || !props.online || !canControl(props.device)) return;
    notice.value = send(props.device, angleEntity.value.name, Math.round(solved.value.solution.rotation * 10) / 10)
      ? 'Installation angle requested. This is not proof that the firmware applied it.'
      : 'Angle request failed.';
  }
</script>

<template>
  <section aria-label="Calibration">
    <h3>Calibration</h3>
    <p class="nh-muted">
      Stand at marked room positions and capture the live target for each. Positions are raw measured millimetres;
      room positions are the same millimetres used by the room outline. A result only changes the unsaved draft.
    </p>
    <p v-if="staleRows.length" class="nh-muted">
      Row {{ staleRows.join(', ') }} lost its live slot; recapture it before solving.
    </p>
    <div v-for="(row, index) in rows" :key="index" class="nh-calibration-row">
      <label>
        Live slot
        <select class="nh-input" :value="row.slot ?? ''" @change="row.slot = Number(($event.target as HTMLSelectElement).value) || null; capture(index)">
          <option value="">Not captured</option>
          <option v-for="track in options" :key="track.slot" :value="track.slot">
            Slot {{ track.slot }} · raw {{ Math.round(track.rawX) }}, {{ Math.round(track.rawY) }} mm
            {{ track.state === 'live' ? '' : `(${track.state})` }}
          </option>
        </select>
      </label>
      <span v-if="row.capturedAt" class="nh-muted">
        Captured {{ row.measuredX }}, {{ row.measuredY }} mm at {{ new Date(row.capturedAt).toLocaleTimeString() }}
      </span>
      <label>
        Room X (mm)
        <input v-model="row.roomX" class="nh-input is-num" type="number" inputmode="numeric" />
      </label>
      <label>
        Room Y (mm)
        <input v-model="row.roomY" class="nh-input is-num" type="number" inputmode="numeric" />
      </label>
      <UiButton v-if="rows.length > 2" size="sm" variant="ghost" @click="removeRow(index)">Remove</UiButton>
    </div>
    <UiButton v-if="rows.length < CALIBRATION_MAX_PAIRS" size="sm" variant="ghost" @click="addRow">Add pair</UiButton>

    <template v-if="solved.error">
      <p class="nh-muted">{{ solved.error }}</p>
    </template>
    <template v-else-if="solved.solution">
      <dl class="nh-calibration-result">
        <dt>Rotation</dt>
        <dd>{{ solved.solution.rotation.toFixed(1) }}°</dd>
        <dt>Translation</dt>
        <dd>{{ Math.round(solved.solution.x) }}, {{ Math.round(solved.solution.y) }} mm</dd>
        <dt>Residual</dt>
        <dd>{{ Math.round(solved.solution.rms) }} mm RMS · worst {{ Math.round(solved.solution.worst) }} mm</dd>
      </dl>
      <p v-if="solved.warning" class="nh-muted">{{ solved.warning }}</p>
      <UiButton :disabled="staleRows.length > 0" @click="apply">Apply to room draft</UiButton>
      <UiButton
        v-if="angleEntity"
        variant="ghost"
        :disabled="!online || !canControl(device) || busy"
        @click="writeAngle">
        Write angle to firmware (unverified)
      </UiButton>
    </template>
    <p v-else class="nh-muted">Select a live slot and both room coordinates for at least two pairs.</p>
    <p v-if="angleEntity" class="nh-muted">
      Mapped angle entity: {{ exposeDisplayLabel(angleEntity) }}. Firmware angle semantics can differ; only write it
      when you accept that.
    </p>
    <UiButton size="sm" variant="ghost" @click="reset">Reset pairs</UiButton>
    <p v-if="notice" role="status">{{ notice }}</p>
  </section>
</template>

<style scoped>
  .nh-calibration-row {
    display: grid;
    grid-template-columns: minmax(0, 1.4fr) repeat(2, minmax(0, 1fr)) auto;
    gap: 0.5rem;
    align-items: end;
    padding-block: 0.4rem;
  }
  .nh-calibration-row label {
    display: grid;
    gap: 0.3rem;
  }
  .nh-calibration-result {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.2rem 0.8rem;
  }
  .nh-calibration-result dt {
    font-weight: 600;
  }
  .nh-calibration-result dd {
    margin: 0;
  }
  @media (max-width: 640px) {
    .nh-calibration-row {
      grid-template-columns: 1fr;
    }
  }
</style>
