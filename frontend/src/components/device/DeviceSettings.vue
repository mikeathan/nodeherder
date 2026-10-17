<script setup lang="ts">
  import DeviceConfigurationSettings from '@/components/settings/DeviceConfigurationSettings.vue';
  defineProps<{ id: string }>();
</script>

<template>
  <DeviceConfigurationSettings :id="id" />
</template>
