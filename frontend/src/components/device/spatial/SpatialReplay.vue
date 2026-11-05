<script setup lang="ts">
  import { computed } from 'vue';
  import type { ReplayFrame } from '@/domain/spatial/replay';
  import { frameAt } from '@/domain/spatial/replay';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    frames: readonly ReplayFrame[];
    time: number;
    playing: boolean;
    speed: number;
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
</script>

<template>
  <section aria-label="Session replay">
    <h3>Replay this session</h3>
    <p class="nh-muted">
      Recorded in this browser only (last five minutes); nothing is stored on the hub and replay never feeds
      automations or history. The canvas shows recorded raw positions while this view is open.
    </p>
    <template v-if="frames.length">
      <p role="status">{{ label }} · {{ current?.samples.length ?? 0 }} slot(s)</p>
      <label class="nh-replay-range">
        Replay time
        <input
          class="nh-input"
          type="range"
          :min="frames[0].at"
          :max="frames[frames.length - 1].at"
          :value="time"
          step="100"
          @input="emit('seek', Number(($event.target as HTMLInputElement).value))" />
      </label>
      <div class="nh-replay-controls">
        <UiButton size="sm" @click="emit('toggle')">{{ playing ? 'Pause' : 'Play' }}</UiButton>
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
    <p v-else class="nh-muted">No frames recorded yet. Keep the presence tab open while the target moves.</p>
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
