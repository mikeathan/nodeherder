<script setup lang="ts">
  /* ESPHome entry points in the Devices header: add flow plus management dialog. */
  import { ref } from 'vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ESPHomeAddDialog from '@/components/device/esphome/ESPHomeAddDialog.vue';
  import ESPHomeManageDialog from '@/components/device/esphome/ESPHomeManageDialog.vue';

  const adding = ref(false);
  const managing = ref(false);

  /** Opens management after a successful enrollment so the new device is visible. */
  function onAdded() {
    managing.value = true;
  }
</script>

<template>
  <UiButton icon="add" @click="adding = true">Add ESPHome device</UiButton>
  <UiButton icon="settings" @click="managing = true">Manage ESPHome</UiButton>
  <ESPHomeAddDialog :open="adding" @close="adding = false" @added="onAdded" />
  <ESPHomeManageDialog :open="managing" @close="managing = false" />
</template>
