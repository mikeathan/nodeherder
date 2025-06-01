<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '@/store';
  import { DashboardGroup, DashboardGroups, DeviceGroup } from '@/types/settings.type';
  import EntityCard from './cards/EntityCard.vue';

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
  const props = defineProps({
    editMode: { type: Boolean, default: false },
  });
</script>
<template>
  <div class="dashboard-container">
    <div v-for="group in dashboardGroups" :key="group.name" class="dashboard-group">
      <h4 class="dashboard-title">{{ group.name }}</h4>
      <div class="card-container" :class="{ 'edit-mode': editMode }">
        <div v-if="editMode" class="tools">
          <span class="pi pi-pen-to-square" />
          <span class="pi pi-trash" />
          <span class="pi pi-plus add-icon" /> ?? maybethey can all move to the bottom ?
        </div>
        <div v-for="item in flattenDeviceGroup(group)" :key="`${item.deviceId}-${item.expose}`" class="card-item">
          <EntityCard :id="item.deviceId" :name="item.expose" compact />
        </div>
        <div v-if="editMode" class="add-icon-tool">
          <span class="pi pi-plus add-icon" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .dashboard-container {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    max-width: 100%;
  }

  .dashboard-group {
    padding: 0.8rem;
  }

  .dashboard-title {
    font-size: 1.2rem;
    font-weight: bold;
    margin-bottom: 1rem;
  }

  .card-container {
    column-count: 2;
    column-gap: 0.5rem;
    max-width: 400px;
    /* margin: 0 auto; */

    border-radius: 12px;
    padding: 16px;
    position: relative;
    margin-bottom: 20px;
  }

  .card-container.edit-mode {
    border: 2px dotted #d3d3d3;
  }

  .card-item {
    margin-bottom: 0.5rem;
    width: 100%;
    display: inline-block;
    break-inside: avoid;
  }

  .tools {
    position: absolute;
    top: -30px;
    right: -10px;
    padding: 4px 8px;
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .add-icon-tool {
    border: 2px dotted #d3d3d3;
    border-radius: 8px;
    padding: 10px 12px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s ease;
    margin-top: 0.5rem;
    margin-left: -1rem;
  }

  .icon {
    font-size: 14px;
    cursor: pointer;
  }
</style>
