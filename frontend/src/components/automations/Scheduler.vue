<script setup lang="ts">
import { ref, watch, PropType, computed } from "vue";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";
import Selection from "@/components/input/Selection.vue";
import { createButtons } from "@/configs/automation/trigger-dropdown.config";
import { TimeSchedule } from "@/types/automation";
import { TimeScheduleTypes } from "@/contracts/automations";
import TimePicker from "@/components/input/TimePicker.vue";
import { toTimePicker } from "@/contracts/controls";
import { emitClosePanel } from "@/mixins/useAutomationsEventBus";

const props = defineProps({
  schedules: {
    type: Object as PropType<TimeSchedule[]>,
    default: [] as TimeSchedule[],
    required: true,
  },
});

const schedules = ref<TimeSchedule[]>({} as TimeSchedule[])

watch(
  () => props.schedules,
  () => {
    // make a deep copy to make it not reactive
    schedules.value = JSON.parse(JSON.stringify(props.schedules)) as TimeSchedule[]
  }, { immediate: true }
)

const emit = defineEmits<{
  (e: 'save', schedules: TimeSchedule[]): void,
}>()

const buttonPanelItems = computed(() => {

  const canClear = schedules.value.length != 0;
  const canAdd = schedules.value.length != TimeScheduleTypes.length;
  // TODO
  // const isModified = props.schedules.length != schedules.value.length;
  // const isValid = isModified && canClear && schedules.value.every(k => k.startAt != '' && k.type != undefined);

  return createButtons([
    {
      name: "Save",
      click: save,
      disabled: false
    },
    {
      name: "Add",
      click: addSchedule,
      disabled: !canAdd,
    },
    {
      name: "Clear",
      click: clear,
      disabled: !canClear
    }
  ]);
});

function addSchedule() {
  schedules.value.push({} as TimeSchedule)
}

function save() {
  emit('save', schedules.value)
  emitClosePanel('Scheduler')
}

function clear() {
  schedules.value = []
}

function updateStartAtTime(schedule: TimeSchedule, value: any) {
  schedule.startAt = value;
}

function removeSchedule(schedule: TimeSchedule) {
  schedules.value = schedules.value.filter(x => x.startAt != schedule.startAt && x.type != schedule.type);
}
function updateType(schedule: TimeSchedule, value: any) {
  schedule.type = value;
}

</script>
<template>
  <div>

    <h2>Schedules</h2>

    <ButtonPanel :buttons="buttonPanelItems">
    </ButtonPanel>

    <div v-for="schedule in schedules" :key="schedule.type">

      <div class="row ">
        <div class="col-sm-4 pt-3">
          <Selection :value="schedule.type" :items="TimeScheduleTypes" position="center"
            @updated="(t) => updateType(schedule, t)">
          </Selection>
        </div>
        <div class="col-sm-4 ">
          <TimePicker :value="schedule.startAt" @updated="(e) => updateStartAtTime(schedule, e)" />
          <!-- 
          <v-text-field v-model="schedule.startAt" label="Picker in menu" prepend-icon="mdi-clock-time-four-outline" readonly>

            <v-time-picker v-model="schedule.startAt" />

          </v-text-field> -->


        </div>
        <div class="col-sm-1 pt-4">
          <span class="fa fa-trash-alt fa-sm" @click="removeSchedule(schedule)"> </span>
        </div>
      </div>
    </div>
  </div>
</template>
<!-- 
<VueDatePicker :model-value="toTimePicker(schedule.startAt)"
            @update:model-value="(e: any) => updateStartAtTime(schedule, e)" time-picker model-type="HH:mm"
            placeholder="Enter time" :is-24="true" :esc-close="true" dark /> -->