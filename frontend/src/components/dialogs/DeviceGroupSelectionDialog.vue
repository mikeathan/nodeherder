<script setup lang="ts">
  import { ref, watchEffect, computed } from 'vue';
  import { store } from '../../store/index';
  import { Device } from '@/types/device';
  import Selection from '@/components/input/Selection.vue';
  import { KeyValuePair } from '@/types/types.type';
  import { DashboardGroup, DeviceGroup } from '@/types/settings.type';
  import MultiSelection from '../input/MultiSelection.vue';

  const props = defineProps<{
    show: boolean;
    title?: string;
    message?: string;
    dashboardGroup: DashboardGroup;
  }>();

  const emit = defineEmits<{
    (e: 'confirm', deviceGroup: DeviceGroup): void;
    (e: 'close', value: boolean): void;
  }>();

  function select() {
    emit('confirm', selectedDeviceGroup.value);
    close();
  }

  const dialogTitle = () => props.title ?? 'Selection';
  const dialogMessage = () => props.message ?? '';

  const selectedDeviceGroup = ref<DeviceGroup>({} as DeviceGroup);

  const showDialog = ref<boolean>(props.show);
  watchEffect(() => (showDialog.value = props.show));

  const deviceList = computed(() => {
    const devices = store.getters['hub/listAllDevices']() as Device[];
    if (devices == undefined) {
      console.log('no devices found devices');
      return {} as KeyValuePair<string>;
    }

    return devices.reduce<KeyValuePair<string>>((acc, item) => {
      acc[item.friendly_name] = item.id;
      return acc;
    }, {});
  });

  const exposeList = computed(() => {
    if (selectedDeviceGroup.value == null) {
      return Array<string>();
    }
    const deviceId = selectedDeviceGroup.value.deviceId;
    const device = store.getters['hub/findDevice'](deviceId) as Device;
    if (device == undefined) {
      console.log('exposeList empty', deviceId);
      return Array<string>();
    }

    return Object.entries(device.exposes).map(([i, e]) => e.name);
  });

  function selectDevice(deviceId: string) {
    selectedDeviceGroup.value = {
      deviceId: deviceId,
      exposes: props.dashboardGroup.deviceGroup[deviceId]?.exposes ?? [],
    };
  }

  function close() {
    emit('close', false);
    showDialog.value = false;
  }

  function isValid(): boolean {
    const group = selectedDeviceGroup.value;
    if (group?.deviceId == null || group.exposes == null || group.exposes?.length == 0) {
      return false;
    }

    const existingExposes = props.dashboardGroup.deviceGroup?.[group.deviceId]?.exposes;
    if (existingExposes) {
      return !group.exposes.every((e) => existingExposes.includes(e));
    }

    return true;
  }
</script>

<template>
  <Dialog v-model:visible="showDialog" modal :header="dialogTitle()" :style="{ width: '25rem' }" @hide="close()">
    <div v-if="dialogMessage()" class="mb-3 text-sm text-color-secondary">
      {{ dialogMessage() }}
    </div>
    <div class="flex items-center gap-4 mt-2 mb-4">
      <Selection
        label="Select device"
        :value="selectedDeviceGroup.deviceId"
        :items="deviceList"
        @updated="selectDevice"
        :filter="true" />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <MultiSelection
        :filter="true"
        label="Select entities"
        showClear
        :values="selectedDeviceGroup.exposes"
        :items="exposeList"
        @updated="(values: any) => { selectedDeviceGroup.exposes= values }" />
    </div>
    <div class="flex justify-end gap-2">
      <Button type="button" label="Cancel" severity="secondary" @click="close()" />
      <Button type="button" label="Save" :disabled="isValid() == false" @click="select()" />
    </div>
  </Dialog>
</template>
