<script setup lang="ts">
  import { ref, computed, watch, PropType } from 'vue';
  import { EqualityOperators } from '../../../contracts/automations';
  import { AutomationCondition, ExposeCondition, TimeCondition } from '../../../types/automation.type.js';
  import { store } from '../../../store/index';
  import { allExposeFilter } from '@/configs/automation/device.config';
  import ExposeDataInput from '../../controls/ExposeDataInput.vue';
  import Selection from '../../input/Selection.vue';
  import ExposeSelector from '@/components/controls/ExposeSelector.vue';
  import TimePicker from '@/components/input/TimePicker.vue';
  import { convertTimeToDate, getNowTime, toHourMinuteString } from '@/contracts/controls';

  const props = defineProps({
    item: {
      type: Object as PropType<TimeCondition>,
      default: {} as TimeCondition,
      required: true,
    },
    id: {
      type: String,
      default: '',
      required: true,
    },
  });

  const defaultEndTime = '23:59';
  const hasTimeRange = computed(() => {
    return condition.value.timeRange != undefined;
  });

  const condition = ref<TimeCondition>({} as TimeCondition);
  const startTimeIsEnabled = ref(false);
  const endTimeIsEnabled = ref(false);

  const emit = defineEmits<{
    (e: 'update', condition: TimeCondition): void;
    (e: 'delete'): void;
  }>();

  watch(
    () => props.item,
    () => {
      condition.value = JSON.parse(JSON.stringify(props.item)) as TimeCondition;
      startTimeIsEnabled.value = condition.value.timeRange != undefined;
      endTimeIsEnabled.value = condition.value.timeRange != undefined;
    },
    { immediate: true }
  );


  function getStartAtTime(): Date {
    if (!condition.value.timeRange) {
      return getNowTime();
    }
    return convertTimeToDate(condition.value.timeRange?.startAt);
  }

  function getEndAtTime(): Date {
    if (!condition.value.timeRange) {
      return convertTimeToDate(defaultEndTime);
    }
    const endAt = condition.value.timeRange?.endAt == '' ? defaultEndTime : condition.value.timeRange?.endAt;
    return convertTimeToDate(endAt);
  }

  function updateStartAtTime(startAtTime: Date) {
    condition.value.timeRange!.startAt = toHourMinuteString(startAtTime);
    emit('update', condition.value);
  }

  function validateStartTime(startAtTime: Date): string | null {
    const endAtTime = getEndAtTime();
    const startHour = startAtTime.getHours();
    const endHour = endAtTime.getHours();
    const startMinutes = startAtTime.getMinutes();
    const endMinutes = endAtTime.getMinutes();

    if (startHour > endHour) {
      endTimeIsEnabled.value = false;
      return 'Start time must be before end time';
    }

    if (startHour === endHour) {
      if (startMinutes >= endMinutes) {
        endTimeIsEnabled.value = false;
        return 'Start time must be before end time';
      }
    }

    endTimeIsEnabled.value = true;
    return null;
  }

  function validateEndTime(endAtTime: Date): string | null {
    const startAtTime = getStartAtTime();
    const startHour = startAtTime.getHours();
    const endHour = endAtTime.getHours();
    const startMinutes = startAtTime.getMinutes();
    const endMinutes = endAtTime.getMinutes();

    if (endHour < startHour) {
      startTimeIsEnabled.value = false;
      return 'End time must be after start time';
    }

    if (endHour === startHour) {
      if (endMinutes <= startMinutes) {
        startTimeIsEnabled.value = false;

        return 'End time must be after start time';
      }
    }

    startTimeIsEnabled.value = true;
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
      <TimePicker
        label="Start At"
        :value="getStartAtTime()"
        :disabled="!startTimeIsEnabled"
        @updated="(e) => updateStartAtTime(e)"
        :validation="validateStartTime" />
    </div>
    <div class="col-sm-2">
      <TimePicker
        label="End At"
        :value="getEndAtTime()"
        :disabled="!endTimeIsEnabled"
        @updated="(e) => updateEndAtTime(e)"
        :validation="validateEndTime" />
    </div>   
  </div>
</template>
