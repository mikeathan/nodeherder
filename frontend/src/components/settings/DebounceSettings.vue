<script setup lang="ts">
  import { computed, PropType, ref } from 'vue';
  import { DeviceDebounce } from '@/types/settings.type';
  import Selection from '../input/Selection.vue';
  import { store } from '../../store/index';

  const props = defineProps({
    id: {
      type: String,
      required: true,
    },
    value: {
      type: Object as PropType<DeviceDebounce>,
      required: true,
    },
  });
  const emit = defineEmits<{
    (e: 'update', value: DeviceDebounce): void;
  }>();

  const items = ref<DeviceDebounce>(props.value);
  const selectedExpose = ref<string | null>(Object.keys(items.value)[0]);
  const exposeDebounce = computed(() => {
    return Object.keys(props.value);
  });
  function removeSelectedExposeDebounce() {
    if (!selectedExpose.value) {
      return;
    }

    delete items.value[selectedExpose.value];
    selectedExpose.value = null;
    emit('update', items.value);
  }
</script>

<template>
  <div class="grid">
    <div class="col-12 sm:col-10 flex items-center">
      <Selection label="expose" :value="selectedExpose" text="Expose" :items="exposeDebounce" />
      <div class="flex ml-2">
        <Button
          icon="pi pi-trash"
          variant="text"
          rounded
          size="small"
          :disabled="!selectedExpose"
          @click="removeSelectedExposeDebounce()" />
        <Button icon="pi pi-plus" variant="text" rounded size="small" />
      </div>
    </div>
  </div>
  <div v-if="selectedExpose">
    <TimeInterval :id="id" :value="value[selectedExpose]" />
  </div>
</template>

<!-- <Button icon="pi pi-trash" variant="text" rounded @click="removeTriggerExpose(expose)" />
      </div>
    </div>
  </div>

  <div class="pt-4 flex align-items-center justify-content-center">
    <Button
      style="width: 99%"
      icon="pi pi-plus"
      label="Add Expose"
      @click="addNewExpose()"
      text
      size="small"
      :disabled="action.id == ''" /> -->
