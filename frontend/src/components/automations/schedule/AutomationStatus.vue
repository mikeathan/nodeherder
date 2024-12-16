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
<style scoped></style>
<template>
  <div v-if="hasSchedules" class="form-check form-switch">
    <!-- <label for="automation-status" class="pe-2"
      >Scheduled</label
    > -->
    <!-- <input
      id="automation-status"
      class="form-check-input"
      type="checkbox"
      role="switch"
      v-model="automation.enabled"
      :disabled="true" /> -->
    <!-- <ToggleSwitch
      v-model="automation.enabled"
      :disabled="true" /> -->
    <Button
      label="Scheduled"
      icon="pi pi-clock"
      size="small"
      text />

    <!-- <i
      class="fa-solid fa-clock ps-1"
      @click="scheduleClick()"></i> -->

    <!-- <div class="col-1">
      <Button
        icon="pi pi-clock"
        variant="text"
        size="small"
        @click="scheduleClick()" />
    </div> -->
  </div>
  <div v-else class="form-check form-switch">
    <label class="pe-2">Enabled</label>
    <ToggleSwitch v-model="automation.enabled" />
    <!-- <input
        class="form-check-input"
        type="checkbox"
        role="switch"
        v-model="automation.enabled" /> -->
  </div>
</template>
