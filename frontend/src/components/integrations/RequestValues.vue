<script setup lang="ts">
  import { computed } from 'vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import { fixedValue, valueSource } from '@/domain/integrations';
  import type { ConnectionValue } from '@/types/integrations.type';
  const props = defineProps<{
    modelValue: Record<string, ConnectionValue>;
    choices: { value: string; label: string }[];
    title: string;
    pathMode?: boolean;
  }>();
  const emit = defineEmits<{ 'update:modelValue': [value: Record<string, ConnectionValue>] }>();
  const rows = computed(() => Object.entries(props.modelValue));
  function update(key: string, value: ConnectionValue) {
    emit('update:modelValue', { ...props.modelValue, [key]: value });
  }
  function rename(old: string, name: string) {
    if (old === name || !name || name in props.modelValue) return;
    emit('update:modelValue', Object.fromEntries(rows.value.map(([key, value]) => [key === old ? name : key, value])));
  }
  function remove(key: string) {
    emit('update:modelValue', Object.fromEntries(rows.value.filter(([name]) => key !== name)));
  }
  function add() {
    let i = 1;
    while (`parameter_${i}` in props.modelValue) i++;
    update(`parameter_${i}`, '');
  }
  const literal = (v: ConnectionValue) => (typeof v === 'object' ? '' : String(v ?? ''));
  const kind = (v: ConnectionValue) => (v === null ? 'null' : typeof v === 'object' ? 'string' : typeof v);
</script>
<template>
  <fieldset class="connection-form">
    <legend>{{ title }}</legend>
    <p v-if="!rows.length" class="connection-muted">No parameters.</p>
    <div v-for="([key, value], index) in rows" :key="key" class="connection-form">
      <div class="connection-row">
        <p v-if="pathMode" class="connection-muted">Path segment {{ index + 1 }}</p>
        <label v-else class="connection-field"
          ><span>Provider field {{ index + 1 }}</span
          ><input :value="key" @change="rename(key, ($event.target as HTMLInputElement).value)" /></label
        ><label class="connection-field"
          ><span>Use this value</span
          ><select
            :value="valueSource(value)"
            @change="
              update(
                key,
                ($event.target as HTMLSelectElement).value ? { ref: ($event.target as HTMLSelectElement).value } : ''
              )
            ">
            <option value="">A fixed value</option>
            <option v-for="choice in choices" :key="choice.value" :value="choice.value">{{ choice.label }}</option>
            <option
              v-if="valueSource(value) && !choices.some((c) => c.value === valueSource(value))"
              :value="valueSource(value)">
              Unavailable saved value — choose another
            </option>
          </select></label
        >
      </div>
      <div v-if="!valueSource(value)" class="connection-row">
        <label class="connection-field"
          ><span>Value type</span
          ><select
            :value="kind(value)"
            @change="update(key, fixedValue(literal(value), ($event.target as HTMLSelectElement).value))">
            <option value="string">Text</option>
            <option value="number">Number</option>
            <option value="boolean">True or false</option>
            <option value="null">Empty value</option>
          </select></label
        ><label class="connection-field"
          ><span>Fixed value</span
          ><select
            v-if="typeof value === 'boolean'"
            :value="String(value)"
            @change="update(key, ($event.target as HTMLSelectElement).value === 'true')">
            <option value="true">True</option>
            <option value="false">False</option></select
          ><span v-else-if="value === null" class="connection-muted">JSON null will be sent.</span
          ><input
            v-else
            :value="literal(value)"
            @input="update(key, fixedValue(($event.target as HTMLInputElement).value, kind(value)))"
        /></label>
      </div>
      <div>
        <UiButton variant="ghost" @click="remove(key)">Remove parameter {{ index + 1 }}</UiButton>
      </div>
    </div>
    <div><UiButton @click="add">Add parameter</UiButton></div>
  </fieldset>
</template>
