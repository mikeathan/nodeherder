<script setup lang="ts">
import { computed } from 'vue';
import { store } from '../../../store/index';
import { DashboardGroups, DashboardGroup } from '@/types/settings.type';
import DashboardGroupComponent from './DashboardGroup.vue';

import Panel from 'primevue/panel';

import { emitOpenConfirmationDialog, emitOpenExposeSelectionDialog, emitOpenInputDialogEvent, emitOpenDeviceSelectionDialog } from '@/contracts/dialog-events';

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

function openAddDeviceGroupConfirmationDialog(groupName: string) {
  const props = {
    title: 'Input',
    message: 'Add device group'
  }
  emitOpenDeviceSelectionDialog((args) => createNewDeviceGroup(groupName, args), props)
}

function openCreateNewDashboardGroupDialog() {
  const props = {
    title: 'Create New Dashboard Group',
    message: 'Enter Group Name',
  }
  emitOpenInputDialogEvent(createNewDashboardGroup, props);
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

function createNewDeviceGroup(groupName: string, deviceId: string) {
  if (!deviceId) {
    console.log('deviceId is empty');
    return;
  }

  dashboardGroups.value[groupName].deviceGroup[deviceId] = {
    deviceId: deviceId,
    exposes: [],
  };
}

function createNewDashboardGroup(groupName: string) {
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
  const exists = dashboardroup.deviceGroup[deviceId].exposes.some(e => e == exposeName)
  if (exists) {
    console.log('expose ', exposeName, ' already exists');
    // TODO: show error message
    return;
  }
  dashboardroup.deviceGroup[deviceId].exposes.push(exposeName);
  store.dispatch('hub/saveDashboardGroup', dashboardroup as DashboardGroup);
}

const updateDeviceGroup = (group: DashboardGroup) => {
  dashboardGroups.value[group.name] = group;

  store.dispatch('hub/saveDashboardGroup', group);
};


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
        <span class="flex items-center cursor-pointer ">
          <span @click="openAddDeviceGroupConfirmationDialog(group.name)" class="mr-3 ">
            <i class="pi pi-plus-circle text-md" style="color: #f44336" />
          </span>
          <span @click="openDeleteDeviceGroupConfirmationDialog(group.name)" class="mr-3">
            <i class="pi pi-trash text-md" style="color: #f44336" />
          </span>
          {{ group.name }}
        </span>
      </div>
    </template>
    <DashboardGroupComponent :dashboardGroup="group" @update="(g) => updateDeviceGroup(g)"
      @insert="(id) => openAddDeviceExposeDialog(group, id)" />
  </Panel>

</template>
