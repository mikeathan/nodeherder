<script setup lang="ts">
  /* Devices: every device as a card with live values and controls (spec 007 US-03). */
  import { computed, ref } from 'vue';
  import { RouteName } from '@/types/router';
  import { useHub } from '@/composables/useHub';
  import { DEFAULT_FILTERS, DeviceFilters, filterDevices } from '@/domain/network';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import DeviceCard from '@/components/entity/DeviceCard.vue';
  import DeviceFilterBar from '@/components/entity/DeviceFilterBar.vue';

  const { devices, initialized, isDeviceDisabled } = useHub();
  const filters = ref<DeviceFilters>({ ...DEFAULT_FILTERS });
  const shown = computed(() => filterDevices(devices.value, filters.value, (d) => isDeviceDisabled(d.id)));
</script>

<template>
  <div>
    <UiPageHeader title="Devices" :subtitle="`${devices.length} devices`">
      <template #actions>
        <UiButton icon="list" :to="{ name: RouteName.DeviceList }">Device list</UiButton>
      </template>
    </UiPageHeader>
    <DeviceFilterBar v-model="filters" :devices="devices" :shown="shown.length" />
    <div v-if="!initialized && !devices.length" class="nh-dcards" aria-busy="true">
      <div v-for="n in 6" :key="n" class="nh-card nh-skel-card"><span class="nh-skel" /><span class="nh-skel" /><span class="nh-skel" /></div>
    </div>
    <UiEmpty v-else-if="!devices.length" icon="devices" title="No devices yet" text="Pair a Zigbee device with Permit join, or send data from a Wi-Fi device to the hub." />
    <UiEmpty v-else-if="!shown.length" icon="search" title="No devices match" text="Try a different search or clear the filters.">
      <UiButton @click="filters = { ...DEFAULT_FILTERS }">Clear filters</UiButton>
    </UiEmpty>
    <div v-else class="nh-dcards">
      <DeviceCard v-for="d in shown" :key="d.id" :device="d" />
    </div>
  </div>
</template>

<style scoped>
  .nh-dcards {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(100%, 19rem), 1fr));
    gap: var(--nh-space);
    align-items: start;
  }
  .nh-skel-card {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 1rem;
  }
</style>
