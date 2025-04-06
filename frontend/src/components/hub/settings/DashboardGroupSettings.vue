<script setup lang="ts">
import { computed, ref } from 'vue';
import { store } from '../../../store/index';
import { DashboardGroups, DashboardGroup, DeviceGroup } from '@/types/settings.type';

const dashboardGroups = computed(() => {
  return store.getters['hub/dashboardGroups']() as DashboardGroups;
});

// Form states
const displayGroupDialog = ref(false);
const displayDeviceDialog = ref(false);
const displayExposeDialog = ref(false);
const selectedGroup = ref<DashboardGroup | null>(null);
const selectedDeviceGroup = ref<DeviceGroup | null>(null);
const formState = ref<{
  groupName: string;
  deviceId: string;
  expose: string;
}>({
  groupName: '',
  deviceId: '',
  expose: '',
});

const editMode = ref<'add' | 'edit'>('add');
const expandedGroups = ref<Record<string, boolean>>({});
const expandedDevices = ref<Record<string, boolean>>({});

// Group operations
const openNewGroupDialog = () => {
  formState.value.groupName = '';
  editMode.value = 'add';
  displayGroupDialog.value = true;
};

const openEditGroupDialog = (group: DashboardGroup) => {
  selectedGroup.value = group;
  formState.value.groupName = group.name;
  editMode.value = 'edit';
  displayGroupDialog.value = true;
};

const saveGroup = () => {
  if (editMode.value === 'add') {
    store.dispatch('hub/addDashboardGroup', {
      name: formState.value.groupName,
      deviceGroup: {},
    });
  } else {
    if (selectedGroup.value) {
      store.dispatch('hub/updateDashboardGroup', {
        oldName: selectedGroup.value.name,
        newGroup: {
          name: formState.value.groupName,
          deviceGroup: selectedGroup.value.deviceGroup,
        },
      });
    }
  }
  displayGroupDialog.value = false;
};

const deleteGroup = (group: DashboardGroup) => {
  store.dispatch('hub/deleteDashboardGroup', group.name);
};

// Device operations
const openNewDeviceDialog = (group: DashboardGroup) => {
  selectedGroup.value = group;
  formState.value.deviceId = '';
  editMode.value = 'add';
  displayDeviceDialog.value = true;
};

const openEditDeviceDialog = (group: DashboardGroup, device: DeviceGroup) => {
  selectedGroup.value = group;
  selectedDeviceGroup.value = device;
  formState.value.deviceId = device.deviceId;
  editMode.value = 'edit';
  displayDeviceDialog.value = true;
};

const saveDevice = () => {
  if (!selectedGroup.value) return;

  if (editMode.value === 'add') {
    store.dispatch('hub/addDeviceToGroup', {
      groupName: selectedGroup.value.name,
      device: {
        deviceId: formState.value.deviceId,
        exposes: [],
      },
    });
  } else {
    if (selectedDeviceGroup.value) {
      store.dispatch('hub/updateDeviceInGroup', {
        groupName: selectedGroup.value.name,
        oldDeviceId: selectedDeviceGroup.value.deviceId,
        newDevice: {
          deviceId: formState.value.deviceId,
          exposes: selectedDeviceGroup.value.exposes,
        },
      });
    }
  }
  displayDeviceDialog.value = false;
};

const deleteDevice = (group: DashboardGroup, deviceId: string) => {
  store.dispatch('hub/removeDeviceFromGroup', {
    groupName: group.name,
    deviceId: deviceId,
  });
};

// Expose operations
const openNewExposeDialog = (group: DashboardGroup, device: DeviceGroup) => {
  selectedGroup.value = group;
  selectedDeviceGroup.value = device;
  formState.value.expose = '';
  displayExposeDialog.value = true;
};

const saveExpose = () => {
  if (!selectedGroup.value || !selectedDeviceGroup.value) return;

  store.dispatch('hub/addExposeToDevice', {
    groupName: selectedGroup.value.name,
    deviceId: selectedDeviceGroup.value.deviceId,
    expose: formState.value.expose,
  });

  displayExposeDialog.value = false;
};

