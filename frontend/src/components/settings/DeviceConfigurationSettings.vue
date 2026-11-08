<script setup lang="ts">
  // A single settings surface serves global defaults and device exceptions.
  // Containers supply only identity and heading; all persistence remains in useDeviceSettings.
  import { computed } from 'vue';
  import { store } from '@/store';
  import { useDeviceSettings } from '@/composables/useDeviceSettings';
  import { emitOpenConfirmationDialog } from '@/contracts/dialog-events';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiSettingRow from '@/components/ui/UiSettingRow.vue';
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
    <UiCard title="Device options" icon="settingsDevices" :heading-level="3">
      <UiSettingRow label="Device enabled" help="Disabled devices do not process incoming readings.">
        <UiToggle
          :model-value="!(local.disabled ?? defaults.disabled)"
          label="Device enabled"
          :disabled="offline"
          @update:model-value="save({ disabled: !$event })" />
      </UiSettingRow>
      <UiSettingRow label="Record history" help="Store eligible readings for charts using the history intervals below.">
        <UiToggle
          :model-value="local.metricsEnabled ?? defaults.metricsEnabled"
          label="Record history"
          :disabled="offline"
          @update:model-value="save({ metricsEnabled: $event })" />
      </UiSettingRow>
    </UiCard>

    <SensorReportingSettings :id="id" :config="local" :defaults="defaults" @update="save" />

    <div v-if="id && hasOverride" class="device-reset">
      <UiButton variant="ghost" icon="refresh" :disabled="offline" @click="removeOverride">Restore device defaults</UiButton>
      <p class="nh-help"><UiIcon name="info" /> Removes every override for this device.</p>
    </div>
  </div>
  <p v-else role="status" class="nh-muted">Loading device settings…</p>
</template>

<style scoped>
  .device-configuration {
    display: flex;
    flex-direction: column;
    gap: var(--nh-space);
    min-width: 0;
    width: 100%;
    max-width: 62rem;
  }
  .device-configuration > .nh-card + .nh-card {
    margin-top: 0;
  }
  .device-reset {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.75rem;
  }
  .device-reset p {
    margin: 0;
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
  }
</style>
