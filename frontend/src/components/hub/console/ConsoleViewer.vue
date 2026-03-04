<script setup lang="ts">
  import { computed, nextTick, onMounted, ref, watch } from 'vue';
  import { key, store } from '../../../store/index';
  import Toggle from '../../input/Toggle.vue';
  import { consoleCleanupService } from '@/services/console-cleanup.service';
  import { LogMessageType } from '@/types/console.type';
  import { formatTimestamp } from '@/utils/date.utils';
  import { LoggerSettingsType } from '@/types/settings.type';
  import { getConsoleLevelClass, getConsoleLevelSeverity } from '@/contracts/console';
  import Tag from 'primevue/tag';
  import Select from 'primevue/select';

  const levelOptions = ref(['trace', 'debug', 'info', 'warn', 'error', 'fatal']);

  onMounted(() => {
    consoleCleanupService.startTimer(store);
  });

  const loggerSettings = computed(() => {
    const set = store.getters['hub/logger']();
    if (set == undefined) {
      return {} as LoggerSettingsType;
    }
    return store.getters['hub/logger']() as LoggerSettingsType;
  });

  const messages = computed(() => {
    return store.getters['console/messages']() as LogMessageType[];
  });

  const scrollToBottom = () => {
    if (messageContainer.value) {
      messageContainer.value.scrollTop = messageContainer.value.scrollHeight;
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

  function updateSettings(key: keyof LoggerSettingsType, value: any) {
    if (value === loggerSettings.value[key]) {
      return;
    }
    const newSettings = { ...loggerSettings.value, [key]: value };
    store.dispatch('hub/saveLoggerSettings', newSettings);
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
      <div class="flex align-items-center gap-3 pb-3">
        <Toggle
          :value="loggerSettings.enableRemoteLogger"
          :valueOn="true"
          :valueOff="false"
          @update="(v: boolean) => updateSettings('enableRemoteLogger', v)" />
        <Select
          :modelValue="loggerSettings.level"
          :options="levelOptions"
          @update:modelValue="(v: any) => updateSettings('level', v)"
          class="w-auto text-sm h-2rem" />
        <Button label="Clear" @click="clearConsole" variant="text" icon="pi pi-delete-left" />
      </div>
      <div ref="messageContainer" class="message-container">
        <div v-for="(message, index) in messages" :key="message.timestamp" class="flex align-items-center mb-1">
          <div style="min-width: 70px">
            <Tag :severity="getConsoleLevelSeverity(message.level)" :value="message.level" class="w-full" />
          </div>
          <span class="text-400 ml-3 mr-3 font-mono text-sm">{{ formatTimestamp(message.timestamp) }}</span>
          <span style="color: #ff79c6; font-family: monospace" class="text-sm border-0">{{ message.message }}</span>
        </div>
      </div>
    </template>
  </Card>
</template>
