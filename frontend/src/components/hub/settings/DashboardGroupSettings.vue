<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '../../../store/index';
  import { DashboardGroups, DashboardGroup, DeviceGroup } from '@/types/settings.type';
  import ExposeSelectionDialog from '../../dialogs/ExposeSelectionDialog.vue';
  import DashboardGroupComponent from './DashboardGroup.vue';

  import Panel from 'primevue/panel';

  const dashboardGroups = computed(() => {
    return store.getters['hub/dashboardGroups']() as DashboardGroups;
  });

  const showSelectExposeDialog = ref(false);
  const dialogDeviceGroupId = ref<string>('');
  

  const getDashboardGroups = computed(() => {
    return Object.values(dashboardGroups.value);
  });

  function addDeviceExpose(expose: string) {
    if (!expose) {
      return;
    }

    //eed to add it in the devicegroup
    // dashboardGroups.value[dialogDeviceGroupId.value].deviceGroup[dialogDeviceGroupId.value].exposes.push(expose);
    // emit update store
  }

  function openExposeDialog(dashboardroup: DashboardGroup, deviceId: string) {

    maybe we need some context her to store the info required
    
    dialogDeviceGroupId.value = deviceId;
    // Show the dialog
    showSelectExposeDialog.value = true;
  }

  const deleteExpose = (group: DeviceGroup, expose: string) => {
    group.exposes = group.exposes.filter((e: string) => e != expose);

    // emit update store
  };

  const updateDeviceGroup = (group: DashboardGroup) => {
    dashboardGroups.value[group.name] = group;

    // emit update store
  };

</script>
<style scoped></style>

<template>
  <h3>Dashboard Groups</h3>

  <div class="p-4">
    <Panel v-for="(group, index) in getDashboardGroups" :key="index" :header="group.name" toggleable :collapsed="true">
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
    @close="showSelectExposeDialog = false" />
 
</template>
