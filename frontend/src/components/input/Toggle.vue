<script setup lang="ts">
import { computed, ref } from "vue";

const emit = defineEmits<{
    (e: 'update', value: any): void,
}>()

export interface Props {
    value: any,
    valueOn: any,
    valueoff: any,
    minimal?: boolean
}

const props = withDefaults(defineProps<Props>(), {
    minimal: false,
})

const hasValue = computed(() => props.value != null || props.value != undefined);
const showOnOffLabel = computed(() => props.minimal && hasValue && props.valueoff != null && props.valueOn != null);

function valueChanged(event: Event): void {
    emit('update', (event.target as HTMLInputElement).checked ? props.valueOn : props.valueoff);
}

</script>
<template>
    <button v-if="showOnOffLabel" type="button" class="btn btn-link">OFF</button>
    <!-- <ToggleSwitch v-model="checked" /> -->
    <div class="form-check form-switch form-check-inline align-middle me-0">
        <input class="form-check-input" type="checkbox" :checked="props.value == props.valueOn" @change="valueChanged"
            :disabled="!hasValue">
    </div>
    <button v-if="showOnOffLabel" type="button" class="btn btn-link">ON</button>
</template>



