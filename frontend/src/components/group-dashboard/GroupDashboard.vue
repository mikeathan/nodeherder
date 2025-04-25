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
  // TESTING
</script>
<style scoped>
  /* .dashboard {
  background-color: var(--p-card-background);
  height: 100vh;
  overflow: auto;
} */
</style>

<template>
  <!-- <div class="grid gap-2 p-2" style="margin: 0; padding: 0.5rem">
    <div class="col-12 sm:col-6 md:col-4 lg:col-3 xl:col-2" v-for="group in dashboardGroups" :key="group.name">
      <div v-for="device in group.deviceGroup" :key="device.deviceId">
        <div v-for="expose in device.exposes" :key="expose">
          <EntityCard :id="device.deviceId" :name="expose" compact />
        </div>
      </div>
    </div>
  </div>  -->

  <div v-for="group in dashboardGroups" :key="group.name" class="mb-4">
    <h4 class="mt-2">{{ group.name }}</h4>
    <div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 4px; width: max-content">
      <div
        v-for="item in flattenDeviceGroup(group)"
        :key="item.deviceId + '-' + item.expose"
        style="display: flex; align-items: flex-end">
        <EntityCard :id="item.deviceId" :name="item.expose" compact />
      </div>
    </div>
  </div>
</template>
