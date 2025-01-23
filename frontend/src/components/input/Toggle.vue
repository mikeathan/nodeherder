<script setup lang="ts">
  import { computed, ref, PropType } from 'vue';

  const emit = defineEmits<{
    (e: 'update', value: any): void;
  }>();

  const props = defineProps({
    value: { type: null, require: true },
    valueOn: { type: null, require: true },
    valueOff: { type: null, require: true },
  });

  const hasValue = computed(() => props.value != null || props.value != undefined);

  const isChecked = computed({
    get() {
      return props.value === props.valueOn;
    },
    set(newValue) {
      emit('update', newValue ? props.valueOn : props.valueOff);
    },
  });
</script>
<template>
  <ToggleSwitch v-model="isChecked" :disabled="!hasValue" />
</template>
