<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '../../../store/index';
  import { DashboardGroups, DashboardGroup } from '@/types/settings.type';
  import ExposeSelectionDialog from '../../dialogs/ExposeSelectionDialog.vue';
  import DashboardGroupComponent from './DashboardGroup.vue';

  import Panel from 'primevue/panel';

  const dashboardGroups = computed(() => {
    return store.getters['hub/dashboardGroups']() as DashboardGroups;
  });

  // TODO
  // temporary: will need to be refactored so dialog is placed in app
  // and controlled via eventbus events
  const showSelectExposeDialog = ref(false);
  const showConfirmDialog = ref(false);

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
</script>
<style scoped></style>

<template>
  <h3>Dashboard Groups</h3>

  <div class="p-4">
    <Panel v-for="(group, index) in getDashboardGroups" :key="index" toggleable :collapsed="true">
      <template #header>
        <div class="flex items-center w-full">
          <span class="flex items-center cursor-pointer">
            <Button
              icon="pi pi-trash"
              class="p-button-text p-button-rounded p-button-danger pb-5"
              aria-label="Delete"
              @click="openDeleteDeviceGroupConfirmationDialog(group.name)" />
            <!-- <i class="pi pi-trash small  me-3 mt-1 cursor-pointer"  style="color: #e74c3c;font-size: 1rem" /> -->
            {{ group.name }}
          </span>
        </div>
      </template>
      <DashboardGroupComponent
        :dashboardGroup="group"
        @update="(g) => updateDeviceGroup(g)"
        @insert="(id) => openExposeDialog(group, id)" />
    </Panel>
  </div>
  <ExposeSelectionDialog
    :id="dialogDeviceGroupId"
    :show="showSelectExposeDialog"
    @update="addDeviceExpose"
    @close="closeExposeDialog" />
  <ConfirmDialog
    :show="showConfirmDialog"
    @confirm="deleteDeviceGroup(dialogDeviceGroupName)"
    @close="showConfirmDialog = false" />
</template>
