<script setup lang="ts">
  /*
   * Logger settings (spec 059 FR-03): remote log streaming and the hub's minimum level.
   * Saves send the complete `LoggerSettingsType` through `hub/saveLoggerSettings`; the
   * Vuex getter object is never mutated.
   */
  import { computed } from 'vue';
  import { store } from '@/store';
  import type { LoggerSettingsType } from '@/types/settings.type';
  import { LOGGER_LEVELS, withLoggerSetting } from '@/domain/settings/policy';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiSelect from '@/components/ui/UiSelect.vue';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import UiSettingRow from '@/components/ui/UiSettingRow.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const settings = computed<LoggerSettingsType | undefined>(() => store.getters['hub/logger']());
  const offline = computed(() => store.getters['ws/getConnectionStatus'] !== 'connected');
  const levelOptions = LOGGER_LEVELS.map((level) => ({ value: level, label: level }));

  function save<K extends keyof LoggerSettingsType>(key: K, value: LoggerSettingsType[K]) {
    if (settings.value?.[key] === value) return;
    store.dispatch('hub/saveLoggerSettings', withLoggerSetting(settings.value, key, value));
  }
</script>

<template>
  <section class="nh-section" aria-labelledby="logger-heading">
    <header class="nh-section-head">
      <h2 id="logger-heading">Logger</h2>
      <p class="nh-help">Live hub logs also appear in the Console screen.</p>
    </header>

    <p v-if="offline" class="nh-alert is-warn" role="status">
      <UiIcon name="offline" />
      <span>Connect to the hub to change logger settings.</span>
    </p>

    <UiCard title="Hub logs" icon="settingsLogger" :heading-level="3">
      <UiSettingRow
        label="Send hub logs to this browser"
        help="Streams every message at or above the chosen level while this tab is open.">
        <UiToggle
          :model-value="settings?.enableRemoteLogger === true"
          label="Send hub logs to this browser"
          :disabled="offline"
          @update:model-value="save('enableRemoteLogger', $event)" />
      </UiSettingRow>
      <UiSettingRow
        label="Minimum level"
        help="Lower levels are detailed but noisy; higher levels only report problems."
        for-id="logger-level">
        <UiSelect
          id="logger-level"
          :model-value="settings?.level ?? 'info'"
          :options="levelOptions"
          label="Minimum level"
          :disabled="offline"
          @update:model-value="save('level', String($event))" />
      </UiSettingRow>
    </UiCard>
  </section>
</template>
