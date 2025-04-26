<script setup lang="ts">
  import { computed } from 'vue';
  import { store } from '@/store';
  import { DashboardGroup, DashboardGroups, DeviceGroup } from '@/types/settings.type';
  import EntityCard from '../expose/EntityCard.vue';

  const dashboardGroups = computed(() => {
    return store.getters['hub/dashboardGroups']() as DashboardGroups;
  });

  // TESTING just get first expose of thefirst group
  function getDeviceFromGroup(group: DashboardGroup): DeviceGroup {
    const deviceGroup = Object.entries(group.deviceGroup).at(0)?.[1];

    return deviceGroup as DeviceGroup;
  }

  function flattenDeviceGroup(group: DashboardGroup): any[] {
    return Object.entries(group.deviceGroup).flatMap(([key, value]) =>
      value.exposes.map((expose) => ({
        deviceId: value.deviceId,
        expose,
      }))
    );
  }

  // to use for dynamically setting the column count
  //   :style="{ columnCount: calculateColumnCount(flattenDeviceGroup(group).length) }"
  function calculateColumnCount(length: number): number {
    if (length >= 6) return 3;
    if (length >= 3) return 2;
    return 1;
  }

  // TESTING
</script>
<template>
  <div v-for="group in dashboardGroups" :key="group.name" class="mb-4">
    <h4 class="mt-2">{{ group.name }}</h4>
    <div class="grid-container">
      <div v-for="item in flattenDeviceGroup(group)" :key="item.deviceId + '-' + item.expose" class="grid-item">
        <EntityCard :id="item.deviceId" :name="item.expose" compact />
      </div>
    </div>
  </div>
</template>

<style scoped>
  .grid-container {
    column-count: 3;
    column-gap: 0.5rem;
    max-width: 520px;
    margin: 0 auto;
  }

  .grid-item {
    margin-bottom: 0.5rem;
    width: 100%;
    display: inline-block;
    break-inside: avoid;
  }
</style>
