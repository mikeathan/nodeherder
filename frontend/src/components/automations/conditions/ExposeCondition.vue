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
  import { convertTimeToDate } from '@/contracts/controls';

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

  const emit = defineEmits<{
    (e: 'update', condition: AutomationCondition): void;
    (e: 'delete'): void;
  }>();

  watch(
    () => props.item,
    () => {
      condition.value = props.item;
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

  const hasTimeRange = computed(() => {
    return condition.value.timeRange != null;
  });

  function onAddTimeRange(): void {
    condition.value.timeRange = {
      startAt: '',
      endAt: '',
    };
  }

  function onRemoveTimeRange(): void {
    condition.value.timeRange = null;
  }
</script>
<style scoped>
  .p-accordion .p-accordion-header {
    height: 10px; /* Adjust the height as needed */
    line-height: 20px; /* Adjust the line height as needed */
  }
</style>
<template>
  <div class="row">
    <div class="col-sm-4">
      <ExposeSelector
        :id="props.id"
        :value="condition.name"
        @updated="exposeSelected"
        :filter="allExposeFilter()"
        :disabled="hasExposeName()" />
    </div>
    <div class="col-sm-2">
      <Selection
        :value="condition.equality"
        @updated="operatorUpdated"
        :items="exposeOperators"
        :disabled="!hasExposeName()" />
    </div>
    <div class="col-sm-5">
      <ExposeDataInput
        :id="props.id"
        :name="condition.name"
        :value="condition.value"
        @updated="dataUpdated"
        :disabled="!hasExposeName()" />
    </div>
    <div class="col-sm-1 d-flex">
      <Button v-if="!hasTimeRange" icon="pi pi-plus-circle" variant="text" rounded small @click="onAddTimeRange" />
      <Button v-else icon="pi pi-minus-circle" variant="text" rounded small @click="onRemoveTimeRange" />
      <Button icon="pi pi-trash" variant="text" rounded small @click="onDelete" />
    </div>
  </div>
  <div v-if="condition.timeRange" class="row">
    <Accordion value="0">
      <AccordionHeader> Time Range </AccordionHeader>
      <AccordionContent>
        <div class="row">
          <div class="col-sm-4">
            <TimePicker :value="convertTimeToDate(condition.timeRange.startAt)" />
          </div>
          <div class="col-sm-4">
            <TimePicker :value="convertTimeToDate(condition.timeRange.endAt)" />
          </div>
        </div>
      </AccordionContent>
    </Accordion>
  </div>
</template>
