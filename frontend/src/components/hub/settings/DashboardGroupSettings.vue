<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '../../../store/index';
  import { DashboardGroups, DashboardGroup, DeviceGroup } from '@/types/settings.type';
  import Panel from 'primevue/panel';

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
<style scoped></style>
<template>
  <h3>Dashboard Groups</h3>

  <div class="p-4"> 
    <Card>
      <template #title>
        groupName
      </template>
      <template #content>
        <DataView :value="Object.values(dashboardGroups)" layout="grid" data-key="id">
          <template #grid="slotProps">
            <div class="col-12 md:col-6 lg:col-4 xl:col-3 p-2"> 
              <div class="p-4 border-1 surface-border border-round surface-card h-full flex flex-column"> 
                <div class="mb-3">
                  <strong class="block mb-1">Device ID:</strong>
                  <span class="text-color-secondary">{{ slotProps.items.deviceId }}</span>
                </div>
                <div>
                  <strong class="block mb-2">Exposed Properties:</strong>
                  <div class="flex flex-wrap gap-1"> 
                    <Tag v-for="prop in slotProps.items.expose"
                         :key="prop"
                         :value="prop"
                         severity="info"
                         class="mr-1 mb-1"> 
                    </Tag>
                  </div>
                </div>
              </div>
            </div>
          </template>

          <template #empty>
            <div class="p-4 text-center">No devices found in this group.</div>
          </template>
        </DataView>
      </template>
    </Card>
  </div>


  <!-- <div v-for="group in Object.values(dashboardGroups)" :key="group.name" class="col-12 mb-3">
    <Panel toggleable>
      <template #header>
        <div class="flex items-center gap-2">
          <Avatar image="https://primefaces.org/cdn/primevue/images/avatar/amyelsner.png" shape="circle" />
          <span class="font-bold">{{ group.name }}</span>
        </div>
      </template>
      <template #footer>
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div class="flex items-center gap-2">
            <Button icon="pi pi-user" rounded text></Button>
            <Button icon="pi pi-bookmark" severity="secondary" rounded text></Button>
          </div>
          <span class="text-surface-500 dark:text-surface-400">Updated 2 hours ago</span>
        </div>
      </template>

      <DataTable :value="Object.values(group.deviceGroup)" tableStyle="min-width: 50rem">
        <Column field="deviceId" header="deviceId"></Column>
        <Column field="exposes" header="exposes">
          <template #body="slotProps">
            <Chip
              v-for="expose in slotProps.data.exposes"
              :key="expose"
              :label="expose"
              removable
              class="bg-primary-100 text-primary-700"
              />
          </template>
        </Column>
      </DataTable>
      <div v-for="deviceGroup in group.deviceGroup" :key="deviceGroup.deviceId" class="col-12 mb-3">
        {{ deviceGroup.deviceId }} 
          <Chip
            v-for="expose in deviceGroup.exposes"
            :key="expose"
            :label="expose"
            removable
            class="bg-primary-100 text-primary-700"
            @remove="deleteExpose(group, deviceGroup.deviceId, expose)" />
      </div>
    </Panel>
  </div> -->
  <!-- <DataView :value="groupArray" layout="list" data-key="id">
    <template #list="slotProps">
      <div v-for="(item, index) in slotProps.items" :key="index">
 
        <div class="flex justify-content-between align-items-center cursor-pointer">
          <div class="font-medium">{{ item.name }}</div>
          <div class="flex">
            <Button
              icon="pi pi-pencil"
              class="p-button-text p-button-rounded mr-2"
              @click.stop="openEditGroupDialog(item)" />
            <Button
              icon="pi pi-trash"
              class="p-button-text p-button-rounded p-button-danger mr-2"
              @click.stop="deleteGroup(item)" />
            <Button
              :icon="expandedGroups[item.name] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
              class="p-button-text p-button-rounded" />
          </div>
        </div>
      </div>
    </template>
  </DataView> -->

  ----------------------------------------------------------
  <div class="grid">
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
  </Dialog>
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
