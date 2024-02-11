
<script setup lang="ts">
import { key } from "@/store";
import { KeyyValuePair } from "@/types/types";
import { PropType, computed, ref, watchEffect, watch } from "vue";

const props = defineProps({
    name: String,
    value: {
        type: null,
        required: true
    },
    items: {
        type: Object as PropType<KeyyValuePair<any>>,
        required: true
    }
});

const emit = defineEmits<{
    (e: 'update', value: any): void,
}>()

const selectedValue = ref<any>(null)

watch(
    () => props.value,
    () => {
        selectedValue.value = props.value
    }, { immediate: true }
)

const id = getID()
function isChecked(value: any) {
    return selectedValue.value == value
}

function selectionChanged(event: any) {
    if (event.target.value == null) {
        return;
    }

    selectedValue.value = typeof props.value == 'number' ? parseInt(event.target.value) : event.target.value
    emit("update", selectedValue.value);
}

function getID() {
    return (new Date()).getTime();
}

</script>

<template>
    <div>
        <form>
            <div v-for="(key, value) in props.items" class="btn-group" role="group">
                <div>
                    <input type="radio" class="btn-check" name="options-outlined" :id="`radioSelection${value}${id}`"
                        :value="key" :checked="isChecked(key)" @change="selectionChanged" :disabled="selectedValue == null"
                        onclick="this.blur()">
                    <label class=" btn btn-outline-secondary" :for="`radioSelection${value}${id}`"> {{ value }}</label>
                </div>
            </div>
        </form>
    </div>
</template>

<!-- <div class="input-group align-items-center"><div class="btn-group me-2"><button type="button" class="btn btn-outline-secondary" title="Coolest temperature supported">coolest</button><button type="button" class="btn btn-outline-secondary" title="Cool temperature (250 mireds / 4000 Kelvin)">cool</button><button type="button" class="btn btn-outline-secondary active" title="Neutral temperature (370 mireds / 2700 Kelvin)">neutral</button><button type="button" class="btn btn-outline-secondary" title="Warm temperature (454 mireds / 2200 Kelvin)">warm</button><button type="button" class="btn btn-outline-secondary" title="Warmest temperature supported">warmest</button></div><input min="150" max="500" type="range" class="form-range form-control border-0" value="370"><input type="number" class="form-control ms-1" min="150" max="500" style="max-width: 100px;" value="370"><span class="input-group-text" style="min-width: 66px;">mired</span></div> -->

