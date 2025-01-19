<script setup lang="ts">
  import { computed, PropType, ref } from 'vue';
  import { getExposes } from '@/contracts/device';
  import { store } from '@/store/index';
  import { Device, DeviceFilter } from '@/types/device';
  import Selection from '@/components/input/Selection.vue';

  const props = defineProps({
    id: {
      type: String,
      default: '',
      required: true,
    },
    value: {
      type: String,
      default: '',
      required: false,
    },
    filter: {
      type: Function as PropType<DeviceFilter>,
      default: (expose: any) => true,
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
    (e: 'updated', valueid: string): void;
  }>();

  const selectedExpose = ref<string>(props.value);

  const exposeList = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (device == undefined) {
      console.log('exposeList empty', props.id);

      return Array<string>();
    }

    return getExposes(device, props.filter);
  });

  function exposeSelected(value: string) {
    selectedExpose.value = value;
    emit('updated', selectedExpose.value);
  }
</script>

<template>
  <Selection
    :label="props.label"
    :value="props.value"
    :disabled="props.disabled"
    @updated="exposeSelected"
    :items="exposeList" />
</template>
