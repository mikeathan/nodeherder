<script setup lang="ts">
import { computed, PropType, ref } from "vue";
import { Automation } from "@/types/automation";
import { emitOpenSchedulerPanelEvent } from "@/contracts/panel-events";

const emit = defineEmits(['cancel'])

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

const automation = ref<Automation>(props.automation)
const hasSchedules = computed(() =>
    automation.value.schedules && automation.value.schedules.length != 0)

function scheduleClick() {
    if (props.clickToOpen) {
        emitOpenSchedulerPanelEvent(automation.value)
    }
}
</script>

<template>
    <ToggleButton v-model="automation.enabled" onLabel="Enabled" offLabel="Disabled" size="small" />

    <div class="form-check form-switch ">
        <div v-if="hasSchedules">
            <label class="form-check-label ">Scheduled</label>
            <input class="form-check-input " type="checkbox" role="switch" v-model="automation.enabled"
                :disabled="true" />
            <i class="fa-solid fa-clock ps-1" @click="scheduleClick()"></i>
        </div>
        <div v-else>
            <label class=" form-check-label ">Enabled</label>
            <input class="form-check-input" type="checkbox" role="switch" v-model="automation.enabled" />
        </div>
    </div>

</template>
