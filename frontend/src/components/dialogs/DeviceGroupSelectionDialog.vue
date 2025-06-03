<script setup lang="ts">
  import { ref, watchEffect, computed } from 'vue';
  import { store } from '../../store/index';
  import { Device } from '@/types/device';
  import Selection from '@/components/input/Selection.vue';
  import { KeyValuePair } from '@/types/types.type';
  import { DashboardGroup } from '@/types/settings.type';
import MultiSelection from '../input/MultiSelection.vue';

  const props = defineProps<{
    show: boolean;
    title?: string;
    message?: string;
    dashboardGroup: DashboardGroup;
  }>();

  const emit = defineEmits(['confirm', 'close']);

  function select() {
    //   emit('confirm', selectedDevice.value);
    close();
  }

  const dialogTitle = () => props.title ?? 'Selection';
  const dialogMessage = () => props.message ?? '';

  const selectedDevice = ref<string | null>(null);
  const selectedExposes = ref<string[]>(Array<string>());;

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
    if (selectedDevice.value == null) {
      return Array<string>();
    }
    const device = store.getters['hub/findDevice'](selectedDevice.value) as Device;
    if (device == undefined) {
      console.log('exposeList empty', selectedDevice.value);
      return Array<string>();
    }

      TODO
    // TODO use dashboardGroup to preselect exposes for selected device
    

    return Object.entries(device.exposes).map(([i, e]) => e.name);
  });

    function selectDevice(deviceId: string) {
      selectedDevice.value = deviceId;
      selectedExposes.value = [];
    }
  function close() {
    emit('close', false);
    showDialog.value = false;
  }

  function isValid(): boolean {
    return false;
  }
</script>

<template>
  <Dialog v-model:visible="showDialog" modal :header="dialogTitle()" :style="{ width: '25rem' }" @hide="close()">
    <div v-if="dialogMessage()" class="mb-3 text-sm text-color-secondary">
      {{ dialogMessage() }}
    </div>
    <div class="flex items-center gap-4 mt-2 mb-4">
      <Selection label="Select device" :value="selectedDevice" :items="deviceList" @updated="selectDevice" />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <MultiSelection label="Select entities" showClear :values="selectedExposes" :items="exposeList" @updated="(values: any) => { selectedExposes= values }" />
    </div>
    <div class="flex justify-end gap-2">
      <Button type="button" label="Cancel" severity="secondary" @click="close()" />
      <Button type="button" label="Save" :disabled="isValid() == false" @click="select()" />
    </div>
  </Dialog>
</template>
