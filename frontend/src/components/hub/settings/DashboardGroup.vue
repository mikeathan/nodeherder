<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '../../../store/index';
  import { DashboardGroup, DeviceGroup } from '@/types/settings.type';

  import { Device } from '@/types/device';

  const props = defineProps<{
    dashboardGroup: DashboardGroup;
  }>();

  const emit = defineEmits<{
    (e: 'update', dashboardGroup: DashboardGroup): void;
    (e: 'insert', deviceId: string): void;
  }>();

  const dashboardGroup = ref<DashboardGroup>(props.dashboardGroup);
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

  function insertExpose(deviceGroup: DeviceGroup) {
    emit('insert', deviceGroup.deviceId);
  }

  // TODO: if there is only one expose and user wants to delete it,
  // message to say that the devicegroup will be remove and then remove it
  const deleteExpose = (deviceGroup: DeviceGroup, expose: string) => {
    deviceGroup.exposes = deviceGroup.exposes.filter((e: string) => e != expose);
    dashboardGroup.value.deviceGroup[deviceGroup.deviceId] = deviceGroup;

    emit('update', { ...dashboardGroup.value });
  };

  const deleteDeviceGroup = (deviceGroup: DeviceGroup) => {
    delete dashboardGroup.value.deviceGroup[deviceGroup.deviceId];

    emit('update', { ...dashboardGroup.value });
  };
</script>

<template>
  <DataView :value="getDeviceGroupArray(props.dashboardGroup)" data-key="deviceId">
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
                <i class="pi pi-times-circle text-sm cursor-pointer" @click="deleteExpose(item, prop)" />
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
                @click="insertExpose(item)" />
              <Button
                icon="pi pi-trash"
                class="p-button-text p-button-rounded p-button-danger"
                @click="deleteDeviceGroup(item)"
                aria-label="Delete" />
            </div>
          </div>
        </div>
      </div>
    </template>
  </DataView>
</template>
