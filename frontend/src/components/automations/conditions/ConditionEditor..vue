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
