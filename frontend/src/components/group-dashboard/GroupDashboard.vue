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
  <div class="grid">
    <div class="col-12 md:col-6 lg:col-3 xg:col-2" v-for="group in dashboardGroups" :key="group.name">
      <div v-for="device in group.deviceGroup" :key="device.deviceId">
        <div v-for="expose in device.exposes" :key="expose">
          <EntityCard :id="device.deviceId" :name="expose" />
        </div>
      </div>
    </div>
  </div>
</template>
