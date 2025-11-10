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

  const isEditMode = ref(false);
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

  function openImportDashboardGroupsConfirmationDialog() {
    const props = {
      title: 'Import Dashboard Groups',
      message: `Import new dashboard groups from file.\n\nAre you sure you want to continue?`,
    };
    emitOpenConfirmationDialog(importDashboardGroups, props);
  }

  function addNewDashboardGroup(grouName: string) {
    if (!grouName) {
      // TODO: emit error message
      alert('groupName is empty');
      return;
    }

    if (dashboardGroups.value[grouName]) {
      alert('groupName already exists');

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
      alert('groupName or newName is empty');
      return;
    }
    if (dashboardGroups.value[newName]) {
      alert('newName already exists');
      return;
    }

    // TODO: this is wrong, this should be done in the ws event response in case the request is not successful
    const group = dashboardGroups.value[groupName];
    delete dashboardGroups.value[groupName];

    group.name = newName;
    dashboardGroups.value[newName] = group;
    store.dispatch('hub/renameDashboardGroup', { oldName: groupName, newName: newName });
  }

  function deleteDeviceExpose(groupName: string, deviceId: string, exposeName: string) {
    if (!groupName || !deviceId || !exposeName) {
      alert('groupName or deviceId or exposeName is empty');
      return;
    }

    // TODO: this is wrong, this should be done in the ws event response in case the request is not successful
    const deviceGroupExposes = dashboardGroups.value[groupName].deviceGroup[deviceId].exposes;
    const idx = deviceGroupExposes.indexOf(exposeName);
    if (idx === -1) {
      alert('exposeName not found');
      return;
    }
    deviceGroupExposes.splice(idx, 1);
    store.dispatch('hub/saveDashboardGroup', dashboardGroups.value[groupName] as DashboardGroup);
  }

  function deleteDeviceGroup(groupName: string) {
    if (!groupName) {
      alert('groupName is empty');
      return;
    }

    // TODO: this is wrong, this should be done in the ws event response in case the request is not successful
    delete dashboardGroups.value[groupName];
    store.dispatch('hub/deleteDashboardGroup', groupName);
  }

  function exportDashboardGroups() {
    const dashboardGroupsJson = JSON.stringify({ dashboardGroups: dashboardGroups.value }, null, 2);

    const blob = new Blob([dashboardGroupsJson], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'dashboard-groups.json';
    link.click();
    URL.revokeObjectURL(url);
  }

  function importDashboardGroups() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    input.onchange = (event) => {
      const file = (event.target as HTMLInputElement).files?.[0];
      if (!file) {
        return;
      }
      const reader = new FileReader();
      reader.onload = (event) => {
        const dashboardGroupsJson = JSON.parse(event.target?.result as string);
        store.dispatch('hub/importDashboardGroups', dashboardGroupsJson.dashboardGroups);
      };
      reader.readAsText(file);
    };
    input.click();
  }

  function allowExport(): boolean {
    return Object.keys(dashboardGroups.value).length > 0;
  }

  function toggleEditMode() {
    isEditMode.value = !isEditMode.value;
  }
</script>

<template>
  <div>
    <Button icon="pi pi-cog" severity="secondary" size="small" @click="toggleEditMode" />
    <template v-if="isEditMode">
      <Button
        icon="pi pi-plus"
        severity="secondary"
        size="small"
        @click="openNewDashboardGroupDialog"
        label="New Group" />
      <Button
        icon="pi pi-download"
        severity="secondary"
        size="small"
        @click="exportDashboardGroups"
        :disabled="!allowExport()"
        label="Export" />
      <Button
        icon="pi pi-upload"
        severity="secondary"
        size="small"
        @click="openImportDashboardGroupsConfirmationDialog()"
        label="Import" />
    </template>
  </div>

  <div class="dashboard-container">
    <div v-for="group in dashboardGroups" :key="group.name" class="dashboard-group">
      <div class="dashboard-title">{{ group.name }}</div>
      <div class="card-container" :class="{ 'edit-mode': isEditMode }">
        <div v-for="item in flattenDeviceGroup(group)" :key="`${item.deviceId}-${item.expose}`" class="card-item">
          <EntityCard
            :id="item.deviceId"
            :name="item.expose"
            compact
            :is-selected="isEditMode && selectedCard == getDeviceGroupId(item.deviceId, item.expose)"
            @selected="handleCardSelected"
            @delete="openDeleteDeviceExposeConfirmationDialog(group.name, $event.id, $event.name)" />
        </div>
        <div v-if="isEditMode" class="icon-tools">
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
    padding-left: 12px;
    padding-bottom: 5px;
    line-height: 1.4;
    color: #e0e0e0;
    letter-spacing: 0.25px;
  }

  .card-container {
    column-count: 2;
    column-gap: 0.5rem;
    max-width: 400px;
    border-radius: 12px;
    padding: 5px 10px 5px 10px;
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
