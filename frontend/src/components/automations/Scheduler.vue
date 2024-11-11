<script setup lang="ts">
import { ref, watch, PropType, computed } from "vue";
import { getActionType, ActionType, } from "@/contracts/automations"
import { Automation, AutomationTrigger, AutomationTriggerAction, TimeSchedule } from "@/types/automation";
import { emitClosePanel } from "@/mixins/useAutomationsEventBus";
import { PanelComponents } from "@/mixins/usePanelComponents";
import Dropdown from "@/components/controls/Dropdown.vue";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";
import Selection from "@/components/input/Selection.vue";
import { createButtons, createSaveDeleteButtonItems } from "@/configs/automation/trigger-dropdown.config";
import { ButtonPanelType } from "@/types/controls.type";
import VueDatePicker from '@vuepic/vue-datepicker';
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

const time = ref<{ hours: number | string; minutes: number }>({ hours: 12, minutes: 34 });

</script>
<template>
  <div>

    <h2>Schedules</h2>

    <ButtonPanel :buttons="buttonPanelItems">
    </ButtonPanel>

    <VueDatePicker v-model="time" text-input time-picker model-type="HH.mm" placeholder="Enter time" :is-24="true"
      :esc-close="true" dark />

    <div v-for="schedule in schedules" :key="schedule.name">
      {{ schedule.startAt }}
      <VueDatePicker v-model="schedule.startAt" time-picker model-type="HH.mm" placeholder="Enter time" :is-24="true"
        :esc-close="true" dark />
    </div>
  </div>
</template>
@