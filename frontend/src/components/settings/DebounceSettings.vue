<script setup lang="ts">
  import { computed, PropType, ref } from 'vue';
  import { DeviceDebounce } from '@/types/settings.type';
  import Selection from '../input/Selection.vue';

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

  const selectedExpose = ref<string | null>(Object.keys(props.value)[0] || null);
  const exposeDebounce = computed(() => {
    return Object.keys(props.value);
  });
  function removeExposeDebounce(expose: string) {
    delete exposeDebounce.value[expose];
    // emit('update', value);
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
          @click="removeExposeDebounce(selectedExpose)" />
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
