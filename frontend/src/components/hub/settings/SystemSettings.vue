<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue';
  import { get } from '@/contracts/api';
  import UiButton from '@/components/ui/UiButton.vue';

  type Entry = { label: string; bytes: number; retentionDays?: number };
  type Usage = { entries: Entry[]; dataDirBytes: number; diskFreeBytes: number; diskTotalBytes: number; metricsRetentionDays: number; positionsRetentionDays: number };
  const usage = ref<Usage | null>(null);
  const error = ref('');
  const loading = ref(false);
  const entries = computed(() => usage.value?.entries ?? []);
  function size(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(0)} KB`;
    if (bytes < 1024 ** 3) return `${(bytes / 1024 ** 2).toFixed(1)} MB`;
    return `${(bytes / 1024 ** 3).toFixed(2)} GB`;
  }
  async function load() {
    loading.value = true; error.value = '';
    try {
      const response = await get('/system/storage');
      if (!response.ok) throw new Error('Storage usage could not be read.');
      usage.value = (await response.json()) as Usage;
    } catch (cause) { error.value = (cause as Error).message; }
    finally { loading.value = false; }
  }
  onMounted(load);
</script>

<template>
  <section aria-label="Storage usage">
    <p class="nh-muted">Read-only sizes on this hub. Positions are retained for {{ usage?.positionsRetentionDays ?? 7 }} days; device metrics for {{ usage?.metricsRetentionDays ?? 10 }} days.</p>
    <p v-if="error" role="alert">{{ error }}</p>
    <p v-else-if="loading" role="status">Reading storage usage…</p>
    <table v-else-if="usage" class="nh-storage-table">
      <thead><tr><th scope="col">Store</th><th scope="col">Size</th><th scope="col">Retention</th></tr></thead>
      <tbody>
        <tr v-for="entry in entries" :key="entry.label">
          <th scope="row">{{ entry.label }}</th>
          <td>{{ size(entry.bytes) }}</td>
          <td>{{ entry.retentionDays ? `${entry.retentionDays} days` : '—' }}</td>
        </tr>
        <tr><th scope="row">Data directory total</th><td>{{ size(usage.dataDirBytes) }}</td><td>—</td></tr>
        <tr v-if="usage.diskTotalBytes"><th scope="row">Disk</th><td>{{ size(usage.diskFreeBytes) }} free of {{ size(usage.diskTotalBytes) }}</td><td>—</td></tr>
      </tbody>
    </table>
    <UiButton size="sm" variant="ghost" :disabled="loading" @click="load">Refresh</UiButton>
  </section>
</template>

<style scoped>
  .nh-storage-table { width: 100%; border-collapse: collapse; margin-block: 0.75rem; }
  .nh-storage-table th, .nh-storage-table td { text-align: start; padding: 0.4rem 0.5rem; border-block-end: 1px solid var(--nh-border); }
</style>