const deleteExpose = (group: DashboardGroup, deviceId: string, expose: string) => {
  store.dispatch('hub/removeExposeFromDevice', {
    groupName: group.name,
    deviceId: deviceId,
    expose: expose,
  });
};

// Toggle expand/collapse
const toggleGroup = (groupName: string) => {
  expandedGroups.value[groupName] = !expandedGroups.value[groupName];
};

const toggleDevice = (deviceId: string) => {
  expandedDevices.value[deviceId] = !expandedDevices.value[deviceId];
};

// Helper function to get device group array for a dashboard group
const getDeviceGroupArray = (group: DashboardGroup) => {
  return Object.values(group.deviceGroup);
};
</script>

<template>
  <h3>Dashboard Groups</h3>
  <Accordion multiple>
    <AccordionPanel v-for="(group, index) in Object.values(dashboardGroups)" :key="group.name" :value="String(index)">
      <AccordionHeader >
        <div class="w-full flex items-center py-1 min-h-0 text-sm">
          <div class="font-medium truncate">
            {{ group.name }}
          </div>

          <div class="p-4 ml-auto flex gap-2">
            <Button icon="pi pi-pencil" class="p-button-text p-button-rounded h-6"
              @click.stop="openEditGroupDialog(group)" />
            <Button icon="pi pi-trash" class="p-button-text p-button-rounded p-button-danger h-6"
              @click.stop="deleteGroup(group)" />
          </div>
        </div>
      </AccordionHeader>
      <AccordionContent class="p-3 border-top-1">

        <Accordion multiple>
          <AccordionPanel v-for="(deviceGroup, dgIndex) in group.deviceGroup" :key="deviceGroup.deviceId"
            :value="String(dgIndex)">
            <AccordionHeader>
              {{ deviceGroup.deviceId }}
            </AccordionHeader>

            <AccordionContent>
              <Chip v-for="expose in deviceGroup.exposes" :key="expose" :label="expose" removable
                class="bg-primary-100 text-primary-700" @remove="deleteExpose(group, deviceGroup.deviceId, expose)" />
            </AccordionContent>
          </AccordionPanel>
        </Accordion>

      </AccordionContent>
    </AccordionPanel>
  </Accordion>

  ----------------------------------------------------------
  TEST
  ----------------------------------------------------------
  <div class="dashboard-editor1">
    <div class="surface-card p-4 shadow-2 border-round">
      <div class="flex justify-content-between align-items-center mb-4">
        <h2 class="text-xl font-medium m-0">Dashboard Groups</h2>
        <Button icon="pi pi-plus" label="Add Group" class="p-button-outlined" @click="openNewGroupDialog" />
      </div>

      <div class="grid">
        <div v-for="group in Object.values(dashboardGroups)" :key="group.name" class="col-12 mb-3">

          <div class="surface-card border-round shadow-1">
            <div class="p-3 flex justify-content-between align-items-center cursor-pointer"
              @click="toggleGroup(group.name)">
              <div class="font-medium">{{ group.name }}</div>
              <div class="flex">
                <Button icon="pi pi-pencil" class="p-button-text p-button-rounded mr-2"
                  @click.stop="openEditGroupDialog(group)" />
                <Button icon="pi pi-trash" class="p-button-text p-button-rounded p-button-danger mr-2"
                  @click.stop="deleteGroup(group)" />
                <Button :icon="expandedGroups[group.name] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
                  class="p-button-text p-button-rounded" />
              </div>
            </div>

            <div v-if="expandedGroups[group.name]" class="p-3 border-top-1 surface-border">
              <div v-if="getDeviceGroupArray(group).length > 0">
                <div v-for="device in getDeviceGroupArray(group)" :key="device.deviceId" class="mb-3">
                  <div class="surface-100 border-round p-3">
                    <div class="flex justify-content-between align-items-center cursor-pointer"
                      @click="toggleDevice(device.deviceId)">
                      <div class="font-medium">{{ device.deviceId }}</div>
                      <div class="flex">
                        <Button icon="pi pi-pencil" class="p-button-text p-button-rounded mr-2"
                          @click.stop="openEditDeviceDialog(group, device)" />
                        <Button icon="pi pi-trash" class="p-button-text p-button-rounded p-button-danger mr-2"
                          @click.stop="deleteDevice(group, device.deviceId)" />
                        <Button :icon="expandedDevices[device.deviceId] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
                          class="p-button-text p-button-rounded" />
                      </div>
                    </div>

                    <div v-if="expandedDevices[device.deviceId]" class="mt-3">
                      <div v-if="device.exposes.length > 0" class="flex flex-wrap gap-2 mb-3">
                        <Chip v-for="expose in device.exposes" :key="expose" :label="expose" removable
                          class="bg-primary-100 text-primary-700"
                          @remove="deleteExpose(group, device.deviceId, expose)" />
                      </div>
                      <div v-else class="text-500 text-sm mb-3">No exposes added</div>
                      <Button label="Add Expose" icon="pi pi-plus" class="p-button-text p-button-sm"
                        @click="openNewExposeDialog(group, device)" />
                    </div>
                  </div>
                </div>
              </div>
              <div v-else class="text-500 p-3 text-center">No devices in this group</div>

              <div class="flex justify-content-center mt-3">
                <Button label="Add Device" icon="pi pi-plus" class="p-button-outlined p-button-sm"
                  @click="openNewDeviceDialog(group)" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="Object.values(dashboardGroups).length === 0" class="text-500 text-center p-5">
        No dashboard groups found. Create your first group!
      </div>
    </div>

    <Dialog v-model:visible="displayGroupDialog" :style="{ width: '90%', maxWidth: '400px' }"
      :header="editMode === 'add' ? 'Add Group' : 'Edit Group'" :modal="true" class="p-fluid">
      <div class="field">
        <label for="groupName">Group Name</label>
        <InputText id="groupName" v-model="formState.groupName" required autofocus class="w-full" />
      </div>
      <template #footer>
        <Button label="Cancel" icon="pi pi-times" text @click="displayGroupDialog = false" />
        <Button label="Save" icon="pi pi-check" @click="saveGroup" />
      </template>
    </Dialog>

    <Dialog v-model:visible="displayDeviceDialog" :style="{ width: '90%', maxWidth: '400px' }"
      :header="editMode === 'add' ? 'Add Device' : 'Edit Device'" :modal="true" class="p-fluid">
      <div class="field">
        <label for="deviceId">Device ID</label>
        <InputText id="deviceId" v-model="formState.deviceId" required autofocus class="w-full" />
      </div>
      <template #footer>
        <Button label="Cancel" icon="pi pi-times" text @click="displayDeviceDialog = false" />
        <Button label="Save" icon="pi pi-check" @click="saveDevice" />
      </template>
    </Dialog>

    <Dialog v-model:visible="displayExposeDialog" :style="{ width: '90%', maxWidth: '400px' }" header="Add Expose"
      :modal="true" class="p-fluid">
      <div class="field">
        <label for="expose">Expose</label>
        <InputText id="expose" v-model="formState.expose" required autofocus class="w-full" />
      </div>
      <template #footer>
        <Button label="Cancel" icon="pi pi-times" text @click="displayExposeDialog = false" />
        <Button label="Add" icon="pi pi-check" @click="saveExpose" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.dashboard-editor .p-button {
  border-radius: 20px;
}

.dashboard-editor .p-button-text {

  padding: 0.5rem;
}

@media screen and (max-width: 576px) {
  .dashboard-editor .p-button-label {
    display: none;
  }

  .dashboard-editor .p-button-icon {
    margin-right: 0;
  }

  .dashboard-editor h2 {
    font-size: 1.25rem;
  }
}
</style>
