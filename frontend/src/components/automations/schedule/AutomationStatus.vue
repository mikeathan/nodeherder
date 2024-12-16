<script setup lang="ts">
import { computed, PropType, ref } from 'vue';
import { Automation } from '@/types/automation';
import { emitOpenSchedulerPanelEvent } from '@/contracts/panel-events';

const emit = defineEmits(['cancel']);

const props = defineProps({
  automation: {
    type: Object as PropType<Automation>,
    default: {} as Automation,
    required: true,
  },
  clickToOpen: {
    type: Boolean,
    default: false,
  },
});

const automation = ref<Automation>(props.automation);
const hasSchedules = computed(
  () =>
    automation.value.schedules &&
    automation.value.schedules.length != 0,
);

function scheduleClick() {
  if (props.clickToOpen) {
    emitOpenSchedulerPanelEvent(automation.value);
  }
}
</script>
<template>
  <div v-if="hasSchedules">
    <!-- <input
      id="automation-status"
      class="form-check-input"
      type="checkbox"
      role="switch"
      v-model="automation.enabled"
      :disabled="true" /> -->

    <Button label="Scheduled" icon="pi pi-clock" size="small" text @click="scheduleClick()" />
  </div>
  <div v-else>

    <ToggleButton v-model="automation.enabled" onLabel="Enabled" offLabel="Disabled" onIcon="pi pi-check"
      offIcon="pi pi-times" severity="success" size="small" />

  </div>
</template>
