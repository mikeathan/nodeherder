<script setup lang="ts">
  /*
   * Entry for the device-backed wireless observation integration.
   *
   * Sits inside the board as a group of one, using the same row layout as a configured instance, so
   * it reads as a peer rather than something appended to the page. What it cannot have is a schedule
   * or a refresh: there is nothing to fetch. Where those would be, it says what it reads from and
   * whether that is reporting.
   *
   * It has no "remove": there is no configuration to delete, only a device to un-enrol. In edit mode
   * it can be hidden like any other entry, which is the closest equivalent and is persisted with the
   * rest of the board.
   */
  import { computed } from 'vue';
  import { useRouter } from 'vue-router';
  import { useHub } from '@/composables/useHub';
  import { observationDevices, OBSERVATION_ROUTE } from '@/domain/integrations/device-backed';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{
    editing: boolean;
    busy: boolean;
  }>();

  const emit = defineEmits<{ hide: [] }>();
  const router = useRouter();

  const { devices } = useHub();
  const sensors = computed(() => observationDevices(devices.value));

  /** Devices declaring the capability that are not currently online. */
  const offline = computed(() =>
    sensors.value.filter((device) => device.availability !== 'online').map((device) => device.friendly_name),
  );

  const summary = computed(() => {
    const count = sensors.value.length;
    return `Reading through ${count} ${count === 1 ? 'sensor' : 'sensors'}`;
  });

  function openRadar(): void {
    void router.push({ name: OBSERVATION_ROUTE });
  }
</script>

<template>
  <article class="observation-entry" aria-label="Wireless observation">
    <span class="observation-entry-icon" aria-hidden="true"><UiIcon name="signal" /></span>

    <div class="observation-entry-main">
      <div class="observation-entry-heading">
        <h3>
          <RouterLink :to="{ name: OBSERVATION_ROUTE }" class="observation-entry-link">
            Wireless observation
          </RouterLink>
        </h3>
        <span class="observation-entry-type">Bluetooth</span>
      </div>

      <p class="observation-entry-summary">
        <strong>{{ summary }}</strong>
        <span>{{ sensors.map((sensor) => sensor.friendly_name).join(', ') }}</span>
      </p>

      <!-- The line another entry uses for its schedule. This has none, so it says what it does
           instead — the absence of a refresh action should not read as a missing control. -->
      <p class="observation-entry-schedule">Updated continuously while a sensor is connected</p>

      <p v-if="offline.length > 0" class="observation-entry-status is-warning">
        Not reporting: {{ offline.join(', ') }}
      </p>
      <p v-else class="observation-entry-status">All sensors reporting</p>
    </div>

    <div class="observation-entry-actions">
      <UiButton size="sm" variant="primary" :disabled="props.busy" @click="openRadar">Open radar</UiButton>
      <!-- Hiding is the edit-mode action for every entry; this follows the same rule. -->
      <UiButton
        v-if="props.editing"
        size="sm"
        variant="ghost"
        :disabled="props.busy"
        @click="emit('hide')">
        Hide
      </UiButton>
    </div>
  </article>
</template>

<style scoped>
  /*
   * Its own class family, matching the instance entry's proportions rather than sharing
   * `.integration-entry`: that class means "a configured instance" to the board and to six browser
   * tests, so a second element carrying it silently changed their counts.
   */
  .observation-entry {
    display: flex;
    align-items: flex-start;
    gap: 0.8rem;
    min-width: 0;
    padding: 0.9rem 1rem;
    border: var(--nh-border-w) solid var(--nh-border, #2a3441);
    border-radius: var(--nh-radius-l, 0.6rem);
    background: var(--nh-surface, transparent);
  }

  .observation-entry-icon {
    display: grid;
    place-items: center;
    flex: none;
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 50%;
    background: var(--nh-surface-2, rgba(74, 158, 255, 0.14));
    color: var(--nh-accent, #4a9eff);
  }

  .observation-entry-icon :deep(svg) {
    width: 1.3rem;
    height: 1.3rem;
  }

  .observation-entry-main {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    min-width: 0;
    flex: 1;
  }

  .observation-entry-heading {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .observation-entry-heading h3 {
    margin: 0;
    font-size: 0.95rem;
    overflow-wrap: anywhere;
  }

  .observation-entry-link {
    text-decoration: none;
  }

  .observation-entry-link:hover {
    text-decoration: underline;
  }

  .observation-entry-type {
    color: var(--nh-text-muted, #8b9bab);
    font-size: 0.72rem;
  }

  .observation-entry-summary {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 0.6rem;
    margin: 0.15rem 0 0;
    font-size: 0.84rem;
    overflow-wrap: anywhere;
  }

  .observation-entry-summary strong {
    font-weight: 600;
  }

  .observation-entry-schedule {
    margin: 0.15rem 0 0;
    color: var(--nh-text-muted, #8b9bab);
    font-size: 0.75rem;
  }

  .observation-entry-status {
    margin: 0.2rem 0 0;
    font-size: 0.75rem;
    color: var(--nh-text-muted, #8b9bab);
  }

  .observation-entry-status.is-warning {
    color: var(--nh-warning, #d9a441);
  }

  .observation-entry-actions {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    flex: none;
  }

  @media (max-width: 640px) {
    .observation-entry {
      flex-wrap: wrap;
    }

    .observation-entry-actions {
      width: 100%;
      justify-content: flex-end;
    }
  }
</style>
