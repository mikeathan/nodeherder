<script setup lang="ts">
  // Both settings pages edit the same policy patches. Vuex owns saved values;
  // derived summaries explain inheritance without creating copied device defaults.
  import { computed, ref, watch } from 'vue';
  import { store } from '@/store';
  import type { DeviceConfig } from '@/types/settings.type';
  import { findDevice } from '@/composables/useHub';
  import { effectiveCategoryInterval, effectiveFieldInterval, effectiveLiveInterval, formatLiveInterval, intervalSeconds, resolveFieldSetting, updateReportingInterval, updateReportingThreshold, validConfirmationTimeout } from '@/domain/reporting-settings';
  import { exposeLabel } from '@/domain/exposes';
  import UiButton from '@/components/ui/UiButton.vue';
  import ReportingRateControl from './ReportingRateControl.vue';
  import ReportingSensorPicker from './ReportingSensorPicker.vue';

  const props = withDefaults(defineProps<{ id?: string; config: Partial<DeviceConfig>; defaults: DeviceConfig }>(), { id: '' });
  const emit = defineEmits<{ (e: 'update', value: Partial<DeviceConfig>): void }>();
  const draftFields = ref<string[]>([]);
  const draftHistory = ref<string[]>([]);
  const error = ref('');
  const errorKey = ref('');
  watch(() => props.id, () => { draftFields.value = []; draftHistory.value = []; error.value = ''; });
  const exposes = computed(() => findDevice(props.id)?.exposes ?? {});
  const categories = [{ name: 'diagnostic', label: 'Diagnostics', help: 'Signal quality, battery readings and other device health measurements.' }, { name: 'measurement', label: 'Measurements', help: 'Distance, temperature, power and other readings used in daily control.' }, { name: 'config', label: 'Configuration', help: 'Read-only numeric configuration reports. Writable confirmations stay immediate.' }];
  const fields = computed<string[]>(() => [...new Set([...draftFields.value, ...Object.keys(props.config.reportingThresholds ?? {}), ...Object.keys(props.config.liveIntervalOverrides ?? {})])]);
  const historyFields = computed<string[]>(() => [...new Set([...draftHistory.value, ...Object.keys(props.config.metricOverrides ?? {}), ...Object.keys(props.config.debounceOverrides ?? {})])]);
  const available = (selected: string[]) => Object.values(exposes.value)
    .filter((e) => e.type === 'numeric' && e.access_mode === 'read' && !selected.includes(e.name))
    .map((e) => ({ name: e.name, label: `${exposeLabel(e.name)}${e.unit ? ` (${e.unit})` : ''}` }));
  const offline = computed(() => store.getters['ws/getConnectionStatus'] !== 'connected');
  const deviceConfig = computed(() => props.id ? props.config as DeviceConfig : undefined);
  const globalConfig = computed(() => props.id ? props.defaults : props.config as DeviceConfig);
  const unit = (name: string) => exposes.value[name]?.unit ?? 'native units';
  const live = (name: string) => effectiveLiveInterval(deviceConfig.value, globalConfig.value, name, String(exposes.value[name]?.category ?? 'measurement'));
  const categoryRule = (category: string, history = false) => effectiveCategoryInterval(deviceConfig.value, globalConfig.value, category, history);
  const categoryInherited = computed(() => props.id ? 'Use shared default' : 'No category rule · every change');
  const historyInterval = (name: string) => effectiveFieldInterval(deviceConfig.value, globalConfig.value, name, String(exposes.value[name]?.category ?? 'measurement'), true);
  const rateHint = (category: string, history = false) => {
    const rule = categoryRule(category, history);
    return `${formatLiveInterval(intervalSeconds(rule.value))} · ${props.id && props.config[history ? 'defaultDebounceByCategory' : 'liveIntervalByCategory']?.[category] ? 'Only this device' : props.id ? 'Shared default' : 'All devices'}`;
  };
  const deltaHint = (name: string) => {
    const rule = resolveFieldSetting(deviceConfig.value?.reportingThresholds, globalConfig.value.reportingThresholds, name, 0, 'No minimum change');
    return rule.value === 0 ? `Every change · ${rule.source}` : `Changes of ${rule.value} ${unit(name)} or more · ${rule.source}`;
  };

  const historyInherited = (name: string) => {
    const rule = resolveFieldSetting(undefined, props.id ? props.defaults.metricOverrides : undefined,
      name, (exposes.value[name]?.category ?? 'measurement') === 'measurement', 'Sensor category');
    return `Use shared rule · ${rule.value ? 'Record' : 'Exclude'}`;
  };

  function saveInterval(key: 'liveIntervalByCategory' | 'liveIntervalOverrides' | 'defaultDebounceByCategory' | 'debounceOverrides', name: string, input: string) {
    errorKey.value = `${key}:${name}`;
    const next = updateReportingInterval(props.config[key] ?? {}, name, input);
    error.value = next ? '' : 'Enter zero or a positive interval, up to one year. Leave blank to inherit.';
    if (next) emit('update', { [key]: next });
  }
  function saveThreshold(name: string, input: string) {
    errorKey.value = `threshold:${name}`;
    const next = updateReportingThreshold(props.config.reportingThresholds ?? {}, name, input);
    error.value = next ? '' : 'Enter 0 or a positive number.';
    if (next) emit('update', { reportingThresholds: next });
  }
  function saveTimeout(input: string) {
    errorKey.value = 'timeout';
    const value = input.trim() ? Number(input) : undefined;
    if (value !== undefined && !validConfirmationTimeout(value)) { error.value = 'Enter a positive whole number of milliseconds within the browser timer range.'; return; }
    error.value = '';
    emit('update', { commandConfirmationTimeoutMs: value });
  }
  function removeField(name: string, history = false) {
    const draft = history ? draftHistory : draftFields;
    draft.value = draft.value.filter((field) => field !== name);
    const patch: Partial<DeviceConfig> = {};
    const keys = history ? ['metricOverrides', 'debounceOverrides'] as const : ['reportingThresholds', 'liveIntervalOverrides'] as const;
    for (const key of keys) {
      const next: Record<string, unknown> = { ...props.config[key] }; delete next[name]; Object.assign(patch, { [key]: next });
    }
    emit('update', patch);
  }
  function saveHistory(name: string, input: string) {
    const next: Record<string, boolean> = { ...props.config.metricOverrides };
    if (input === '') delete next[name];
    else Object.assign(next, { [name]: input === 'true' });
    emit('update', { metricOverrides: next });
  }
  function addField(name: string, history = false) {
    errorKey.value = 'field';
    const selected = history ? historyFields.value : fields.value;
    if (selected.includes(name)) { error.value = 'This sensor already has an adjustment.'; return; }
    if (!name) { error.value = 'Choose a sensor first.'; return; }
    // Adding a row is presentation only. Save only the setting the user edits.
    const draft = history ? draftHistory : draftFields;
    draft.value = [...draft.value, name];
    error.value = '';
  }
