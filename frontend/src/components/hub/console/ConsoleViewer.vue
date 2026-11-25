<script setup lang="ts">
  /*
   * Logs: bounded live stream with local filtering and export. Level filtering is local;
   * the hub's logger level lives in Settings → Logs & diagnostics so the two controls cannot
   * disagree. Follow only while the user is at the bottom; never steal scroll.
   */
  import { computed, nextTick, onMounted, ref, watch } from 'vue';
  import { store } from '@/store';
  import { RouteName } from '@/types/router';
  import { consoleCleanupService } from '@/services/console-cleanup.service';
  import { downloadTextFile } from '@/services/logs';
  import type { LogMessageType } from '@/types/console.type';
  import { formatTimestamp } from '@/utils/date.utils';
  import { getConsoleLevelSeverity } from '@/contracts/console';
  import Tag from 'primevue/tag';
  import Select from 'primevue/select';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const levels = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'];
  const levelOptions = [{ value: 'all', label: 'All levels' }, ...levels.map((level) => ({ value: level, label: level }))];
  const level = ref('all');
  const query = ref('');
  const problemsOnly = ref(false);
  const following = ref(true);
  const messageContainer = ref<HTMLDivElement | null>(null);

  onMounted(() => consoleCleanupService.startTimer(store));

  const messages = computed(() => store.getters['console/messages']() as LogMessageType[]);
  const filtered = computed(() => {
    const search = query.value.trim().toLowerCase();
    return messages.value.filter((message) => {
      const messageLevel = message.level.toLowerCase();
      if (level.value !== 'all' && messageLevel !== level.value) return false;
      if (problemsOnly.value && !['warn', 'error', 'fatal'].includes(messageLevel)) return false;
      return !search || `${message.message} ${message.level}`.toLowerCase().includes(search);
    });
  });

  function scrollToBottom() {
    if (messageContainer.value) messageContainer.value.scrollTop = messageContainer.value.scrollHeight;
  }

  // New rows only matter when the user is following; deep-watching the array would walk every
  // message on each update.
  watch(
    () => filtered.value.length,
    () => {
      if (following.value) void nextTick(scrollToBottom);
    }
  );

  function onScroll() {
    const element = messageContainer.value;
    if (!element) return;
    following.value = element.scrollHeight - element.scrollTop - element.clientHeight < 40;
  }

  function download() {
    const body = filtered.value.map((message) => `${message.timestamp} ${message.level.toUpperCase()} ${message.message}`).join('\n');
    downloadTextFile(`nodeherder-logs-${new Date().toISOString().slice(0, 10)}.log`, body);
  }

  function clearConsole() {
    store.commit('console/clear');
  }
</script>

