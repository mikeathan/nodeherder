<script setup lang="ts">
  import { ref, watch, PropType, computed } from 'vue';
  import ButtonPanel from '@/components/controls/ButtonPanel.vue';
  import Selection from '@/components/input/Selection.vue';
  import { createButtons } from '@/configs/automation/trigger-dropdown.config';
  import { TimeSchedule } from '@/types/automation.type.js';
  import { TimeScheduleTypes } from '@/contracts/automations';
  import TimePicker from '@/components/input/TimePicker.vue';
  import { convertTimeToDate, toHourMinuteString, getNowTimeString } from '@/contracts/controls';
  import { emitClosePanel } from '@/mixins/useAutomationsEventBus';

  const props = defineProps({
    schedules: {
      type: Object as PropType<TimeSchedule[]>,
      default: [] as TimeSchedule[],
      required: true,
    },
  });

  const schedules = ref<TimeSchedule[]>([] as TimeSchedule[]);

  watch(
    () => props.schedules,
    () => {
      if (props.schedules) {
        // make a deep copy to make it not reactive
        schedules.value = JSON.parse(JSON.stringify(props.schedules)) as TimeSchedule[];
      }
    },
    { immediate: true }
  );

  const emit = defineEmits<{
    (e: 'save', schedules: TimeSchedule[]): void;
  }>();

  const buttonPanelItems = computed(() => {
    const canClear = schedules.value.length != 0;
    const canAdd = schedules.value.length != TimeScheduleTypes.length;
    // TODO
    // const isModified = props.schedules.length != schedules.value.length;
    // const isValid = isModified && canClear && schedules.value.every(k => k.startAt != '' && k.type != undefined);

    return createButtons([
      {
        label: 'Save',
        command: save,
        disabled: false,
      },
      {
        label: 'Add',
        command: addSchedule,
        disabled: !canAdd,
      },
      {
        label: 'Clear',
        command: clear,
        disabled: !canClear,
      },
    ]);
  });

  function addSchedule() {
    const newSchedule: TimeSchedule = {
      startAt: getNowTimeString(),
      type: 'enable',
    };

    schedules.value.push(newSchedule);
  }

  function save() {
    emit('save', schedules.value);
    emitClosePanel('Scheduler');
  }

  function clear() {
    schedules.value = [];
  }

  function updateStartAtTime(schedule: TimeSchedule, value: Date) {
    schedule.startAt = toHourMinuteString(value);
  }

  function removeSchedule(schedule: TimeSchedule) {
    schedules.value = schedules.value.filter((x) => x.startAt != schedule.startAt && x.type != schedule.type);
  }
  function updateType(schedule: TimeSchedule, value: any) {
    schedule.type = value;
  }
</script>

<style scoped>
  @media (max-width: 768px) {
    .p-button {
      width: 100%;
      text-align: center;
    }
  }
</style>
<template>
  <div>
    <h2>Schedules</h2>
    <div class="pt-3"></div>
    <ButtonPanel :buttons="buttonPanelItems" severity="secondary" />
    <div class="pt-3"></div>
    <div v-for="schedule in schedules" :key="schedule.type">
      <div class="grid">
        <div class="col-12 sm:col-4">
          <Selection :value="schedule.type" :items="TimeScheduleTypes" @updated="(t) => updateType(schedule, t)" />
        </div>
        <div class="col-12 sm:col-4">
          <TimePicker :value="convertTimeToDate(schedule.startAt)" @updated="(e) => updateStartAtTime(schedule, e)" />
        </div>
        <div class="col-12 sm:col-1">
          <Button icon="pi pi-trash" variant="text" rounded @click="removeSchedule(schedule)" class="delete-button" />
        </div>
      </div>
    </div>
  </div>
</template>
