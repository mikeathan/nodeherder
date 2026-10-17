<script setup lang="ts">
  // A single settings surface serves global defaults and device exceptions.
  // Containers supply only identity; all persistence remains in useDeviceSettings.
  import { computed } from 'vue';
  import { store } from '@/store';
  import { useDeviceSettings } from '@/composables/useDeviceSettings';
  import { emitOpenConfirmationDialog } from '@/contracts/dialog-events';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import SensorReportingSettings from './SensorReportingSettings.vue';
  import type { DeviceConfig } from '@/types/settings.type';

  const props = withDefaults(defineProps<{ id?: string }>(), { id: '' });
  const { settings, hasOverride, save, deleteOverride } = useDeviceSettings(() => props.id);
  const defaults = computed<DeviceConfig>(() => store.getters['hub/deviceDefaults']());
  const local = computed<Partial<DeviceConfig>>(() => props.id && !hasOverride.value ? {} : settings.value);
  const offline = computed(() => store.getters['ws/getConnectionStatus'] !== 'connected');
  function removeOverride() {
    emitOpenConfirmationDialog(deleteOverride, { title: 'Restore device defaults', message: 'Remove all settings overrides for this device and inherit the current defaults?' });
  }
</script>

<template>
  <div v-if="defaults" class="device-configuration">
    <header>
      <h2 v-if="id">Device settings</h2>
      <p>{{ id ? 'Use shared defaults or adjust reporting for this device.' : 'Shared device settings. Individual sensor rules take priority over category defaults.' }}</p>
    </header>
    <section class="device-options" aria-label="Device options">
      <div>
        <div>
          <strong>Device enabled</strong>
          <p>Disabled devices do not process incoming readings.</p>
        </div>
        <UiToggle
          :model-value="!(local.disabled ?? defaults.disabled)"
          label="Device enabled"
          :disabled="offline"
          @update:model-value="save({ disabled: !$event })" />
      </div>
      <div>
        <div>
          <strong>Record history</strong>
          <p>Store eligible readings for charts using the history intervals below.</p>
        </div>
        <UiToggle
          :model-value="local.metricsEnabled ?? defaults.metricsEnabled"
          label="Record history"
          :disabled="offline"
          @update:model-value="save({ metricsEnabled: $event })" />
      </div>
    </section>
    <SensorReportingSettings :id="id" :config="local" :defaults="defaults" @update="save" />
    <div v-if="id && hasOverride" class="device-reset">
      <UiButton variant="ghost" :disabled="offline" @click="removeOverride">Restore device defaults</UiButton>
      <p>Removes every override for this device.</p>
    </div>
  </div>
  <p v-else role="status" class="nh-muted">Loading device settings…</p>
</template>

<style scoped>
  .device-configuration { display: grid; gap: 1.25rem; min-width: 0; width: 100%; max-width: 62rem; }
  h2 { margin: 0; font-size: 1.2rem; font-weight: 600; }
  p { color: var(--nh-text-muted); font-size: .85rem; line-height: 1.6; margin: .4rem 0 0; }
  .device-options { padding: .25rem 1.25rem; border: 1px solid var(--nh-border); border-radius: var(--nh-radius-m); }
  .device-options > div { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding: 1rem 0; }
  .device-options > div + div { border-top: 1px solid var(--nh-border); }
  strong { font-size: .9rem; font-weight: 600; }
  .device-reset { display: flex; align-items: center; flex-wrap: wrap; gap: .75rem; }
  .device-reset p { margin: 0; font-size: .78rem; }
</style>