</script>

<template>
  <div class="reporting-settings">
    <p v-if="offline" role="status">Connect to the hub to change reporting settings.</p>
    <section class="reporting-card" aria-labelledby="live-heading">
      <h3 id="live-heading">Dashboard &amp; activity updates</h3>
      <p>{{ id ? 'Changes here apply only to this device.' : 'Shared defaults apply to all devices unless they have an override.' }} Automations always use unfiltered readings.</p>
      <div v-for="category in categories" :key="`${id}:${category.name}`" class="reporting-category">
        <div>
          <h4>{{ category.label }}</h4>
          <p>{{ category.help }}</p>
          <small>{{ rateHint(category.name) }}</small>
        </div>
        <ReportingRateControl
          :label="`${category.label} updates`"
          :value="config.liveIntervalByCategory?.[category.name]"
          :inherited="categoryInherited"
          :error="errorKey === `liveIntervalByCategory:${category.name}` ? error : ''"
          :disabled="offline" @change="saveInterval('liveIntervalByCategory', category.name, $event)" />
      </div>
      <small>A custom interval limits update frequency and keeps the latest reading. Individual sensor adjustments take priority. Events and command confirmations stay immediate.</small>
    </section>

    <fieldset class="reporting-card" aria-describedby="field-help">
      <legend>Sensor adjustments</legend>
      <p id="field-help">Ignore small fluctuations in selected sensors. All other sensors keep their current rules.</p>
      <p v-if="!fields.length" class="empty">No individual adjustments. Category rules above still apply.</p>
      <article v-for="name in fields" :key="`${id}:${name}`" class="reporting-field">
        <div class="reporting-heading">
          <h4>{{ exposeLabel(name) }}</h4>
          <UiButton variant="ghost" size="sm" :disabled="offline" @click="removeField(name)">Reset {{ exposeLabel(name) }} adjustment</UiButton>
        </div>
        <label class="threshold">
          <span>Ignore changes smaller than ({{ unit(name) }})</span>
          <input type="number" min="0" step="any" :value="config.reportingThresholds?.[name] ?? ''"
            :aria-label="`Ignore changes smaller than for ${name}`" :aria-invalid="errorKey === `threshold:${name}` && !!error || undefined"
            :placeholder="String(resolveFieldSetting(deviceConfig?.reportingThresholds, globalConfig.reportingThresholds, name, 0, '').value)"
            :disabled="offline" @change="saveThreshold(name, ($event.target as HTMLInputElement).value)" />
          <small v-if="error && errorKey === `threshold:${name}`" role="alert" class="reporting-error">{{ error }}</small>
          <small>{{ deltaHint(name) }}</small>
          <small>0 shows every change. Compared with the last displayed value; leave blank to use the shared rule.</small>
        </label>
        <details class="field-advanced">
          <summary>Advanced update timing · {{ formatLiveInterval(live(name).value) }}</summary>
          <ReportingRateControl :label="`Update timing for ${name}`" :value="config.liveIntervalOverrides?.[name]"
            inherited="Use category or shared sensor rule" :disabled="offline"
            :error="errorKey === `liveIntervalOverrides:${name}` ? error : ''"
            @change="saveInterval('liveIntervalOverrides', name, $event)" />
          <small>{{ formatLiveInterval(live(name).value) }} · {{ live(name).source }}</small>
        </details>
      </article>
      <ReportingSensorPicker :key="`${id}:live-picker`" label="Sensor to adjust" :options="available(fields)" :manual="!id" :disabled="offline" @add="addField($event)" />
    </fieldset>

    <details class="reporting-card history-settings">
      <summary>History recording <span>Chart storage{{ historyFields.length ? ` · ${historyFields.length} sensor ${historyFields.length === 1 ? 'adjustment' : 'adjustments'}` : '' }}</span></summary>
      <p>Separate from dashboard updates. Recording must be enabled for the device. History saves the first eligible change in each interval; skipped changes are not saved later.</p>
      <div v-for="category in categories" :key="`${id}:${category.name}`" class="reporting-category">
        <div><h4>{{ category.label }}</h4><small>{{ rateHint(category.name, true) }}</small></div>
        <ReportingRateControl :label="`${category.label} history timing`" :value="config.defaultDebounceByCategory?.[category.name]"
          :inherited="categoryInherited" :disabled="offline"
          :error="errorKey === `defaultDebounceByCategory:${category.name}` ? error : ''"
          @change="saveInterval('defaultDebounceByCategory', category.name, $event)" />
      </div>
      <p>Measurements are included by default. Diagnostics are excluded unless you choose to record a sensor below.</p>
      <fieldset class="history-fields">
        <legend>History sensor adjustments</legend>
        <article v-for="name in historyFields" :key="`${id}:${name}`" class="reporting-field">
          <div class="reporting-heading">
            <h4>{{ exposeLabel(name) }}</h4>
            <UiButton variant="ghost" size="sm" :disabled="offline" @click="removeField(name, true)">Reset {{ exposeLabel(name) }} history</UiButton>
          </div>
          <div class="reporting-field-grid">
            <label><span>Save to history</span>
              <select :aria-label="`History for ${name}`" :value="config.metricOverrides?.[name] === undefined ? '' : String(config.metricOverrides[name])"
                :disabled="offline" @change="saveHistory(name, ($event.target as HTMLSelectElement).value)">
                <option value="">{{ historyInherited(name) }}</option><option value="true">Record</option><option value="false">Exclude</option>
              </select>
            </label>
            <ReportingRateControl :label="`History timing for ${name}`" :value="config.debounceOverrides?.[name]"
              inherited="Use category or shared sensor rule" :disabled="offline"
              :error="errorKey === `debounceOverrides:${name}` ? error : ''"
              @change="saveInterval('debounceOverrides', name, $event)" />
          </div>
          <small>{{ formatLiveInterval(intervalSeconds(historyInterval(name).value)) }} · {{ historyInterval(name).source }}</small>
        </article>
        <ReportingSensorPicker :key="`${id}:history-picker`" label="History sensor" :options="available(historyFields)" :manual="!id" :disabled="offline" @add="addField($event, true)" />
      </fieldset>
    </details>

    <details class="reporting-card">
      <summary>Advanced <span>Command confirmation</span></summary>
      <p>How long controls wait for a device to confirm a command.</p>
      <label><span>Command confirmation timeout (milliseconds)</span>
        <input type="number" min="1" step="1" :value="config.commandConfirmationTimeoutMs ?? ''"
          aria-label="Command confirmation timeout (milliseconds)" :placeholder="`Inherit · ${defaults.commandConfirmationTimeoutMs ?? ''} ms`"
          :disabled="offline" @change="saveTimeout(($event.target as HTMLInputElement).value)" />
        <small v-if="error && errorKey === 'timeout'" role="alert" class="reporting-error">{{ error }}</small>
        <small>Leave blank to use the shared timeout.</small>
      </label>
    </details>
    <p v-if="error && errorKey === 'field'" role="alert" class="reporting-error">{{ error }}</p>
  </div>
