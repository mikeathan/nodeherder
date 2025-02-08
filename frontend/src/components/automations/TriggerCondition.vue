<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { EqualityOperators } from '../../contracts/automations';
import { AutomationCondition } from '../../types/automation.type.js';
import { store } from '../../store/index';
import { allExposeFilter } from '@/configs/automation/device.config';
import ExposeDataInput from '../controls/ExposeDataInput.vue';
import Selection from '../input/Selection.vue';
import ExposeSelector from '@/components/controls/ExposeSelector.vue';

const props = defineProps({
  id: {
    type: String,
    default: '',
    required: true,
  },
  name: {
    type: String,
    default: '',
  },
  operator: {
    type: String,
    default: '',
  },
  data: null,
});

const data = ref<any>(null);
const operator = ref<string>('');
const name = ref<string>('');

const emit = defineEmits<{
  (e: 'update:name', name: string): void;
  (e: 'update:value', property: any): void;
  (e: 'update:operator', data: string): void;
  (e: 'update', condition: AutomationCondition): void;
}>();

watch(
  () => props.name,
  () => {
    name.value = props.name;
  },
  { immediate: true }
);

watch(
  () => props.data,
  () => {
    data.value = props.data;
  },
  { immediate: true }
);

watch(
  () => props.operator,
  () => {
    operator.value = props.operator;
  },
  { immediate: true }
);

function exposeSelected(value: string): void {
  if (value == '') {
    return;
  }

  name.value = value;
  data.value = '';
  emit('update:name', value);
}

function operatorUpdated(event: string): void {
  operator.value = event;
  emit('update:operator', event);
}

function dataUpdated(value: any): void {
  // cast true/false to boolean
  if (value == 'true' || value == 'false') {
    value = Boolean(value == 'true');
  }
  data.value = value;
  emit('update:value', value);
}

const exposeOperators = computed(() => {
  if (name.value == '') {
    return [];
  }

  var device = store.getters['hub/findDevice'](props.id);
  if (!device || device.exposes[name.value] == undefined) {
    return [];
  }

  const feature = device.exposes[name.value];
  switch (feature.type) {
    case 'binary':
    case 'enum':
      return Array<string>(EqualityOperators[0]);
    default:
      return EqualityOperators;
  }
});
</script>

<template>
  <div class="row">
    <div class="col-sm-4">
      <ExposeSelector :id="props.id" :value="name" @updated="exposeSelected" :filter="allExposeFilter()"
        :disabled="name != ''" />
    </div>
    <div class="col-sm-3">
      <Selection :value="operator" @updated="operatorUpdated" :items="exposeOperators" :disabled="name == ''" />
    </div>
    <div class="col-sm-5">
      <ExposeDataInput :id="props.id" :name="name" :value="data" @updated="dataUpdated" :disabled="name == ''" />
    </div>
  </div>
</template>
