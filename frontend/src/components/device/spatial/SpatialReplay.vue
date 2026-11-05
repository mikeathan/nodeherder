<script setup lang="ts">
  import { computed } from 'vue';
  import type { ReplayFrame } from '@/domain/spatial/replay';
  import { firstRecordedFrame, frameAt, lastRecordedFrame } from '@/domain/spatial/replay';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    frames: readonly ReplayFrame[];
    time: number;
    playing: boolean;
    speed: number;
    hasPositions: boolean;
  }>();
  const emit = defineEmits<{
    seek: [at: number];
    toggle: [];
    'update:speed': [speed: number];
  }>();
  const current = computed(() => frameAt(props.frames, props.time));
  const label = computed(() =>
    current.value ? new Date(current.value.at).toLocaleTimeString() : 'No recorded frames'
  );
  const recordedFrames = computed(() => props.frames.filter((frame) => frame.samples.length > 0));
  const first = computed(() => firstRecordedFrame(props.frames));
  const last = computed(() => lastRecordedFrame(props.frames));
  const span = computed(() =>
    first.value && last.value
      ? `${new Date(first.value.at).toLocaleTimeString()}–${new Date(last.value.at).toLocaleTimeString()}`
      : ''
  );
</script>

<template>
  <section aria-label="Session replay">
    <h3>Replay this session</h3>
    <p class="nh-muted">
      This is a rolling five-minute buffer held in this browser tab's memory — there is no recordings list, nothing
      is written to the hub, and reloading the page clears it. Playback never feeds automations or history.
    </p>
    <template v-if="hasPositions">
      <p role="status">
        {{ label }} · {{ current?.samples.length ?? 0 }} slot(s) ·
        {{ recordedFrames.length }} frame(s) with positions · {{ span }}
      </p>
      <label class="nh-replay-range">
        Replay time
        <input
          class="nh-input"
          type="range"
          :min="first?.at ?? frames[0].at"
          :max="last?.at ?? frames[frames.length - 1].at"
          :value="time"
          step="100"
          @input="emit('seek', Number(($event.target as HTMLInputElement).value))" />
      </label>
      <div class="nh-replay-controls">
        <UiButton size="sm" :disabled="!hasPositions" @click="emit('toggle')">{{ playing ? 'Pause' : 'Play' }}</UiButton>
        <label>
          Speed
          <select
            class="nh-input"
            :value="speed"
            @change="emit('update:speed', Number(($event.target as HTMLSelectElement).value))">
            <option :value="1">1×</option>
            <option :value="2">2×</option>
            <option :value="4">4×</option>
          </select>
        </label>
      </div>
    </template>
    <p v-else-if="frames.length" class="nh-muted">
      Frames have arrived, but no target positions are recorded yet. Replay is disabled until a target reports.
    </p>
    <p v-else class="nh-muted">
      Nothing recorded in this tab yet. Keep the Presence tab open while the target moves once.
    </p>
  </section>
</template>

<style scoped>
  .nh-replay-range {
    display: grid;
    gap: 0.3rem;
  }
  .nh-replay-controls {
    display: flex;
    gap: 0.75rem;
    align-items: end;
    flex-wrap: wrap;
    margin-block-start: 0.5rem;
  }
  .nh-replay-controls label {
    display: grid;
    gap: 0.3rem;
  }
</style>
