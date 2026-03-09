<script setup lang="ts">
import { ref, watchEffect, computed } from 'vue';
import { store } from '../../store/index';
import { Device } from '@/types/device';
import Selection from '@/components/input/Selection.vue';

const props = defineProps<{
  id: string;
  show: boolean;
  title?: string;
  message?: string;
}>();

const emit = defineEmits(['confirm', 'close']);

function select() {
  emit('confirm', selectedExpose.value);
  close();
}

const selectedExpose = ref<string | null>(null);
const showDialog = ref<boolean>(props.show);
const exposeList = computed(() => {
  const device = store.getters['hub/findDevice'](props.id) as Device;
  if (device == undefined) {
    console.log('exposeList empty', props.id);
    return Array<string>();
  }

  return Object.entries(device.exposes).map(([i, e]) => e.name);
});

watchEffect(() => (showDialog.value = props.show));

function close() {
  emit('close', false);
  showDialog.value = false;
}

function isValid() {
  return selectedExpose.value != null;
}
const dialogTitle = () => props.title ?? 'Selection';
const dialogMessage = () => props.message ?? '';
</script>

<template>
  <Dialog v-model:visible="showDialog" modal :header="dialogTitle()" :style="{ width: '25rem' }"@hide="close()">
    <div v-if="dialogMessage()" class="mb-3 text-sm text-color-secondary">
      {{ dialogMessage() }}
    </div>
    <div class="flex items-center gap-4 mb-4">
      <Selection :value="selectedExpose" :items="exposeList" @updated="(value: any) => { selectedExpose = value }" />
    </div>
    <div class="flex justify-end gap-2">
      <Button type="button" label="Cancel" severity="secondary" @click="close()"></Button>
      <Button type="button" label="Save" :disabled="isValid() == false" @click="select()"></Button>
    </div>
  </Dialog>
</template>
