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
    leftLabel: { type: String, default: '' },
    rightLabel: { type: String, default: '' },
    showIcon: { type: Boolean, default: false },
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
  <div class="flex items-center gap-2">
    <span v-if="props.leftLabel" class="text-sm">{{ props.leftLabel }}</span>
    <ToggleSwitch v-model="isChecked" :disabled="!hasValue || props.disabled">
      <template v-if="showIcon" #handle="{ checked }">
        <i :class="['!text-xs pi', { 'pi-check': checked, 'pi-times': !checked }]" />
      </template>
    </ToggleSwitch>
    <span v-if="props.rightLabel" class="text-sm">{{ props.rightLabel }}</span>
  </div>
</template>