</template>

<style scoped>
  .reporting-settings { display: grid; gap: 1rem; min-width: 0; }
  .reporting-card { min-width: 0; margin: 0; padding: 1.25rem; border: 1px solid var(--nh-border); border-radius: var(--nh-radius-m); background: var(--nh-surface); }
  h3, h4 { margin: 0; font-size: 1rem; font-weight: 600; }
  p { color: var(--nh-text-muted); font-size: .85rem; line-height: 1.6; margin: .65rem 0 1rem; }
  small, .empty, summary span { color: var(--nh-text-muted); font-size: .78rem; line-height: 1.5; }
  .reporting-category { display: grid; grid-template-columns: minmax(0, 1fr) minmax(12rem, 19rem); gap: 1rem; padding: 1rem 0; border-top: 1px solid var(--nh-border); align-items: start; }
  .reporting-category p { margin: .3rem 0; }
  .reporting-heading { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: .5rem; }
  .reporting-field { border-top: 1px solid var(--nh-border); padding: 1rem 0; }
  .threshold { max-width: 28rem; margin-top: .8rem; }
  .field-advanced { margin-top: 1rem; }
  .field-advanced > div { max-width: 28rem; margin-top: .75rem; }
  .field-advanced > small { display: block; margin-top: .5rem; }
  .reporting-field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; margin: .75rem 0; }
  label { display: grid; gap: .4rem; min-width: 0; font-size: .85rem; }
  input, select { width: 100%; min-width: 0; min-height: 2.75rem; padding: .65rem; border: 1px solid var(--nh-border); border-radius: var(--nh-radius-s); background: var(--nh-surface); color: var(--nh-text); font: inherit; }
  legend, summary { font-size: 1rem; font-weight: 600; }
  legend { padding: 0 .35rem; } summary { cursor: pointer; } summary span { font-weight: 400; margin-left: .5rem; }
  .history-fields { border: 0; padding: 0; min-width: 0; }
  .reporting-error { color: var(--nh-danger); }
  @media (max-width: 640px) {
    .reporting-card { padding: 1rem; }
    .reporting-category, .reporting-field-grid { grid-template-columns: minmax(0, 1fr); }
    summary span { display: block; margin: .25rem 0 0 1rem; }
  }
</style>
