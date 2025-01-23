<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue';
  import { store } from '../../../store/index';
  import HistorySettings from './HistorySettings.vue';
  import { HistorySettingsType, LoggerSettingsType } from '@/types/settings.type';
  import LoggerSettings from './LoggerSettings.vue';

  const historySettings = computed(() => {
    return store.getters['hub/history']() as HistorySettingsType;
  });

  const loggerSettings = computed(() => {
    return store.getters['hub/logger']() as LoggerSettingsType;
  });
</script>

<template>
  <Card>
    <template #title>
      <h2>Settings</h2>
    </template>
    <template #content>
      <div class="pt-3">
        <HistorySettings :settings="historySettings" />
      </div>
      <div class="pt-3">
        <LoggerSettings :settings="loggerSettings" />
      </div>
    </template>
  </Card>
</template>
