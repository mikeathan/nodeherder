<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '../../../store/index';
  import { DashboardGroups, DashboardGroup } from '@/types/settings.type';
  import ExposeSelectionDialog from '../../dialogs/ExposeSelectionDialog.vue';
  import DashboardGroupComponent from './DashboardGroup.vue';
  import ConfirmDialog from '../../dialogs/ConfirmDialog.vue';
  import InputDialog from '../../dialogs/InputDialog.vue';

  import Panel from 'primevue/panel';

  const dashboardGroups = computed(() => {
    return store.getters['hub/dashboardGroups']() as DashboardGroups;
  });

  // TODO
  // temporary: will need to be refactored so dialog is placed in app
  // and controlled via eventbus events
  const showSelectExposeDialog = ref(false);
  const showConfirmDialog = ref(false);
  const showInputDialog = ref(false);

  const dialogDeviceGroupId = ref<string>('');
  const dialogDashboardGroup = ref<DashboardGroup | null>(null);
  const dialogDeviceGroupName = ref<string>('');
  // TODO: refactor to use event bus

  const getDashboardGroups = computed(() => {
    if (!dashboardGroups.value) {
      return [];
    }
    return Object.values(dashboardGroups.value);
  });

  function openExposeDialog(dashboardroup: DashboardGroup, deviceId: string) {
    dialogDeviceGroupId.value = deviceId;
    dialogDashboardGroup.value = dashboardroup;

    // Show the dialog
    showSelectExposeDialog.value = true;
  }

  function closeExposeDialog() {
    showSelectExposeDialog.value = false;
    dialogDeviceGroupId.value = '';
    dialogDashboardGroup.value = null;
  }

  function openDeleteDeviceGroupConfirmationDialog(groupName: string) {
    showConfirmDialog.value = true;
    dialogDeviceGroupName.value = groupName;
  }

  function openCreateNewDashboardGroupDialog() {
    showInputDialog.value = true;
  }

  const addDeviceExpose = (expose: string) => {
    if (!expose || !dialogDashboardGroup.value || !dialogDeviceGroupId.value) {
      return;
    }

    dialogDashboardGroup.value.deviceGroup[dialogDeviceGroupId.value].exposes.push(expose);
    store.dispatch('hub/saveDashboardGroup', dialogDashboardGroup.value as DashboardGroup);
  };

  const deleteDeviceGroup = (groupName: string) => {
    delete dashboardGroups.value[groupName];
    store.dispatch('hub/deleteDashboardGroup', groupName);
  };

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
  <ExposeSelectionDialog
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
    @close="showInputDialog = false" />
</template>
