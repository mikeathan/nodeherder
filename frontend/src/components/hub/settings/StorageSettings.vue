<script setup lang="ts">
  /*
   * Storage & retention (spec 059 FR-02): the hub's retention policy and a read-only
   * space report together, so the same numbers are not shown in two sections.
   * Retention saves through the existing `hub/saveHistorySettings` action; usage is read
   * from the existing `GET /system/storage` endpoint.
   */
  import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
  import { store } from '@/store';
  import { get } from '@/contracts/api';
  import type { HistorySettingsType } from '@/types/settings.type';
  import type { TimeInterval, TimeUnit } from '@/types/types.type';
  import {
    RETENTION_DEFAULT,
    RETENTION_FIELDS,
    RETENTION_UNITS,
    canPersistRetention,
    isSameRetention,
    parseRetentionInterval,
    retentionValues,
    withRetentionField,
    type RetentionFieldId,
  } from '@/domain/settings/policy';
  import { formatQuantity } from '@/utils/units';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiSelect from '@/components/ui/UiSelect.vue';
  import UiSettingRow from '@/components/ui/UiSettingRow.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  interface StorageEntry {
    label: string;
    bytes: number;
    retentionDays?: number;
  }

  interface StorageUsage {
    entries: StorageEntry[];
    dataDirBytes: number;
    diskFreeBytes: number;
    diskTotalBytes: number;
    metricsRetentionDays: number;
    positionsRetentionDays: number;
  }

  const INVALID_MESSAGE = 'Enter a whole number of at least one minute, up to one year.';

  const history = computed<HistorySettingsType | undefined>(() => store.getters['hub/history']());
  const values = computed(() => retentionValues(history.value));
  const offline = computed(() => store.getters['ws/getConnectionStatus'] !== 'connected');
  /**
   * Saved policy is unknown until the HTTP snapshot lands: `hub/history` then reports the
   * placeholder config, where both intervals are `0 days`. Editing from that state would
   * send the untouched field as zero, which the hub reads as "keep nothing", so the card
   * stays locked until a full policy is available.
   */
  const ready = computed(() => store.getters['hub/isInitialized']() === true
    && !!history.value
    && canPersistRetention(history.value));
  const retentionLocked = computed(() => offline.value || !ready.value);

  const errors = ref<Partial<Record<RetentionFieldId, string>>>({});

  /** Draft text per field so a rejected value stays visible while the user corrects it. */
  const drafts = ref<Partial<Record<RetentionFieldId, string>>>({});

  const usage = ref<StorageUsage | null>(null);
  const usageError = ref('');
  const usageLoading = ref(false);

  const unitOptions = RETENTION_UNITS.map((unit) => ({ value: unit.value, label: unit.label }));

  /** Blank until a real policy is known; never render the placeholder's `0 days` as if saved. */
  function fieldValue(field: RetentionFieldId): string {
    if (!ready.value) return '';
    return drafts.value[field] ?? String(values.value[field].value);
  }

  /**
   * Unit shown beside the number. Saving is blocked until a real policy loads, so falling
   * back to the documented default here only affects the disabled loading state.
   */
  function fieldUnit(field: RetentionFieldId): TimeUnit {
    return ready.value ? values.value[field].unit : RETENTION_DEFAULT[field].unit;
  }

  /** Keep the typed text so a rejected value stays visible while the user corrects it. */
  function setDraft(field: RetentionFieldId, text: string) {
    drafts.value = { ...drafts.value, [field]: text };
  }

  /**
   * Persist one field. Invalid input reports an error and sends nothing; an unchanged
   * policy sends nothing either. The saved object is what `hub/history` provides, so an
   * untouched field is never rewritten from display-only values.
   *
   * Refuses while the snapshot is missing: the placeholder config reports both intervals as
   * zero, so a save would silently zero the field the user did not touch.
   */
  function persist(field: RetentionFieldId, interval: TimeInterval) {
    if (retentionLocked.value) return;
    errors.value = { ...errors.value, [field]: '' };
    drafts.value = { ...drafts.value, [field]: undefined };
    const current = history.value ?? RETENTION_DEFAULT;
    const next = withRetentionField(current, field, interval);
    if (isSameRetention(current, next)) return;
    store.dispatch('hub/saveHistorySettings', next);
  }

  /** Commit the typed number with the current unit. */
  function saveValue(field: RetentionFieldId) {
    const interval = parseRetentionInterval(fieldValue(field), fieldUnit(field));
    if (!interval) {
      errors.value = { ...errors.value, [field]: INVALID_MESSAGE };
      return;
    }
    persist(field, interval);
  }

  /** Changing the unit keeps the typed number and re-validates against the new bounds. */
  function saveUnit(field: RetentionFieldId, unit: unknown) {
    if (typeof unit !== 'string') return;
    const interval = parseRetentionInterval(fieldValue(field), unit as TimeUnit);
    if (!interval) {
      errors.value = { ...errors.value, [field]: INVALID_MESSAGE };
      return;
    }
    persist(field, interval);
  }

  /** One in-flight read at a time; the previous request is aborted so a slow reply cannot
   *  overwrite a newer one or write to an unmounted component. */
  let usageRequest: AbortController | null = null;

  async function loadUsage() {
    usageRequest?.abort();
    const controller = new AbortController();
    usageRequest = controller;
    usageLoading.value = true;
    usageError.value = '';
    try {
      const response = await get('/system/storage', { signal: controller.signal });
      if (!response.ok) throw new Error('Storage usage could not be read.');
      usage.value = (await response.json()) as StorageUsage;
    } catch (cause) {
      // An aborted read was superseded or the section left; it must not report an error.
      if (controller.signal.aborted) return;
      usageError.value = (cause as Error).message;
    } finally {
      if (usageRequest === controller) {
        usageRequest = null;
        usageLoading.value = false;
      }
    }
  }

  onMounted(loadUsage);
  onBeforeUnmount(() => {
    usageRequest?.abort();
    usageRequest = null;
  });
