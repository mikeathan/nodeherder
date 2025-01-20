<script setup lang="ts">
import { computed, ref, PropType } from 'vue';

const emit = defineEmits<{
  (e: 'update', value: any): void;
}>();




const props = defineProps({
  value: { type: Object as PropType<any>, require: true },
  valueOn: { type: Object as PropType<any>, require: true },
  valueOff: { type: Object as PropType<any>, require: true },
});


const checked = computed(() =>
  props.value == props.valueOn
);

const hasValue = computed(
  () => props.value != null || props.value != undefined
);

function valueChanged(event: Event): void {
  emit(
    'update',
    (event.target as HTMLInputElement).checked
      ? props.valueOn
      : props.valueOff
  );
}
</script>
<template>
  <ToggleSwitch v-model="checked" @change="valueChanged" :disabled="!hasValue" />
</template>
