<script setup lang="ts">
  import { computed, PropType } from 'vue';

  const emit = defineEmits<{
    (e: 'update', value: any): void;
  }>();

  const props = defineProps({
    value: { type: null, require: true },
    valueOn: { type: null, require: true },
    valueOff: { type: null, require: true },
    disabled: { type: Boolean, default: false },
    label: { type: String, default: '' },
  });

  const hasValue = computed(() => props.value != null || props.value != undefined || props.disabled);

  const isChecked = computed({
    get() {
      return props.value === props.valueOn;
    },
    set(newValue) {
      emit('update', newValue ? props.valueOn : props.valueOff);
    },
  });
</script>
<style scoped></style>
<template>
  <ToggleSwitch v-model="isChecked" :disabled="!hasValue || props.disabled" />
  <span v-if="props.label" class="text-sm">{{ props.label }}</span>
</template>