</script>

<template>
  <section class="nh-section" aria-labelledby="storage-heading">
    <header class="nh-section-head">
      <h2 id="storage-heading">Storage &amp; retention</h2>
      <p class="nh-help">History drives charts and exports. Retention applies to every device on this hub.</p>
    </header>

    <p v-if="offline" class="nh-alert is-warn" role="status">
      <UiIcon name="offline" />
      <span>Connect to the hub to change retention. Usage below is the last reading.</span>
    </p>
    <p v-else-if="!ready" class="nh-muted" role="status">Loading current retention…</p>

    <UiCard title="Retention" icon="settingsRetention" :heading-level="3">
      <UiSettingRow
        v-for="field in RETENTION_FIELDS"
        :key="field.id"
        :label="field.label"
        :help="field.help"
        :for-id="`retention-${field.id}`">
        <input
          :id="`retention-${field.id}`"
          class="nh-input is-num"
          type="number"
          min="1"
          step="1"
          inputmode="numeric"
          :value="fieldValue(field.id)"
          :disabled="retentionLocked"
          :aria-invalid="!!errors[field.id] || undefined"
          :aria-describedby="errors[field.id] ? `retention-${field.id}-err` : undefined"
          @input="setDraft(field.id, ($event.target as HTMLInputElement).value)"
          @change="saveValue(field.id)" />
        <UiSelect
          :model-value="fieldUnit(field.id)"
          :options="unitOptions"
          :label="`${field.label} unit`"
          :disabled="retentionLocked"
          @update:model-value="saveUnit(field.id, $event)" />
        <template #error>
          <p v-if="errors[field.id]" :id="`retention-${field.id}-err`" class="nh-err" role="alert">
            <UiIcon name="error" />{{ errors[field.id] }}
          </p>
        </template>
      </UiSettingRow>
    </UiCard>

    <UiCard title="Usage" icon="settingsUsage" :heading-level="3">
      <template #tools>
        <UiButton size="sm" variant="ghost" icon="refresh" :loading="usageLoading" @click="loadUsage">Refresh</UiButton>
      </template>
      <p v-if="usageError" class="nh-alert is-danger" role="alert"><UiIcon name="error" /><span>{{ usageError }}</span></p>
      <p v-else-if="usageLoading && !usage" class="nh-muted" role="status">Reading storage usage…</p>
      <p v-else-if="!usage || !usage.entries.length" class="nh-muted">No storage is in use yet.</p>
      <dl v-else class="nh-kv">
        <template v-for="entry in usage.entries" :key="entry.label">
          <dt>{{ entry.label }}</dt>
          <dd>
            {{ formatQuantity(entry.bytes, 'bytes') }}
            <span v-if="entry.retentionDays" class="nh-muted">· {{ entry.retentionDays }} day retention</span>
          </dd>
        </template>
        <dt>Data directory</dt>
        <dd>{{ formatQuantity(usage.dataDirBytes, 'bytes') }}</dd>
        <template v-if="usage.diskTotalBytes">
          <dt>Disk</dt>
          <dd class="nh-muted">
            {{ formatQuantity(usage.diskFreeBytes, 'bytes') }} free of {{ formatQuantity(usage.diskTotalBytes, 'bytes') }}
          </dd>
        </template>
      </dl>
    </UiCard>
  </section>
</template>
