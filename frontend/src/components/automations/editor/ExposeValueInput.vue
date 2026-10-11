<script setup lang="ts">
  /*
   * Value for an expose, typed the way the hub expects: binary and enum values come from the
   * expose's own values, numbers stay numbers, anything else is text.
   */
  import { computed } from 'vue';
  import { Expose } from '@/types/device';
  import { ExposeTypes } from '@/types/device.type';
  import { exposePresets, formatExposeValue } from '@/domain/exposes';
  import UiSelect from '@/components/ui/UiSelect.vue';

  const props = defineProps<{ expose: Expose | undefined; label: string; id?: string; invalid?: boolean; disabled?: boolean }>();
  const value = defineModel<unknown>({ required: true });

  const binaryOptions = computed(() => {
    const e = props.expose!;
    const values = e.values && 'on' in e.values ? [e.values['on'], e.values['off']] : [true, false];
    return values.map((v) => ({ value: v, label: formatExposeValue(e, v) }));
  });
  const presets = computed(() => (props.expose ? exposePresets(props.expose) : []));
  const unit = computed(() => (props.expose?.unit && props.expose.unit !== 'lqi' ? props.expose.unit : ''));

  function setNumber(event: Event) {
    const raw = (event.target as HTMLInputElement).value;
    value.value = raw === '' ? null : Number(raw);
  }
</script>

<template>
  <UiSelect v-if="expose?.type === ExposeTypes.Binary" :id="id" v-model="value" :options="binaryOptions" :label="label" :invalid="invalid" :disabled="disabled" placeholder="Choose a value" />
  <UiSelect
    v-else-if="expose?.type === ExposeTypes.Enum"
    :id="id"
    v-model="value"
    :options="presets.map((p) => ({ value: p.value, label: p.label }))"
    :label="label"
    :invalid="invalid"
    :disabled="disabled"
    placeholder="Choose a value" />
  <span v-else-if="expose?.type === ExposeTypes.Numeric" class="nh-num">
    <span class="nh-unit-input">
      <input
        :id="id"
        type="number"
        step="any"
        inputmode="decimal"
        class="nh-input is-num"
        :value="typeof value === 'number' ? value : ''"
        :aria-label="label"
        :aria-invalid="invalid || undefined"
        :disabled="disabled"
        @input="setNumber" />
      <em v-if="unit">{{ unit }}</em>
    </span>
    <UiSelect
      v-if="presets.length"
      :model-value="null"
      :options="presets.map((p) => ({ value: p.value, label: p.label }))"
      :label="`${label} preset`"
      placeholder="Preset"
      :disabled="disabled"
      @update:model-value="value = $event" />
  </span>
  <input v-else :id="id" v-model="value" class="nh-input" :aria-label="label" :aria-invalid="invalid || undefined" :disabled="disabled" placeholder="Value" />
</template>

<style scoped>
  .nh-num {
    display: inline-flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    min-width: 0;
  }
  .nh-unit-input {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
  }
  .nh-unit-input em {
    font-style: normal;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
</style>
