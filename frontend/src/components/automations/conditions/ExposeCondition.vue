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

const condition = ref<ExposeCondition>({} as ExposeCondition);

const hasTimeRange = computed(() => {
  return condition.value.timeRange != undefined;
});

const isExpanded = ref(false);
function toggleExpanded(): void {
  isExpanded.value = !isExpanded.value;
}
const emit = defineEmits<{
  (e: 'update', condition: ExposeCondition): void;
  (e: 'delete'): void;
}>();

watch(
  () => props.item,
  () => {
    condition.value = JSON.parse(JSON.stringify(props.item)) as ExposeCondition;
    isExpanded.value = condition.value.timeRange != undefined; // on initial load, expand if timerange set
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

function onDelete(): void {
  emit('delete');
}

function onAddTimeRange(): void {
  condition.value.timeRange = {
    startAt: toHourMinuteString(new Date()),
    endAt: toHourMinuteString(convertTimeToDate('00:00')),
  };

  emit('update', condition.value);
}

function onRemoveTimeRange(): void {
  condition.value.timeRange = undefined;
  emit('update', condition.value);
}

function getStartAtTime(): Date {
  if (!condition.value.timeRange) {
    return getNowTime();
  }
  return convertTimeToDate(condition.value.timeRange?.startAt);
}

function getEndAtTime(): Date {
  if (!condition.value.timeRange) {
    return convertTimeToDate('00:00');
  }
  const endAt = condition.value.timeRange?.endAt == '' ? '00:00' : condition.value.timeRange?.endAt;
  return convertTimeToDate(endAt);
}

function updateStartAtTime(startAtTime: Date) {
  condition.value.timeRange!.startAt = toHourMinuteString(startAtTime);
  emit('update', condition.value);
}

function validateEndTime(endAtTime: Date): string | null {
  const startAtTime = getStartAtTime();
  const startHour = startAtTime.getHours();
  const endHour = endAtTime.getHours();
  const startMinutes = startAtTime.getMinutes();
  const endMinutes = endAtTime.getMinutes();

  const startTotalMinutes = startHour * 60 + startMinutes;
  const endTotalMinutes = endHour * 60 + endMinutes;

  if (endTotalMinutes < startTotalMinutes) {
    return 'End time must be after start time.';
  }
  return null;
}

function updateEndAtTime(endAtTime: Date) {
  condition.value.timeRange!.endAt = toHourMinuteString(endAtTime);
  emit('update', condition.value);
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
    <div class="col-sm-1 d-flex">
      <Button :icon="isExpanded ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" variant="text" rounded small
        @click="toggleExpanded" />
      <Button icon="pi pi-trash" variant="text" rounded small @click="onDelete" />
    </div>
  </div>
  <div v-if="isExpanded" class="row mt-2">
    <div class="col-sm-4">
      <label class="col-form-label">Time Range Activation:</label>
      <div class="row mt-1">
        <div class="col-sm-4">
          <TimePicker label="Start At" :value="getStartAtTime()" :disabled="!hasTimeRange"
            @updated="(e) => updateStartAtTime(e)" />
        </div>
        <div class="col-sm-4">
          <TimePicker label="End At" :value="getEndAtTime()" :disabled="!hasTimeRange"
            @updated="(e) => updateEndAtTime(e)" :validation="validateEndTime" />
        </div>
        <div class="col-sm-4">
          <Button v-if="!hasTimeRange" icon="pi pi-plus-circle" variant="text" rounded small @click="onAddTimeRange" />
          <Button v-else icon="pi pi-minus-circle" variant="text" rounded small @click="onRemoveTimeRange" />
        </div>
      </div>
    </div>
  </div>
</template>
