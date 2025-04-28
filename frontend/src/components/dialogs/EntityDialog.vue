<script setup lang="ts">
  import { ref, watchEffect, computed } from 'vue';
  import { store } from '../../store/index';
  import { Device } from '@/types/device';
  import Selection from '@/components/input/Selection.vue';
  import { KeyValuePair } from '@/types/types.type';

  const props = defineProps<{
    show: boolean;
    id: { type: String; required: true };
    name: { type: String; required: true };
  }>();

  const emit = defineEmits(['close']);


  const showDialog = ref<boolean>(props.show);
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

  watchEffect(() => (showDialog.value = props.show));

  function close() {
    emit('close', false);
    showDialog.value = false;
  }

  const dialogTitle = () => props.title ?? 'Selection';
  const dialogMessage = () => props.message ?? '';
</script>

<template>
  <Dialog v-model:visible="showDialog" modal :header="dialogTitle()" :style="{ width: '25rem' }" @hide="close()">
    <div v-if="dialogMessage()" class="mb-3 text-sm text-color-secondary">
      {{ dialogMessage() }}
    </div>
    <div class="flex items-center gap-4 mb-4">
      <Selection :value="selectedDevice" :items="deviceList" @updated="(value: any) => { selectedDevice = value }" />
    </div>
    <div class="flex justify-end gap-2">
      <Button type="button" label="Cancel" severity="secondary" @click="close()" />
      <Button type="button" label="Save" :disabled="isValid() == false" @click="select()" />
    </div>
  </Dialog>
</template>
