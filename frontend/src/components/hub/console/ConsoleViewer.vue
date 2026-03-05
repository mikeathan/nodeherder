<script setup lang="ts">
  import { computed, nextTick, onMounted, ref, watch } from 'vue';
  import { key, store } from '../../../store/index';
  import Toggle from '../../input/Toggle.vue';
  import { consoleCleanupService } from '@/services/console-cleanup.service';
  import { LogMessageType } from '@/types/console.type';
  import { formatTimestamp } from '@/utils/date.utils';
  import { LoggerSettingsType } from '@/types/settings.type';
  import { getConsoleLevelSeverity } from '@/contracts/console';
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
    () => {
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

<style scoped>
  .message-container {
    max-height: 600px;
    overflow-y: auto;
    padding: 0.5rem;
    background: var(--surface-b);
    border-radius: var(--border-radius);
    border: 1px solid var(--surface-border);
  }

  .log-entry {
    display: flex;
    gap: 0.75rem;
    margin-bottom: 0.25rem;
    align-items: flex-start;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    transition: background 0.2s;
  }

  .log-entry:hover {
    background: var(--surface-hover);
  }

  .log-meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 180px;
    flex-shrink: 0;
  }

  .log-message {
    color: var(--text-color);
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 0.825rem;
    word-break: break-all;
    white-space: pre-wrap;
    flex-grow: 1;
    line-height: 1.4;
  }

  .log-entry.debug .log-message {
    color: #ff79c6;
  }
  .log-entry.trace .log-message {
    color: #bd93f9;
  }
  .log-entry.info .log-message {
    color: #8be9fd;
  }
  .log-entry.warn .log-message {
    color: #f1fa8c;
  }
  .log-entry.error .log-message,
  .log-entry.fatal .log-message {
    color: #ff5555;
  }

  @media screen and (max-width: 768px) {
    .log-entry {
      flex-direction: column;
      gap: 0.125rem;
      padding: 0.5rem;
      margin-bottom: 0.5rem;
      border: 1px solid var(--surface-border);
      border-radius: var(--border-radius);
    }

    .log-entry.debug {
      background: rgba(255, 121, 198, 0.05);
    }
    .log-entry.trace {
      background: rgba(189, 147, 249, 0.05);
    }
    .log-entry.info {
      background: rgba(139, 233, 253, 0.05);
    }
    .log-entry.warn {
      background: rgba(241, 250, 140, 0.05);
    }
    .log-entry.error,
    .log-entry.fatal {
      background: rgba(255, 85, 85, 0.05);
    }

    .log-meta {
      min-width: auto;
      width: 100%;
      justify-content: space-between;
    }
  }
</style>
<template>
  <Card>
    <template #title>
      <h2 class="m-0">Remote logger</h2>
    </template>
    <template #content>
      <div class="flex flex-wrap align-items-center gap-3 pb-3">
        <div class="flex align-items-center gap-2 bg-black-alpha-10 p-2 border-round">
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
          <Button
            label="Clear"
            @click="clearConsole"
            variant="text"
            size="small"
            icon="pi pi-delete-left"
            class="ml-1" />
        </div>
      </div>
      <div ref="messageContainer" class="message-container">
        <div
          v-for="message in messages"
          :key="message.timestamp"
          class="log-entry"
          :class="[message.level.toLowerCase()]">
          <div class="log-meta">
            <Tag
              :severity="getConsoleLevelSeverity(message.level)"
              :value="message.level"
              class="text-xs uppercase font-bold"
              style="min-width: 60px" />
            <span class="text-400 font-mono text-xs opacity-60">{{ formatTimestamp(message.timestamp) }}</span>
          </div>
          <span class="log-message">{{ message.message }}</span>
        </div>
        <div
          v-if="messages.length === 0"
          class="flex flex-column align-items-center justify-content-center p-5 opacity-30">
          <i class="pi pi-inbox text-4xl mb-2"></i>
          <span>No logs available</span>
        </div>
      </div>
    </template>
  </Card>
</template>
