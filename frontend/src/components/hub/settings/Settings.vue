<script setup lang="ts">
import HistorySettings from './HistorySettings.vue';
import LoggerSettings from './LoggerSettings.vue';

</script>

<template>
  <Card>
    <template #title>
      <h2>Settings</h2>
    </template>
    <template #content>
      <div class="pt-3">
        <HistorySettings />
      </div>
      <div class="pt-3">
        <LoggerSettings />
      </div>
    </template>
  </Card>
</template>
