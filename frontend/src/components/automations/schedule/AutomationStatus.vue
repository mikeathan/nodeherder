<script setup lang="ts">
import { computed, PropType, ref } from 'vue';
import { emitOpenSchedulerPanelEvent } from '@/contracts/panel-events';
import { store } from '../../../store/index';
import { Automation } from '@/types/automation.type';

const emit = defineEmits(['cancel']);

const props = defineProps({
  automation: {
    type: Object as PropType<Automation>,
    default: {} as Automation,
    required: true,
  },
});

const automation = ref<Automation>(props.automation);
const hasSchedules = computed(
  () =>
    automation.value.schedules &&
    automation.value.schedules.length != 0
);

function scheduleClick() {
  emitOpenSchedulerPanelEvent(automation.value);
}
function saveAutomation(): void {
  store.dispatch('automations/save', automation.value);
}
</script>
<template>
  <div v-if="hasSchedules">
    <Button label="Scheduled" icon="pi pi-clock" size="small" text @click="scheduleClick()" />
  </div>
  <div v-else>
    <ToggleButton v-model="automation.enabled" onLabel="Enabled" offLabel="Disabled" onIcon="pi pi-check"
      offIcon="pi pi-times" severity="success" size="small" v-on:change="saveAutomation" />
  </div>
</template>
