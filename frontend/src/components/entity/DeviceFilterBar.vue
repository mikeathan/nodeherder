<script setup lang="ts">
  /* Search and filters shared by Devices and Device list (domain/network.ts filterDevices). */
  import { computed } from 'vue';
  import { Device } from '@/types/device';
  import { DeviceFilters } from '@/domain/network';
  import { resolveProtocol } from '@/domain/devices';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiSelect from '@/components/ui/UiSelect.vue';

  const props = defineProps<{ modelValue: DeviceFilters; devices: Device[]; shown: number }>();
  const emit = defineEmits<{ (e: 'update:modelValue', value: DeviceFilters): void }>();

  const set = <K extends keyof DeviceFilters>(key: K, value: DeviceFilters[K]) => emit('update:modelValue', { ...props.modelValue, [key]: value });

  const protocols = computed(() => {
    const ids = [...new Set(props.devices.map((d) => d.connection_type).filter(Boolean))].sort();
    return [{ value: 'all', label: 'All protocols' }, ...ids.map((id) => ({ value: id, label: resolveProtocol(id).label }))];
  });
  const statuses = [
    { value: 'all', label: 'Any status' },
    { value: 'online', label: 'Online' },
    { value: 'offline', label: 'Offline' },
    { value: 'disabled', label: 'Disabled' },
  ];
  const powers = [
    { value: 'all', label: 'Any power' },
    { value: 'battery', label: 'Battery' },
    { value: 'mains', label: 'Mains' },
  ];
</script>

<template>
  <div class="nh-filterbar" role="search">
    <label class="nh-search">
      <UiIcon name="search" />
      <span class="sr-only">Search devices</span>
      <input class="nh-input" type="search" placeholder="Search name, address or model" :value="modelValue.query" @input="set('query', ($event.target as HTMLInputElement).value)" />
    </label>
    <UiSelect :model-value="modelValue.status" :options="statuses" label="Status" @update:model-value="set('status', $event as DeviceFilters['status'])" />
    <UiSelect v-if="protocols.length > 2" :model-value="modelValue.protocol" :options="protocols" label="Protocol" @update:model-value="set('protocol', $event as string)" />
    <UiSelect :model-value="modelValue.power" :options="powers" label="Power source" @update:model-value="set('power', $event as DeviceFilters['power'])" />
    <span class="nh-filter-count" aria-live="polite">{{ shown }} of {{ devices.length }}</span>
  </div>
</template>

<style scoped>
  .nh-filterbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: var(--nh-space);
  }
  .nh-search {
    position: relative;
    display: flex;
    align-items: center;
    flex: 1 1 16rem;
    max-width: 26rem;
  }
  .nh-search > .nh-ic {
    position: absolute;
    left: 0.65rem;
    color: var(--nh-text-muted);
    pointer-events: none;
  }
  .nh-search .nh-input {
    width: 100%;
    padding-left: 2.2rem;
  }
  .nh-filter-count {
    margin-left: auto;
    font-size: 0.8rem;
    color: var(--nh-text-muted);
  }
  @media (max-width: 640px) {
    .nh-search {
      max-width: none;
      flex-basis: 100%;
    }
    .nh-filterbar > .nh-select {
      flex: 1 1 8rem;
    }
  }
</style>
