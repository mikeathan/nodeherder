<script setup lang="ts">
  import { KeyValuePair } from '@/types/types.type';
  import { PropType, ref, watch } from 'vue';

  const props = defineProps({
    name: String,
    value: {
      type: null,
      required: true,
    },
    items: {
      type: Object as PropType<KeyValuePair<any>>,
      required: true,
    },
  });

  const emit = defineEmits<{
    (e: 'update', value: any): void;
  }>();

  const selectedValue = ref<any>(null);

  watch(
    () => props.value,
    () => {
      selectedValue.value = props.value;
    },
    { immediate: true }
  );

  function selectionChanged(value: any) {
    selectedValue.value = typeof props.value == 'number' ? parseInt(value) : value;
    emit('update', selectedValue.value);
  }

  function getID() {
    return new Date().getTime();
  }
</script>

<template>
  <Button
    v-for="(key, value) in props.items"
    :key="key"
    :label="value as string"
    size="small"
    @click="selectionChanged(key)"
    :class="{
      'p-button-primary': selectedValue === key,
      'p-button-outlined': selectedValue !== key,
    }" />
</template>
