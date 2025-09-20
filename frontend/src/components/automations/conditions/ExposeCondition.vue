<script setup lang="ts">
import { ref, computed, watch, PropType } from 'vue';
import { EqualityOperators } from '../../../contracts/automations';
import { AutomationCondition, ExposeCondition } from '../../../types/automation.type.js';
import { store } from '../../../store/index';
import { allExposeFilter } from '@/configs/automation/device.config';
import ExposeDataInput from '../../controls/ExposeDataInput.vue';
import Selection from '../../input/Selection.vue';
import ExposeSelector from '@/components/controls/ExposeSelector.vue';
import TimePicker from '@/components/input/TimePicker.vue';
import { convertTimeToDate, getNowTime, toHourMinuteString } from '@/contracts/controls';

const props = defineProps({
  item: {
    type: Object as PropType<ExposeCondition>,
    default: {} as ExposeCondition,
    required: true,
  },
  id: {
    type: String,
    default: '',
    required: true,
  },
});

// const defaultEndTime = '23:59'
// const hasTimeRange = computed(() => {
//   return condition.value.timeRange != undefined;
// });

const condition = ref<ExposeCondition>({} as ExposeCondition);

const emit = defineEmits<{
  (e: 'update', condition: ExposeCondition): void;
  (e: 'delete'): void;
}>();

watch(
  () => props.item,
  () => {
    condition.value = JSON.parse(JSON.stringify(props.item)) as ExposeCondition;
  },
  { immediate: true }
);

function exposeSelected(value: string): void {
  if (value == '') {
    return;
  }

  condition.value.name = value;
  condition.value.value = '';
  emit('update', condition.value);
}

function operatorUpdated(operator: string): void {
  condition.value.equality = operator;
  emit('update', condition.value);
}

function dataUpdated(value: any): void {
  // cast true/false to boolean
  if (value == 'true' || value == 'false') {
    value = Boolean(value == 'true');
  }
  condition.value.value = value;
  emit('update', condition.value);
}

const exposeOperators = computed(() => {
  if (condition.value.name == '') {
    return [];
  }

  var device = store.getters['hub/findDevice'](props.id);
  if (!device || device.exposes[condition.value.name] == undefined) {
    return [];
  }

  const feature = device.exposes[condition.value.name];
  switch (feature.type) {
    case 'binary':
    case 'enum':
      return Array<string>(EqualityOperators[0]);
    default:
      return EqualityOperators;
  }
});

function hasExposeName(): boolean {
  return condition.value.name != '';
}

</script>

<template>
  <div class="row">
    <div class="col-sm-4">
      <ExposeSelector :id="props.id" :value="condition.name" @updated="exposeSelected" :filter="allExposeFilter()"
        :disabled="hasExposeName()" />
    </div>
    <div class="col-sm-2">
      <Selection :value="condition.equality" @updated="operatorUpdated" :items="exposeOperators"
        :disabled="!hasExposeName()" />
    </div>
    <div class="col-sm-5">
      <ExposeDataInput :id="props.id" :name="condition.name" :value="condition.value" @updated="dataUpdated"
        :disabled="!hasExposeName()" />
    </div>
  </div>
</template>
