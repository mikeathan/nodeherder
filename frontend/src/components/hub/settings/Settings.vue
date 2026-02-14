<script setup lang="ts">
  import HistorySettings from './HistorySettings.vue';
  import LoggerSettings from './LoggerSettings.vue';
  import DeviceDefaultSettings from './DeviceDefaultSettings.vue';
  import MCPSettings from './MCPSettings.vue';
</script>

<template>
  <Card>
    <template #title>
      <h2>Settings</h2>
    </template>
    <template #content>
      <div class="pt-3">
        <DeviceDefaultSettings />
      </div>
      <div class="pt-3">
        <HistorySettings />
      </div>
      <div class="pt-3">
        <LoggerSettings />
      </div>
      <div class="pt-3">
        <MCPSettings />
      </div>
    </template>
  </Card>
</template>
