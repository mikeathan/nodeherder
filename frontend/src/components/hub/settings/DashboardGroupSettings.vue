<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '../../../store/index';
  import { DashboardGroups, DashboardGroup, DeviceGroup } from '@/types/settings.type';
  import ExposeSelectionDialog from '../../dialogs/ExposeSelectionDialog.vue';

  import Panel from 'primevue/panel';
  import { Device } from '@/types/device';
import { ChartComponents } from '@/mixins/useChartComponents';

  const dashboardGroups = computed(() => {
    return store.getters['hub/dashboardGroups']() as DashboardGroups;
  });

  const showSelectExposeDialog = ref(false);
  const dialogDeviceGroup = ref<DeviceGroup>({} as DeviceGroup);
  function deviceNameFromId(id: string): string {
    const device = store.getters['hub/findDevice'](id) as Device;
    if (device == undefined) {
      return '';
    }
    return device.friendly_name;
  }

  const getDeviceGroupArray = (group: DashboardGroup) => {
    return Object.values(group.deviceGroup);
  };

  const getDashboardGroups = computed(() => {
    return Object.values(dashboardGroups.value);
  });

  function addDeviceExpose( expose: string) {
    if (!expose) {
      return;
    }

    will need to split it into multiple ChartComponents
    dahsboard group which we have 
    then device group so it will sned the right events 
    //eed to add it in the devicegroup 
   // dashboardGroups.value[dialogDeviceGroupId.value].deviceGroup[dialogDeviceGroupId.value].exposes.push(expose);
    // emit update store
  }

  function openExposeDialog(dashboardroup: DashboardGroup,deviceGroup: DeviceGroup) {
    dialogDeviceGroup.value = deviceGroup;
    // Show the dialog
    showSelectExposeDialog.value = true;
  }

  const deleteExpose = (group: DeviceGroup, expose: string) => {
    group.exposes = group.exposes.filter((e: string) => e != expose);

    // emit update store
  };
  const deleteDeviceGroup = (group: DashboardGroup, deviceGroup: DeviceGroup) => {
    delete group.deviceGroup[deviceGroup.deviceId];

    // emit update store
  };
</script>
<style scoped></style>

