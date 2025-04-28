<script setup lang="ts">
  import { computed } from 'vue';
  import { store } from '@/store';
  import { DashboardGroup, DashboardGroups, DeviceGroup } from '@/types/settings.type';
  import EntityCard from '../expose/EntityCard.vue';

  const dashboardGroups = computed(() => {
    return store.getters['hub/dashboardGroups']() as DashboardGroups;
  });

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
  // function calculateColumnCount(length: number): number {
  //   if (length >= 6) return 3;
  //   if (length >= 3) return 2;
  //   return 1;
  // }
</script>
<template>
  <div class="dashboard-container">
    <div v-for="group in dashboardGroups" :key="group.name" class="dashboard-group">
      <h4 class="dashboard-title">{{ group.name }}</h4>
      <div class="grid-container">
        <div v-for="item in flattenDeviceGroup(group)" :key="`${item.deviceId}-${item.expose}`" class="grid-item">
          <EntityCard :id="item.deviceId" :name="item.expose" compact />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .dashboard-container {
    display: grid;
    gap: 1rem;
    grid-template-columns: repeat(auto-fit, minmax(min(250px, 100%), 1fr));
    justify-items: stretch;
  }

  .dashboard-group {
    padding: 1rem;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  }

  .dashboard-title {
    font-size: 1.2rem;
    font-weight: bold;
    margin-bottom: 1rem;
  }

  .grid-container {
    column-count: 2;
    column-gap: 0.5rem;
    max-width: 400px;
    /* margin: 0 auto; */
  }

  .grid-item {
    margin-bottom: 0.5rem;
    width: 100%;
    display: inline-block;
    break-inside: avoid;
  }
</style>
