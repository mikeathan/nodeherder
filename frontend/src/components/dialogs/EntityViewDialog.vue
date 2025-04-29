<script setup lang="ts">
  import { ref, watchEffect, computed } from 'vue';
  import { store } from '../../store/index';
  import { Device } from '@/types/device';
  import Selection from '@/components/input/Selection.vue';

  const props = defineProps<{
    show: boolean;
    id: string;
    name: string;
  }>();

  const emit = defineEmits(['close']);

  const showDialog = ref<boolean>(props.show);

  watchEffect(() => (showDialog.value = props.show));

  function close() {
    emit('close', false);
    showDialog.value = false;
  }

  const dialogTitle = () => 'Title';
  const dialogMessage = () => 'Message';
  /* Dynamically compute the styles */
const dialogStyle = computed(() => {
  // Mobile screen styling
  if (window.innerWidth <= 640) {
    return {
      width: '100vw',
      height: '100vh',
      maxHeight: '100vh',
      margin: '0',
      padding: '0',
      transform: 'none',
      borderRadius: '0',
      zIndex: '9999',
    };
  }
  // Default desktop styles
  return {
    width: '500px',
    maxHeight: '80vh',
    overflow: 'auto',
    borderRadius: '1rem',
  };
});
</script>

<template>
  <Dialog
    v-model:visible="showDialog"
    :draggable="false"
    :dismissableMask="true"
    :blockScroll="true"
    modal
    :header="dialogTitle()"
    class="p-dialog-custom entity-modal"
    :breakpoints="{ '640px': '100vw' }"
    :style="dialogStyle"
    @hide="close()">
    <div class="modal-body">
      <p><strong>Name:</strong> {{ dialogMessage() }}</p>
    </div>
  </Dialog>
</template>
<style scoped>

</style>
