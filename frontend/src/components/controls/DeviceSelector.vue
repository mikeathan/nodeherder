<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import { getDevices } from '@/contracts/device';
  import { store } from '@/store/index';
  import { Device, Devices, DeviceFilter } from '@/types/device';
  import Selection from '@/components/input/Selection.vue';

  const props = defineProps({
    filter: {
      type: Function as PropType<DeviceFilter>,
      default: (device: any, expose: any) => true,
      required: false,
    },
    id: {
      type: String,
      default: '',
      required: false,
    },
    label: {
      type: String,
      default: '',
      required: false,
    },
    disabled: {
      type: Boolean,
      default: false,
      required: false,
    },
  });

  const emit = defineEmits<{
    (e: 'updated', id: string, friendlyName: string): void;
  }>();

  const deviceList = computed(() => {
    var devices = store.getters['hub/listAllDevices']() as Devices;
    if (devices == undefined) {
      return {};
    }
    return getDevices(devices, props.filter);
  });

  function deviceSelected(id: string) {
    const device = store.getters['hub/findDevice'](id) as Device;
    if (device == undefined) {
      console.log('DeviceSelector - device id', id, 'not found');
      emit('updated', '', '');
      return;
    }

    emit('updated', device.id, device.friendly_name);
  }
</script>
<template>
  <Selection
    :label="props.label"
    :value="props.id"
    :disabled="props.disabled"
    @updated="(v) => deviceSelected(v)"
    :filter="true"
    :items="deviceList" />
</template>
