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
import { add } from "date-fns";


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

const buttonPanelItems = computed(() => {
  return createButtons([
    {
      name: "Save",
      click: remove,
      disabled: false
    },
    {
      name: "Delete",
      click: remove,
      disabled: false
    },

  ])


});

function addSchedule(event: Event) {
  schedules.value.push({} as TimeSchedule)

}

function remove() {

  console.log("remove")
}


</script>
<template>
  <div>

    <h2>Schedules</h2>

    <ButtonPanel :buttons="buttonPanelItems">
      <button type="button" class="btn btn-light" @click="addSchedule">Add schedule</button>
    </ButtonPanel>

    <div v-for="schedule in schedules" :key="schedule.name">
      {{ schedule.name }} - {{ schedule.startAt }} - {{ schedule.type }}
    </div>
  </div>
</template>
@