<script setup lang="ts">
import { computed } from 'vue';
import { store } from '../../../store/index';
import { DashboardGroups, DashboardGroup } from '@/types/settings.type';
import DashboardGroupComponent from './DashboardGroup.vue';

import Panel from 'primevue/panel';

import { emitOpenConfirmationDialog, emitOpenExposeSelectionDialog, emitOpenInputDialogEvent } from '@/contracts/dialog-events';

const dashboardGroups = computed(() => {
  return store.getters['hub/dashboardGroups']() as DashboardGroups;
});

const getDashboardGroups = computed(() => {
  if (!dashboardGroups.value) {
    return [];
  }
  return Object.values(dashboardGroups.value);
});

function openDeleteDeviceGroupConfirmationDialog(groupName: string) {
  const props = {
    title: 'Question',
    message: 'Are you sure?'
  }
  emitOpenConfirmationDialog(() => deleteDeviceGroup(groupName), props)
}

function openCreateNewDashboardGroupDialog() {
  const props = {
    title: 'Create New Dashboard Group',
    message: 'Enter Group Name',
  }
  emitOpenInputDialogEvent(createNewDeviceGroup, props);
}


function openAddDeviceExposeDialog(dashboardroup: DashboardGroup, deviceId: string) {
  const props = {
    id: deviceId,
    title: 'Select Expose',
    message: 'Select Expose',
  }
  emitOpenExposeSelectionDialog((args) => addNewDeviceExpose(dashboardroup, deviceId, args), props);
}


function deleteDeviceGroup(groupName: string) {
  if (!groupName) {
    console.log('groupName is empty');
    return;
  }

  delete dashboardGroups.value[groupName];
  store.dispatch('hub/deleteDashboardGroup', groupName);
}

function createNewDeviceGroup(groupName: string) {
  if (!groupName) {
    console.log('groupName is empty');
    return;
  }

  dashboardGroups.value[groupName] = {
    name: groupName,
    deviceGroup: {},
  };
}

function addNewDeviceExpose(dashboardroup: DashboardGroup, deviceId: string, exposeName: string) {
  if (!exposeName) {
    console.log('exposeName is empty');
    return;
  }
  dashboardroup.deviceGroup[deviceId].exposes.push(exposeName);
  store.dispatch('hub/saveDashboardGroup', dashboardroup as DashboardGroup);
}

const updateDeviceGroup = (group: DashboardGroup) => {
  dashboardGroups.value[group.name] = group;

  store.dispatch('hub/saveDashboardGroup', group);
};

TODO
we need new button to add new device group
and update dialogs in other components
</script>
<style scoped></style>

<template>
  <h3>Dashboard Groups</h3>

  <div class="flex justify-end mb-4 mt-4">
    <Button icon="pi pi-plus" text label="Add Group" class="p-button-sm p-button-outlined"
      @click="openCreateNewDashboardGroupDialog" />
  </div>

  <Panel v-for="(group, index) in getDashboardGroups" :key="index" toggleable :collapsed="true">
    <template #header>
      <div class="flex items-center w-full">

        <span class="flex items-center cursor-pointer">
          <Button icon="pi pi-trash" class="p-button-text p-button-rounded p-button-danger pb-5" aria-label="Delete"
            @click="openDeleteDeviceGroupConfirmationDialog(group.name)" />
          {{ group.name }}
        </span>
      </div>
    </template>
    <DashboardGroupComponent :dashboardGroup="group" @update="(g) => updateDeviceGroup(g)"
      @insert="(id) => openAddDeviceExposeDialog(group, id)" />
  </Panel>

</template>
