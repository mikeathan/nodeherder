<script setup lang="ts">
  import { KeyValuePair } from '@/types/types.type';
  import { PropType, computed, ref, watch } from 'vue';

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
    disabled: {
      type: Boolean,
      default: false,
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

  const sortedOptions = computed(() => {
    if (!props.items) return []; 
    const options = Object.entries(props.items).map(([label, value]) => ({ label, value }));
    return options.sort((a, b) => {
      // Sort logic
      if (typeof a.value === 'number' && typeof b.value === 'number') {
        return a.value - b.value;
      }
      return String(a.value).localeCompare(String(b.value));
    });
  });
</script>

<template>
  <div class="flex flex-wrap gap-2">
    <Button
      v-for="item in sortedOptions"
      :key="item.value"
      :label="item.label"
      :disabled="props.disabled"
      size="small"
      @click="selectionChanged(item.value)"
      :class="{
        'p-button-primary': selectedValue === item.value,
        'p-button-outlined': selectedValue !== item.value,
      }" />
  </div>
</template>
