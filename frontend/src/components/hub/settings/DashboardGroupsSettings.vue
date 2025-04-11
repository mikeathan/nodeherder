<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '../../../store/index';
  import { DashboardGroups, DashboardGroup } from '@/types/settings.type';
  import ExposeSelectionDialog from '../../dialogs/ExposeSelectionDialog.vue';
  import DashboardGroupComponent from './DashboardGroup.vue';
  import ConfirmDialog from '../../dialogs/ConfirmDialog.vue';
  import InputDialog from '../../dialogs/InputDialog.vue';

  import Panel from 'primevue/panel';
  import {
    DialogEventAction,
    DialogEventActions,
    emitCloseDialog,
    emitOpenDialog,
    OpenDialogEvent,
  } from '@/mixins/useDialogsEventBus';

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
    const events: DialogEventActions = {
      close: () => emitCloseDialog(),
      confirm: (args: any) => {
        delete dashboardGroups.value[groupName];
        store.dispatch('hub/deleteDashboardGroup', groupName);
      },
    };
    const event: OpenDialogEvent = {
      type: 'confirm',
      props: {
        title: 'Question',
        message: 'Are you sure?',
        show: true,
      },
      events: events,
    };
    emitOpenDialog(event);
  }

  function openCreateNewDashboardGroupDialog() {
    const events: DialogEventActions = {
      close: () => emitCloseDialog(),
      confirm: (args: any) => {
        dashboardGroups.value[args] = {
          name: args,
          deviceGroup: {},
        };
      },
    };
    const event: OpenDialogEvent = {
      type: 'input',
      props: {
        title: 'Create',
        message: 'Create New Dashboard Group',
        show: true,
      },
      events: events,
    };
    emitOpenDialog(event);
  }

  function openExposeDialog(dashboardroup: DashboardGroup, deviceId: string) {
    const events: DialogEventActions = {
      close: () => emitCloseDialog(),
      update: (args: any) => {
        if (args) {
          dashboardroup.deviceGroup[deviceId].exposes.push(args);
          store.dispatch('hub/saveDashboardGroup', dashboardroup as DashboardGroup);
        }
      },
    };
    const event: OpenDialogEvent = {
      type: 'exposeSelection',
      props: {
        title: 'Select',
        message: 'Select Expose',
        id: deviceId,
        show: true,
      },
      events: events,
    };
    emitOpenDialog(event);
  }

  const updateDeviceGroup = (group: DashboardGroup) => {
    dashboardGroups.value[group.name] = group;

    store.dispatch('hub/saveDashboardGroup', group);
  };

  const createNewDashboardGroup = (groupName: string) => {
    dashboardGroups.value[groupName] = {
      name: groupName,
      deviceGroup: {},
    };
  };
</script>
<style scoped></style>

<template>
  <h3>Dashboard Groups</h3>

  <div class="flex justify-end mb-4 mt-4">
    <Button
      icon="pi pi-plus"
      text
      label="Add Group"
      class="p-button-sm p-button-outlined"
      @click="openCreateNewDashboardGroupDialog" />
  </div>

  <Panel v-for="(group, index) in getDashboardGroups" :key="index" toggleable :collapsed="true">
    <template #header>
      <div class="flex items-center w-full">
        we need new button to add new device group

        <span class="flex items-center cursor-pointer">
          <Button
            icon="pi pi-trash"
            class="p-button-text p-button-rounded p-button-danger pb-5"
            aria-label="Delete"
            @click="openDeleteDeviceGroupConfirmationDialog(group.name)" />
          {{ group.name }}
        </span>
      </div>
    </template>
    <DashboardGroupComponent
      :dashboardGroup="group"
      @update="(g) => updateDeviceGroup(g)"
      @insert="(id) => openExposeDialog(group, id)" />
  </Panel>
  <!-- <ExposeSelectionDialog
    :id="dialogDeviceGroupId"
    :show="showSelectExposeDialog"
    @update="addDeviceExpose"
    @close="closeExposeDialog" />
  <ConfirmDialog
    :show="showConfirmDialog"
    @confirm="deleteDeviceGroup(dialogDeviceGroupName)"
    @close="showConfirmDialog = false" />

  <InputDialog
    :show="showInputDialog"
    title="New Group"
    message="Enter the dashboard group name"
    @confirm="createNewDashboardGroup"
    @close="showInputDialog = false" /> -->
</template>
