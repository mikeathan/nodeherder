<script setup lang="ts">
  /*
   * Session diagnostics from the replay recorder plus a guided clear-delay measurement against
   * the hub's derived zone occupancy. Measurements are labelled; no cause is inferred.
   */
  import { computed, onBeforeUnmount, ref, watch } from 'vue';
  import type { Device } from '@/types/device';
  import type { SpatialWorkspace } from '@/types/spatial';
  import type { ReplayFrame } from '@/domain/spatial/replay';
  import {
    DIAGNOSTIC_STABLE_MM,
    DIAGNOSTIC_UNSETTLED_MM,
    diagnosticRecording,
    overlapSummary,
    slotDiagnostics,
  } from '@/domain/spatial/diagnostics';
  import { useDerivedPresence } from '@/composables/spatial/useDerivedPresence';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    device: Device;
    workspace: SpatialWorkspace;
    frames: readonly ReplayFrame[];
  }>();
  const derived = useDerivedPresence(() => props.device);
  const recording = computed(() => diagnosticRecording(props.frames));
  const slots = computed(() => slotDiagnostics(props.frames));
  const overlap = computed(() => overlapSummary(props.frames));

  const zoneId = ref(props.workspace.zones[0]?.id ?? '');
  const phase = ref<'idle' | 'waiting-occupied' | 'waiting-clear' | 'done' | 'timeout'>('idle');
  const enteredAt = ref<number | null>(null);
  const clearedAt = ref<number | null>(null);
  let timeout: ReturnType<typeof setTimeout> | null = null;
  const zoneOccupied = computed(() => derived.zones.value.get(zoneId.value)?.occupied ?? null);
  const measuredMs = computed(() =>
    enteredAt.value !== null && clearedAt.value !== null ? clearedAt.value - enteredAt.value : null
  );

  function stopTimeout() {
    if (timeout) clearTimeout(timeout);
    timeout = null;
  }
  function reset() {
    stopTimeout();
    phase.value = 'idle';
    enteredAt.value = null;
    clearedAt.value = null;
  }
  function start() {
    reset();
    if (!zoneId.value) return;
    phase.value = 'waiting-occupied';
    timeout = setTimeout(() => {
      if (phase.value !== 'done') phase.value = 'timeout';
    }, 120_000);
  }
  watch(zoneOccupied, (occupied) => {
    if (phase.value === 'waiting-occupied' && occupied === true) {
      enteredAt.value = Date.now();
      phase.value = 'waiting-clear';
    } else if (phase.value === 'waiting-clear' && occupied === false) {
      clearedAt.value = Date.now();
      stopTimeout();
      phase.value = 'done';
    }
  });
  onBeforeUnmount(stopTimeout);
</script>

<template>
  <section aria-label="Spatial diagnostics">
    <h3>Diagnostics</h3>
    <p class="nh-muted">
      Session measurements from recorded frames. They are evidence, not firmware guarantees, and they do not identify
      a cause.
    </p>
    <p v-if="!recording.frames" class="nh-muted">
      No frames recorded yet. Keep this view open while the target moves for a few seconds.
    </p>
    <template v-else>
      <p>
        {{ recording.frames }} frames over {{ Math.round(recording.durationMs / 1000) }} s
        <span v-if="recording.durationMs">· {{ new Date(recording.from).toLocaleTimeString() }}–{{ new Date(recording.to).toLocaleTimeString() }}</span>
      </p>
      <table class="nh-diagnostics-table">
        <caption>Per-slot stability (σ thresholds {{ DIAGNOSTIC_STABLE_MM }} / {{ DIAGNOSTIC_UNSETTLED_MM }} mm)</caption>
        <thead>
          <tr>
            <th scope="col">Slot</th>
            <th scope="col">Samples</th>
            <th scope="col">Interval</th>
            <th scope="col">σ x / y</th>
            <th scope="col">Movement</th>
            <th scope="col">Rating</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="slot in slots" :key="slot.slot">
            <th scope="row">{{ slot.slot }}</th>
            <td>{{ slot.samples }}</td>
            <td>{{ slot.meanIntervalMs === null ? '—' : `${Math.round(slot.meanIntervalMs)} ms` }}</td>
            <td>{{ Math.round(slot.sigmaX) }} / {{ Math.round(slot.sigmaY) }} mm</td>
            <td>{{ Math.round(slot.movementMm) }} mm</td>
            <td>{{ slot.stability }}</td>
          </tr>
        </tbody>
      </table>
      <p>
        Concurrent slots within 350 mm: {{ overlap.frames }} frame(s)<template v-if="overlap.closestMm !== null"> ·
        closest {{ Math.round(overlap.closestMm) }} mm between slots {{ overlap.pair?.join(' and ') }}</template>.
        This is evidence of overlapping returns, not proof of a ghost.
      </p>
    </template>

    <h3>Measured zone clear delay</h3>
    <p class="nh-muted">
      Measures the time from the hub's derived occupancy becoming true to becoming false. It includes the hub exit
      grace, so it is an upper bound relative to the firmware off delay.
    </p>
    <label>
      Zone
      <select v-model="zoneId" class="nh-input" :disabled="phase !== 'idle'" @change="reset">
        <option value="">Not selected</option>
        <option v-for="zone in workspace.zones" :key="zone.id" :value="zone.id">{{ zone.name }}</option>
      </select>
    </label>
    <p>Derived zone state: {{ zoneOccupied === null ? 'unknown' : zoneOccupied ? 'occupied' : 'clear' }}</p>
    <div class="nh-diagnostic-actions">
      <UiButton size="sm" :disabled="!zoneId || phase === 'waiting-occupied' || phase === 'waiting-clear'" @click="start">
        Start measurement
      </UiButton>
      <UiButton size="sm" variant="ghost" @click="reset">Reset</UiButton>
    </div>
    <p v-if="phase === 'waiting-occupied'">Waiting for the zone to become occupied…</p>
    <p v-else-if="phase === 'waiting-clear'">Occupied since {{ enteredAt ? new Date(enteredAt).toLocaleTimeString() : '' }}; waiting for it to clear…</p>
    <p v-else-if="phase === 'done' && measuredMs !== null" role="status">
      Measured clear delay: {{ (measuredMs / 1000).toFixed(1) }} s (includes hub exit grace).
    </p>
    <p v-else-if="phase === 'timeout'" role="status">No complete enter/clear transition within two minutes.</p>
  </section>
</template>

<style scoped>
  .nh-diagnostics-table {
    width: 100%;
    border-collapse: collapse;
    margin-block: 0.5rem;
  }
  .nh-diagnostics-table th,
  .nh-diagnostics-table td {
    text-align: start;
    padding: 0.25rem 0.5rem;
    border-block-end: 1px solid var(--nh-border);
  }
  .nh-diagnostic-actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    margin-block: 0.5rem;
  }
</style>
