<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '@/store';
  import { DashboardGroup, DashboardGroups, DeviceGroup } from '@/types/settings.type';
  import EntityCard from './cards/EntityCard.vue';
  import {
    emitOpenConfirmationDialog,
    emitOpenDeviceGroupSelectionDialog,
    emitOpenDeviceSelectionDialog,
    emitOpenExposeSelectionDialog,
  } from '@/contracts/dialog-events';
  import { DeviceGroupEventAction } from '@/types/events.type';

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

  function openAddDeviceExposeDialog(dashboardroup: DashboardGroup) {
    const props = {
      dashboardGroup: dashboardroup,
      title: 'Select Device group entities',
    };
    emitOpenDeviceGroupSelectionDialog((args: DeviceGroup) => addNewDeviceExpose(dashboardroup, args), props);
  }

  function addNewDeviceExpose(dashboardGroup: DashboardGroup, deviceGroup: DeviceGroup) {
    const deviceId = deviceGroup.deviceId;
    dashboardGroup.deviceGroup[deviceId] = deviceGroup;
    store.dispatch('hub/saveDashboardGroup', dashboardGroup as DashboardGroup);
  }

  function openDeleteDeviceGroupConfirmationDialog(groupName: string) {
    const props = {
      title: 'Question',
      message: 'Are you sure?',
    };
    emitOpenConfirmationDialog(() => deleteDeviceGroup(groupName), props);
  }

  function deleteDeviceGroup(groupName: string) {
    if (!groupName) {
      console.log('groupName is empty');
      return;
    }

    delete dashboardGroups.value[groupName];
    store.dispatch('hub/deleteDashboardGroup', groupName);
  }
</script>
<template>
  <div class="dashboard-container">
    <div v-for="group in dashboardGroups" :key="group.name" class="dashboard-group">
      <h4 class="dashboard-title">{{ group.name }}</h4>
      <div class="card-container" :class="{ 'edit-mode': editMode }">
        <div v-for="item in flattenDeviceGroup(group)" :key="`${item.deviceId}-${item.expose}`" class="card-item">
          <EntityCard :id="item.deviceId" :name="item.expose" compact :edit-mode="editMode"/>
        </div>
        <div v-if="editMode" class="icon-tools">
          <span class="edit-icon pi pi-pen-to-square" />
          <span class="edit-icon pi pi-trash" @click="openDeleteDeviceGroupConfirmationDialog(group.name)" />
          <span class="edit-icon pi pi-plus" @click="openAddDeviceExposeDialog(group)" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .dashboard-container {
    display: flex;
    flex-wrap: wrap;
    gap: 0.8rem;
    max-width: 100%;
    padding-top: 1rem;
  }

  /* Deskop view */
  .dashboard-group {
    padding: 0.1rem;
  }

  /* Mobile view */
  @media (max-width: 768px) {
    .dashboard-group {
      padding: 0;
      flex: 1 1 200px;
      max-width: 100%;
      box-sizing: border-box;
    }
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

  .icon-tools {
    display: flex;
    justify-content: flex-start;
    gap: 12px;
    margin-top: 12px;
    width: 100%;
    column-span: all;
  }

  .edit-icon {
    border: 2px dotted #d3d3d3;
    border-radius: 8px;
    padding: 12px 14px;
    cursor: pointer;
    font-size: 16px;
    transition: background-color 0.2s ease;
  }
</style>