<template>
  <h3>Dashboard Groups</h3>

  <div class="p-4">
    <Panel v-for="(group, index) in getDashboardGroups" :key="index" :header="group.name" toggleable :collapsed="true">
      <DataView :value="getDeviceGroupArray(group)" data-key="deviceId">
        <template #list="slotProps">
          <div v-for="(item, index) in slotProps.items" :key="index" class="col-12">
            <div class="flex flex-wrap md:flex-nowrap gap-4 items-start">
              <!-- Device  -->
              <div class="w-full md:w-auto flex-shrink-0">
                <div class="text-color-secondary font-semibold">{{ deviceNameFromId(item.deviceId) }}</div>
              </div>

              <!-- Exposes -->
              <div class="w-full md:w-auto flex-grow-1">
                <div class="flex flex-wrap gap-1 **justify-content-start**">
                  <Tag
                    v-for="(prop, propIndex) in item.exposes"
                    :key="prop"
                    severity="info"
                    class="flex align-items-center gap-1 pr-2">
                    <span>{{ prop }}</span>
                    <i class="pi pi-times-circle text-sm cursor-pointer" @click="deleteExpose(item, prop)"/>
                  </Tag>
                </div>
              </div>

              <!-- Button Actions -->
              <div class="flex flex-wrap gap-1 justify-content-center">
                <div class="flex">
                  <Button
                    icon="pi pi-plus"
                    class="p-button-text p-button-rounded mr-2"
                    aria-label="Add"
                    @click="openExposeDialog(group, item)" />
                  <Button
                    icon="pi pi-trash"
                    class="p-button-text p-button-rounded p-button-danger"
                    @click="deleteDeviceGroup(group, item)"
                    aria-label="Delete" />
                </div>
              </div>
            </div>
          </div>
        </template>
      </DataView>
    </Panel>
  </div>
  <ExposeSelectionDialog
    :id="dialogDeviceGroup.deviceId"
    :show="showSelectExposeDialog"
    @update="addDeviceExpose"
    @close="showSelectExposeDialog = false" />
  <!-- <div class="grid">
    <div v-for="group in Object.values(dashboardGroups)" :key="group.name" class="col-12 mb-3">
      <div class="surface-card border-round shadow-1">
        <div class="flex justify-content-between align-items-center cursor-pointer" @click="toggleGroup(group.name)">
          <div class="font-medium">{{ group.name }}</div>
          <div class="flex">
            <Button
              icon="pi pi-pencil"
              class="p-button-text p-button-rounded mr-2"
              @click.stop="openEditGroupDialog(group)" />
            <Button
              icon="pi pi-trash"
              class="p-button-text p-button-rounded p-button-danger mr-2"
              @click.stop="deleteGroup(group)" />
            <Button
              :icon="expandedGroups[group.name] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
              class="p-button-text p-button-rounded" />
          </div>
        </div>

        <div v-if="expandedGroups[group.name]" class="pt-3 border-top-1 surface-border">
          <div v-if="getDeviceGroupArray(group).length > 0">
            <div v-for="device in getDeviceGroupArray(group)" :key="device.deviceId" class="mb-3">
              <div
                class="flex justify-content-between align-items-center cursor-pointer"
                @click="toggleDevice(device.deviceId)">
                <div class="font-medium">{{ device.deviceId }}</div>
                <div class="flex">
                  <Button
                    icon="pi pi-pencil"
                    class="p-button-text p-button-rounded mr-2"
                    @click.stop="openEditDeviceDialog(group, device)" />
                  <Button
                    icon="pi pi-trash"
                    class="p-button-text p-button-rounded p-button-danger mr-2"
                    @click.stop="deleteDevice(group, device.deviceId)" />
                  <Button
                    :icon="expandedDevices[device.deviceId] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
                    class="p-button-text p-button-rounded" />
                </div>

                <div v-if="expandedDevices[device.deviceId]" class="">
                  <div v-if="device.exposes.length > 0" class="flex flex-wrap gap-2 mb-3">
                    <Chip
                      v-for="expose in device.exposes"
                      :key="expose"
                      :label="expose"
                      removable
                      class="bg-primary-100 text-primary-700"
                      @remove="deleteExpose(group, device.deviceId, expose)" />
                  </div>
                  <div v-else class="text-500 text-sm mb-3">No exposes added</div>
                  <Button
                    label="Add Expose"
                    icon="pi pi-plus"
                    class="p-button-text p-button-sm"
                    @click="openNewExposeDialog(group, device)" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <Dialog
    v-model:visible="displayGroupDialog"
    :style="{ width: '90%', maxWidth: '400px' }"
    :header="editMode === 'add' ? 'Add Group' : 'Edit Group'"
    :modal="true"
    class="p-fluid">
    <div class="field">
      <label for="groupName">Group Name</label>
      <InputText id="groupName" v-model="formState.groupName" required autofocus class="w-full" />
    </div>
    <template #footer>
      <Button label="Cancel" icon="pi pi-times" text @click="displayGroupDialog = false" />
      <Button label="Save" icon="pi pi-check" @click="saveGroup" />
    </template>
  </Dialog>

  <Dialog
    v-model:visible="displayDeviceDialog"
    :style="{ width: '90%', maxWidth: '400px' }"
    :header="editMode === 'add' ? 'Add Device' : 'Edit Device'"
    :modal="true"
    class="p-fluid">
    <div class="field">
      <label for="deviceId">Device ID</label>
      <InputText id="deviceId" v-model="formState.deviceId" required autofocus class="w-full" />
    </div>
    <template #footer>
      <Button label="Cancel" icon="pi pi-times" text @click="displayDeviceDialog = false" />
      <Button label="Save" icon="pi pi-check" @click="saveDevice" />
    </template>
  </Dialog>

  <Dialog
    v-model:visible="displayExposeDialog"
    :style="{ width: '90%', maxWidth: '400px' }"
    header="Add Expose"
    :modal="true"
    class="p-fluid">
    <div class="field">
      <label for="expose">Expose</label>
      <InputText id="expose" v-model="formState.expose" required autofocus class="w-full" />
    </div>
    <template #footer>
      <Button label="Cancel" icon="pi pi-times" text @click="displayExposeDialog = false" />
      <Button label="Add" icon="pi pi-check" @click="saveExpose" />
    </template>
  </Dialog>-->
</template>
