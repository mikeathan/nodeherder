<script setup lang="ts">
import { PropType } from 'vue';
import { AutomationCondition } from '../../../types/automation.type.js';
import { ConditionComponents } from '@/mixins/useConditionComponents';

const props = defineProps({

  id: {
    type: String,
    default: '',
    required: true,
  },
  item: {
    type: Object as PropType<AutomationCondition>,
    default: {} as AutomationCondition,
    required: true,
  },
});


const emit = defineEmits<{
  (e: 'update', condition: AutomationCondition): void;
}>();

function update(condition: AutomationCondition): void {
  emit('update', condition);
}
</script>

<template>
  <component :is="ConditionComponents[props.item.type]" v-bind="{
    id: props.id,
    item: props.item,
  }" @update="update" />
</template>


<!-- 
<script setup lang="ts">
import { ref } from 'vue'
import draggable from 'vuedraggable'
import ConditionEditor from './ConditionEditor.vue'
import { AutomationCondition } from '@/types/automation.type'

// this would normally come from props or Vuex/pinia
const conditions = ref<AutomationCondition[]>([
  { id: 'cond-1', type: 'expose', name: '', equality: '', value: '' },
  { id: 'cond-2', type: 'expose', name: '', equality: '', value: '' }
])

function updateCondition(index: number, updated: AutomationCondition) {
  conditions.value[index] = updated
}
</script>

<template>
  <draggable
    v-model="conditions"
    item-key="id"
    handle=".drag-handle"
    class="flex flex-col gap-2"
    ghost-class="drag-ghost"
  >
    <template #item="{ element, index }">
      <div class="flex items-start gap-2 border rounded p-2 bg-surface-50">
        <!-- Drag handle -->
        <i class="pi pi-bars drag-handle cursor-move mt-2 text-gray-500"></i>

        <!-- Your generic condition editor -->
        <ConditionEditor
          :id="element.id"
          :item="element"
          @update="(c) => updateCondition(index, c)"
        />
      </div>
    </template>
  </draggable>
</template>

<style scoped>
.drag-ghost {
  opacity: 0.4;
}
</style> -->