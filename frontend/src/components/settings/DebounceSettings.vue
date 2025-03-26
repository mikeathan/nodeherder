<script setup lang="ts">
import { PropType, ref, watchEffect } from 'vue';
import Selection from '../input/Selection.vue';
import ExposeSelectionDialog from '../dialogs/ExposeSelectionDialog.vue';
import TimeIntervalEditor from '../controls/TimeInterval.vue';
import { createTimeIntervalFromSeconds } from '@/contracts/settings';
import { watch } from 'vue';
import { DeviceDebounce } from '@/types/settings.type';
import { TimeInterval } from '@/types/types.type';

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

const selectedExpose = ref<string | null>();
const showSelectExposeDialog = ref(false);
const items = ref<DeviceDebounce>(props.value);
const exposeList = ref<string[]>([]);

watchEffect(() => (exposeList.value = Object.keys(items.value)));

watch(
  () => props.value,
  () => {
    if (!props.value) {
      return;
    }
    selectedExpose.value = firstExpose();
  },
  { immediate: true }
);

function removeSelectedExposeDebounce() {
  if (!selectedExpose.value) {
    return;
  }

  delete items.value[selectedExpose.value];
  selectedExpose.value = firstExpose();
  emit('update', items.value);
}

function addNewExposeDebounce(expose: string) {
  if (!expose) {
    return;
  }

  items.value[expose] = createTimeIntervalFromSeconds(1);
  selectedExpose.value = expose;
  emit('update', items.value);
}
function firstExpose(): string | null {
  const keys = Object.keys(props.value);
  return keys.length > 0 ? keys[0] : null;
}

function updateExposeDebounce(debounce: TimeInterval) {
  if (!selectedExpose.value) {
    return;
  }
  items.value[selectedExpose.value] = debounce;
  emit('update', items.value);
}
</script>

<template>
  <div class="grid">
    <div class="col-12 sm:col-10 flex items-center">
      <Selection label="expose" :value="selectedExpose" :items="exposeList"
        @updated="(value: any) => { selectedExpose = value }" :disabled="!exposeList.length" />
      <div class="flex ml-2">
        <Button icon="pi pi-trash" variant="text" rounded size="small" :disabled="!selectedExpose"
          @click="removeSelectedExposeDebounce()" />
        <Button icon="pi pi-plus" variant="text" rounded size="small" @click="showSelectExposeDialog = true" />
      </div>
    </div>
  </div>
  <div v-if="selectedExpose">
    <TimeIntervalEditor :id="id" :value="items[selectedExpose]" @update="updateExposeDebounce" />
  </div>
  <ExposeSelectionDialog :id="props.id" :show="showSelectExposeDialog" @update="addNewExposeDebounce"
    @close="showSelectExposeDialog = false" />
</template>
