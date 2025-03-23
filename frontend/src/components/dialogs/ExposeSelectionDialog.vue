<script setup lang="ts">
  import { ref, watchEffect, onMounted, watch, computed } from 'vue';
  import { store } from '../../store/index';
  import { Device } from '@/types/device';
  import Selection from '@/components/input/Selection.vue';

  const props = defineProps<{
    id: string;
    show: boolean;
  }>();

  const emit = defineEmits(['update', 'close']);

  function select() {
    emit('update', selectedExpose.value);
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
</script>

<template>
  <Dialog v-model:visible="showDialog" modal header="Rename device" :style="{ width: '25rem' }">
    <div class="flex items-center gap-4 mb-4">
      <label for="friendlyNameId" class="font-semibold w-24">Expose</label>
      <Selection
        label="expose"
        :value="selectedExpose"
        text="Expose"
        :items="exposeList"
        @updated="(value:any)=>{selectedExpose=value}" />
    </div>
    <div class="flex justify-end gap-2">
      <Button type="button" label="Cancel" severity="secondary" @click="close()"></Button>
      <Button type="button" label="Save" :disabled="isValid() == false" @click="(e) => select()"></Button>
    </div>
  </Dialog>
</template>
