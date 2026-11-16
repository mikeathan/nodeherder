<script setup lang="ts">
  /*
   * Home shortcut for wireless observation.
   *
   * A compact summary rather than a live radar: a canvas inside a tile grid is cramped, and the
   * radar needs room to be readable. This says how many sensors are reporting and how many devices
   * have been heard, and opens the real view.
   *
   * It exists because the radar is no longer a nav item. Without a shortcut the feature would be
   * reachable only through Integrations, which is a place you visit to configure things, not to
   * glance at them.
   */
  import { computed } from 'vue';
  import { useRouter } from 'vue-router';
  import { store } from '@/store';
  import { observationDevices, OBSERVATION_ROUTE } from '@/domain/integrations/device-backed';
  import { snapshotSummary } from '@/domain/observations/presentation';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{
    editing?: boolean;
    busy?: boolean;
  }>();

  const emit = defineEmits<{ remove: [] }>();
  const router = useRouter();

  const devices = computed(() => store.getters['hub/listAllDevices']?.() ?? []);
  const sensors = computed(() => observationDevices(devices.value));
  const snapshot = computed(() => store.getters['observations/snapshot']);

  /**
   * What the radar is currently seeing, or why it is not seeing anything.
   *
   * The two states need different wording: "no devices heard" is normal early on, while "no sensor"
   * means nothing is enrolled and the card should not pretend otherwise.
   */
  const summary = computed(() => {
    if (sensors.value.length === 0) return 'No sensor enrolled';
    if (!snapshot.value) return 'Waiting for the first reading';
    return snapshotSummary(snapshot.value);
  });

  function open(): void {
    void router.push({ name: OBSERVATION_ROUTE });
  }
</script>

<template>
  <article class="nh-observation-tile" aria-label="Wireless observation">
    <header class="nh-observation-head">
      <span class="nh-observation-icon" aria-hidden="true"><UiIcon name="signal" /></span>
      <h3 class="nh-observation-title">Wireless observation</h3>
    </header>

    <p class="nh-observation-summary">{{ summary }}</p>
    <p v-if="sensors.length > 0" class="nh-observation-sensors">
      Via {{ sensors.map((sensor) => sensor.friendly_name).join(', ') }}
    </p>

    <div class="nh-observation-actions">
      <UiButton size="sm" variant="primary" :disabled="props.busy" @click="open">View radar</UiButton>
      <UiButton v-if="props.editing" size="sm" variant="ghost" :disabled="props.busy" @click="emit('remove')">
        Remove
      </UiButton>
    </div>
  </article>
</template>

<style scoped>
  .nh-observation-tile {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    padding: 0.75rem;
    height: 100%;
    border: 1px solid var(--border-subtle, #dee2e6);
    border-radius: 0.6rem;
    background: var(--surface-raised, rgba(255, 255, 255, 0.02));
  }

  .nh-observation-head {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .nh-observation-icon {
    color: var(--accent-strong, #4a9eff);
  }

  .nh-observation-title {
    margin: 0;
    font-size: 0.85rem;
    font-weight: 600;
  }

  .nh-observation-summary {
    margin: 0;
    font-size: 0.78rem;
    color: var(--text-muted, #8b9bab);
  }

  .nh-observation-sensors {
    margin: 0;
    font-size: 0.7rem;
    color: var(--text-subtle, #7c8b9c);
  }

  /* Pushed to the bottom so tiles of different content still align their actions. */
  .nh-observation-actions {
    display: flex;
    gap: 0.35rem;
    margin-top: auto;
    padding-top: 0.4rem;
  }
</style>
