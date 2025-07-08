<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '@/store';
  import { DashboardGroup, DashboardGroups, DeviceGroup } from '@/types/settings.type';
  import { getDeviceGroupId } from '@/contracts/device-group';
  import EntityCard from './cards/EntityCard.vue';
  import {
    emitOpenConfirmationDialog,
    emitOpenDeviceGroupSelectionDialog,
    emitOpenInputDialogEvent,
  } from '@/contracts/dialog-events';

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

  const selectedCard = ref<string | null>(null);

  function handleCardSelected(id: string) {
    selectedCard.value = id;
  }

  function openRenameDashboardGroupDialog(groupName: string) {
    const props = {
      title: 'Rename Dashboard Group',
      message: 'Enter new name',
      value: groupName,
    };
    emitOpenInputDialogEvent((value) => renameDashboardGroup(groupName, value), props);
  }
  function openNewDashboardGroupDialog() {
    const props = {
      title: 'create new dashboard group',
    };
    emitOpenInputDialogEvent((value) => addNewDashboardGroup(value), props);
  }

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
      message: `Delete group ${groupName} ?`,
    };
    emitOpenConfirmationDialog(() => deleteDeviceGroup(groupName), props);
  }

  function openDeleteDeviceExposeConfirmationDialog(groupName: string, deviceId: string, exposeName: string) {
    const props = {
      title: 'Question',
      message: `Delete expose ${exposeName} ?`,
    };
    emitOpenConfirmationDialog(() => deleteDeviceExpose(groupName, deviceId, exposeName), props);
  }

  function addNewDashboardGroup(grouName: string) {
    if (!grouName) {
      // TODO: emit error message
      console.log('groupName is empty');
      return;
    }

    if (dashboardGroups.value[grouName]) {
      console.log('groupName already exists');

      // TODO: emit error message
      return;
    }

    dashboardGroups.value[grouName] = {
      name: grouName,
      deviceGroup: {},
    };
  }

  function renameDashboardGroup(groupName: string, newName: string) {
    if (!groupName || !newName) {
      console.log('groupName or newName is empty');
      return;
    }
    if (dashboardGroups.value[newName]) {
      console.log('newName already exists');
      return;
    }
    const group = dashboardGroups.value[groupName];
    delete dashboardGroups.value[groupName];

    group.name = newName;
    dashboardGroups.value[newName] = group;
    store.dispatch('hub/saveDashboardGroup', dashboardGroups.value[newName] as DashboardGroup);
  }

  function deleteDeviceExpose(groupName: string, deviceId: string, exposeName: string) {
    if (!groupName || !deviceId || !exposeName) {
      console.log('groupName or deviceId or exposeName is empty');
      return;
    }
    const deviceGroupExposes = dashboardGroups.value[groupName].deviceGroup[deviceId].exposes;
    const idx = deviceGroupExposes.indexOf(exposeName);
    if (idx === -1) {
      console.log('exposeName not found');
      return;
    }
    deviceGroupExposes.splice(idx, 1);
    store.dispatch('hub/saveDashboardGroup', dashboardGroups.value[groupName] as DashboardGroup);
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
  <div v-if="editMode" class="toolbar">
    <button class="toolbar-btn" @click="openNewDashboardGroupDialog">
      <i class="pi pi-plus" />
      <span>New Group</span>
    </button>
    <button class="toolbar-btn">
      <i class="pi pi-download" />
      <span>Export Groups</span> TODO
    </button>
  </div>

  <div class="dashboard-container">
    <div v-for="group in dashboardGroups" :key="group.name" class="dashboard-group">
      <div class="dashboard-title">{{ group.name }}</div>
      <div class="card-container" :class="{ 'edit-mode': editMode }">
        <div v-for="item in flattenDeviceGroup(group)" :key="`${item.deviceId}-${item.expose}`" class="card-item">
          <EntityCard
            :id="item.deviceId"
            :name="item.expose"
            compact
            :is-selected="editMode && selectedCard == getDeviceGroupId(item.deviceId, item.expose)"
            @selected="handleCardSelected"
            @delete="openDeleteDeviceExposeConfirmationDialog(group.name, $event.id, $event.name)" />
        </div>
        <div v-if="editMode" class="icon-tools">
          <span class="edit-icon pi pi-pen-to-square" @click="openRenameDashboardGroupDialog(group.name)" />
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
    gap: 0.2rem;
  }

  /* Deskop view */
  .dashboard-group {
    padding: 1rem 0.1rem;
  }

  /* Mobile view */
  @media (max-width: 768px) {
    .dashboard-group {
      padding: 1rem 0;
      flex: 1 1 200px;
      max-width: 100%;
      box-sizing: border-box;
    }
  }

  .dashboard-title {
    font-size: 0.95rem;
    font-weight: 500;
    padding: 8px;
    margin-bottom: 0.5rem;
    line-height: 1.4;
    color: #e0e0e0;
    letter-spacing: 0.25px;
  }

  .card-container {
    column-count: 2;
    column-gap: 0.5rem;
    max-width: 400px;
    border-radius: 12px;
    padding: 8px 10px 8px 10px;
    position: relative;
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

  .toolbar {
    display: flex;
    align-items: center;
    padding: 0.5rem;
    background: #1f1f1f;
    border-radius: 10px;
    margin-bottom: 1rem;
    gap: 0.5rem;
    width: fit-content;
    box-sizing: border-box;
  }

  .toolbar-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    background: #1f1f1f;
    color: #fff;
    padding: 6px 12px;
    border-radius: 8px;
    border: none;
    cursor: pointer;
    font-size: 0.85rem;
    transition: background 0.2s ease;
  }

  .toolbar-btn:hover {
    background: #3a3a3a;
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
