<script setup lang="ts">
import { ref, watch, PropType, computed } from "vue";
import InputBox from '@/components/input/InputBox.vue';

import Dropdown from "@/components/controls/Dropdown.vue";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";
import Selection from "@/components/input/Selection.vue";
import { createButtons, createSaveDeleteButtonItems } from "@/configs/automation/trigger-dropdown.config";
import { ButtonPanelType, TimePicker } from "@/types/controls.type";
import { TimeSchedule } from "@/types/automation";
import { TimeScheduleTypes } from "@/contracts/automations";
import VueDatePicker from '@vuepic/vue-datepicker';
import { toTimePicker } from "@/contracts/controls";
// import '@vuepic/vue-datepicker/dist/main.css'

const date = ref();
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
  (e: 'delete', schedule: TimeSchedule[]): void,
}>()

const buttonPanelItems: ButtonPanelType[] = createButtons([
  {
    name: "Add",
    click: addSchedule,
    disabled: false
  },
  {
    name: "Save",
    click: save,
    disabled: false
  },
  {
    name: "Clear",
    click: clear,
    disabled: false
  },
])



function addSchedule() {
  schedules.value.push({} as TimeSchedule)
}

function save() {

  console.log("save")
}
function clear() {

  console.log("remove")
}

const time = ref<TimePicker>({ hours: 12, minutes: 34 });

function updateTime(schedule: TimeSchedule, value: any) {
  schedule.startAt = value;
}


</script>
<template>
  <div>

    <h2>Schedules</h2>

    <ButtonPanel :buttons="buttonPanelItems">
    </ButtonPanel>

    <div v-for="schedule in schedules" :key="schedule.name">

      <div class="row">
        <div class="col-sm-4  pe-5">
          <Selection :value="schedule.type" :items="TimeScheduleTypes" position="center">
          </Selection>
        </div>
        <div class="col-sm-4 pt-4">
          <VueDatePicker :model-value="toTimePicker(schedule.startAt)"
            @update:model-value="(e: any) => updateTime(schedule, e)" time-picker model-type="HH:mm"
            placeholder="Enter time" :is-24="true" :esc-close="true" dark />
        </div>
        <div class="col-sm-1 pt-4">
          <span class="fa fa-trash-alt fa-sm"> </span>
        </div>
      </div>
    </div>


  </div>
</template>
@