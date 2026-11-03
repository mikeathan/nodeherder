<script setup lang="ts">
  /*
   * Native select (best on phones). Option values may be any JSON value (numbers, booleans,
   * strings); they are matched by index so their type survives, which keeps automation
   * payloads typed the way the hub expects.
   */
  import { computed } from 'vue';

  export type SelectOption = { value: unknown; label: string; disabled?: boolean };

  const props = defineProps<{
    modelValue: unknown;
    options: readonly SelectOption[];
    placeholder?: string;
    label?: string;
    id?: string;
    invalid?: boolean;
    disabled?: boolean;
  }>();
  const emit = defineEmits<{ (e: 'update:modelValue', value: unknown): void }>();

  const selectedIndex = computed(() => {
    const exact = props.options.findIndex((o) => o.value === props.modelValue);
    if (exact >= 0 || props.modelValue === null || props.modelValue === undefined) return exact;
    // stored values may arrive as strings for numeric options (or the reverse)
    return props.options.findIndex((o) => String(o.value) === String(props.modelValue));
  });

  /** Native menus can report `input` before `change`. The first event owns the choice so
   *  a render between them cannot restore the previous value and drop it. */
  let committedOnInput = false;

  function commitSelection(event: Event) {
    const index = Number((event.target as HTMLSelectElement).value);
    const option = props.options[index];
    if (!option || option.disabled) return;
    emit('update:modelValue', option.value);
  }
  function onInput(event: Event) {
    committedOnInput = true;
    commitSelection(event);
  }
  function onChange(event: Event) {
    // The change paired with a committed input must not re-read the restored value.
    if (committedOnInput) {
      committedOnInput = false;
      return;
    }
    commitSelection(event);
  }
</script>

<template>
  <select
    :id="id"
    class="nh-select"
    :value="selectedIndex"
    :aria-label="label"
    :aria-invalid="invalid || undefined"
    :disabled="disabled"
    @input="onInput"
    @change="onChange">
    <option v-if="selectedIndex < 0" :value="-1" disabled>{{ placeholder ?? 'Choose…' }}</option>
    <option v-for="(opt, i) in options" :key="i" :value="i" :disabled="opt.disabled">{{ opt.label }}</option>
  </select>
</template>
