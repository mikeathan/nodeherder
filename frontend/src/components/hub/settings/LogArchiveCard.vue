<script setup lang="ts">
  /*
   * Archived hub log files. Listing is user-initiated (no render-path request); downloads
   * use the existing protected endpoints through `services/logs`.
   */
  import { ref } from 'vue';
  import { deleteLogFile, downloadLogFile, listLogFiles, logFileName, type LogFile } from '@/services/logs';
  import { formatQuantity } from '@/utils/units';
  import RelativeTime from '@/components/ui/RelativeTime.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import { confirm } from '@/composables/useConfirm';

  const files = ref<LogFile[] | null>(null);
  const loading = ref(false);
  const error = ref('');

  async function load() {
    loading.value = true;
    error.value = '';
    try {
      files.value = await listLogFiles();
    } catch (reason) {
      error.value = reason instanceof Error ? reason.message : 'Could not list log files.';
    } finally {
      loading.value = false;
    }
  }

  async function remove(path: string) {
    const name = logFileName(path);
    const ok = await confirm({
      title: `Delete ${name}?`,
      message: 'The archived file is removed from the hub. This cannot be undone.',
      confirmLabel: 'Delete file',
      danger: true,
    });
    if (!ok) return;
    error.value = '';
    try {
      await deleteLogFile(path);
      await load();
    } catch (reason) {
      error.value = reason instanceof Error ? reason.message : 'Could not delete that log file.';
    }
  }

  async function download(path: string) {
    error.value = '';
    try {
      await downloadLogFile(path);
    } catch (reason) {
      error.value = reason instanceof Error ? reason.message : 'Could not download that log file.';
    }
  }
</script>

<template>
  <UiCard title="Archived files" icon="settingsLogger" :heading-level="3">
    <template #tools>
      <UiButton size="sm" variant="ghost" icon="refresh" :loading="loading" @click="load">
        {{ files === null ? 'Load files' : 'Refresh' }}
      </UiButton>
    </template>

    <p v-if="error" class="nh-alert is-danger" role="alert"><UiIcon name="error" /><span>{{ error }}</span></p>
    <p v-else-if="files === null" class="nh-muted">
      Files are listed from the hub when you ask. Downloads use your signed-in session.
    </p>
    <p v-else-if="!files.length" class="nh-muted">No archived log files on the hub yet.</p>
    <ul v-else class="nh-log-files">
      <li v-for="file in files" :key="file.name">
        <span class="nh-log-file-name"><UiIcon name="console" />{{ logFileName(file.name) }}</span>
        <span class="nh-log-file-meta">
          {{ formatQuantity(file.size, 'bytes') }}<template v-if="file.modified"> · <RelativeTime :value="file.modified" /></template>
        </span>
        <UiButton size="sm" icon="download" :label="`Download ${logFileName(file.name)}`" @click="download(file.name)" />
        <UiButton
          size="sm"
          variant="ghost"
          icon="delete"
          class="is-danger"
          :label="`Delete ${logFileName(file.name)}`"
          @click="remove(file.name)" />
      </li>
    </ul>
  </UiCard>
</template>

<style scoped>
  .nh-log-files {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .nh-log-files li {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    min-height: var(--nh-touch-h);
    border-bottom: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-log-files li:last-child {
    border-bottom: 0;
  }
  .nh-log-file-meta {
    color: var(--nh-text-muted);
    font-size: 0.7rem;
    white-space: nowrap;
  }
  .nh-log-file-name {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    min-width: 0;
    flex: 1;
    font-size: 0.82rem;
    overflow-wrap: anywhere;
  }
  .nh-log-file-name .nh-ic {
    width: 0.95rem;
    height: 0.95rem;
    color: var(--nh-text-muted);
    flex: none;
  }
</style>
