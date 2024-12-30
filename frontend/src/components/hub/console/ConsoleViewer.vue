<script setup lang="ts">
  import {
    computed,
    nextTick,
    onMounted,
    ref,
    watch,
  } from 'vue';
  import { key, store } from '../../../store/index';
  import Toggle from '../../input/Toggle.vue';
  import { consoleCleanupService } from '@/services/console-cleanup.service';
  import { LogMessageType } from '@/types/console.type';
  import { formatTimestamp } from '@/utils/date.utils';
  import { LoggerSettingsType } from '@/types/settings';
  import { getConsoleLevelClass } from '@/contracts/console';

  onMounted(() => {
    consoleCleanupService.startTimer(store);
  });

  const loggerSettings = computed(() => {
    if (
      !store.getters['appconfig/initialized']() as Boolean
    ) {
      store.dispatch('ws/emit', { event: 'loadAppConfig' });
    }
    const set = store.getters['appconfig/logger']();
    if (set == undefined) {
      return {} as LoggerSettingsType;
    }
    return store.getters[
      'appconfig/logger'
    ]() as LoggerSettingsType;
  });

  const messages = computed(() => {
    return store.getters[
      'console/messages'
    ]() as LogMessageType[];
  });

  const scrollToBottom = () => {
    if (messageContainer.value) {
      messageContainer.value.scrollTop =
        messageContainer.value.scrollHeight;
    }
  };
  const messageContainer = ref<HTMLDivElement | null>(null);
  watch(
    messages,
    (newMessages, oldMessages) => {
      // if (newMessages.length !== oldMessages?.length) {
      nextTick(() => scrollToBottom());
    },
    { deep: true }
  );
  onMounted(() => {
    scrollToBottom();
  });

  function enableLogging(enabled: boolean) {
    if (
      enabled == loggerSettings.value.enableRemoteLogger
    ) {
      return;
    }
    loggerSettings.value.enableRemoteLogger = enabled;
    store.dispatch(
      'appconfig/saveLoggerSettings',
      loggerSettings.value
    );
  }

  function clearConsole() {
    store.commit('console/clear');
  }
</script>

<style>
  .message-container {
    max-height: 200px;
    overflow-y: auto;
    padding-right: 10px;
  }
</style>
<template>
  <Card>
    <template #title>
      <h2>Remote logger</h2>
    </template>
    <template #content>
      <Toggle
        :minimal="true"
        :value="loggerSettings.enableRemoteLogger"
        :valueOn="true"
        :valueoff="false"
        @update="(v: boolean) => enableLogging(v)">
      </Toggle>
      <Button
        label="Clear"
        @click="clearConsole"
        variant="text"
        icon="pi pi-delete-left" />
      <div class="pb-3"></div>
      <div ref="messageContainer" class="message-container">
        <div
          v-for="(message, index) in messages"
          :key="message.timestamp">
          <span
            style="width: 60px"
            :class="`badge ${getConsoleLevelClass(
              message.level
            )}`"
            >{{ message.level }}</span
          >
          &nbsp;
          <small class="pe-1">{{
            formatTimestamp(message.timestamp)
          }}</small>
          &nbsp;
          <code>{{ message.message }}</code>
        </div>
      </div>
    </template>
  </Card>
</template>