<template>
  <div>
    <UiPageHeader title="Logs" subtitle="Live messages from the hub. Filtering and export are local to this browser.">
      <template #actions>
        <UiButton icon="settings" :to="{ name: RouteName.Settings, query: { section: 'logger' } }">Log settings</UiButton>
      </template>
    </UiPageHeader>

    <div class="nh-card nh-card-body">
      <div class="nh-log-toolbar">
        <label class="nh-log-search">
          <UiIcon name="search" />
          <span class="sr-only">Search log messages</span>
          <input v-model="query" class="nh-input" type="search" placeholder="Search messages…" />
        </label>
        <Select v-model="level" :options="levelOptions" option-label="label" option-value="value" aria-label="Filter by level" class="nh-log-level" />
        <UiButton
          :variant="problemsOnly ? 'primary' : 'default'"
          icon="warn"
          :aria-pressed="problemsOnly"
          @click="problemsOnly = !problemsOnly">
          Problems
        </UiButton>
        <UiButton :variant="following ? 'default' : 'primary'" :icon="following ? 'pause' : 'run'" :aria-pressed="!following" @click="following = !following">
          {{ following ? 'Follow' : 'Paused' }}
        </UiButton>
        <UiButton icon="download" @click="download">Download</UiButton>
        <UiButton icon="delete" variant="ghost" @click="clearConsole">Clear</UiButton>
      </div>

      <div ref="messageContainer" class="message-container" @scroll.passive="onScroll">
        <div
          v-for="(message, index) in filtered"
          :key="`${message.timestamp}-${index}`"
          class="log-entry"
          :class="message.level.toLowerCase()">
          <div class="log-meta">
            <Tag :severity="getConsoleLevelSeverity(message.level)" :value="message.level" class="text-xs uppercase font-bold" style="min-width: 60px" />
            <span class="nh-log-time">{{ formatTimestamp(message.timestamp) }}</span>
          </div>
          <span class="log-message">{{ message.message }}</span>
        </div>
        <div v-if="!filtered.length" class="nh-log-empty">
          <UiIcon name="console" />
          <span>{{ messages.length ? 'No messages match these filters.' : 'No logs yet. New hub messages appear here while this tab is open.' }}</span>
        </div>
      </div>
      <p class="nh-log-foot">
        {{ filtered.length }} of {{ messages.length }} messages
        <span v-if="!following"> · not following new messages</span>
        <span v-else> · following</span>
      </p>
    </div>
  </div>
</template>

<style scoped>
  .nh-log-toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
    padding-bottom: 0.75rem;
  }
  .nh-log-search {
    position: relative;
    display: flex;
    align-items: center;
    flex: 1 1 14rem;
    max-width: 24rem;
  }
  .nh-log-search .nh-ic {
    position: absolute;
    left: 0.65rem;
    color: var(--nh-text-muted);
    pointer-events: none;
  }
  .nh-log-search .nh-input {
    width: 100%;
    padding-left: 2.2rem;
  }
  .nh-log-level {
    min-width: 9rem;
  }
  .message-container {
    max-height: calc(100dvh - 19rem);
    min-height: 16rem;
    overflow-y: auto;
    padding: 0.5rem;
    background: var(--nh-surface-2);
    border-radius: var(--nh-radius-m);
    border: var(--nh-border-w) solid var(--nh-border);
  }
  .log-entry {
    display: flex;
    gap: 0.75rem;
    margin-bottom: 0.25rem;
    align-items: flex-start;
    padding: 0.3rem 0.5rem;
    border-radius: var(--nh-radius-s);
  }
  .log-entry:hover {
    background: var(--nh-surface);
  }
  .log-meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 180px;
    flex-shrink: 0;
  }
  .nh-log-time {
    color: var(--nh-text-muted);
    font-family: var(--nh-font-mono);
    font-size: 0.72rem;
  }
  .log-message {
    color: var(--nh-text);
    font-family: var(--nh-font-mono);
    font-size: 0.825rem;
    overflow-wrap: anywhere;
    white-space: pre-wrap;
    flex-grow: 1;
    line-height: 1.45;
  }
  .log-entry.debug .log-message { color: #c792ea; }
  .log-entry.trace .log-message { color: #9d8cff; }
  .log-entry.warn .log-message { color: var(--nh-warn); }
  .log-entry.error .log-message,
  .log-entry.fatal .log-message { color: var(--nh-danger); }
  .nh-log-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.4rem;
    padding: 2rem 1rem;
    color: var(--nh-text-muted);
    font-size: 0.85rem;
    text-align: center;
  }
  .nh-log-empty .nh-ic {
    width: 1.6rem;
    height: 1.6rem;
  }
  .nh-log-foot {
    margin: 0.5rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.72rem;
  }
  @media (max-width: 768px) {
    .log-entry {
      flex-direction: column;
      gap: 0.15rem;
      padding: 0.5rem;
      margin-bottom: 0.5rem;
      border: var(--nh-border-w) solid var(--nh-border);
    }
    .log-meta {
      min-width: 0;
      width: 100%;
      justify-content: space-between;
    }
  }
</style>
