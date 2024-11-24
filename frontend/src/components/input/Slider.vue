<script setup lang="ts">
import { ref, watch, watchEffect } from "vue";
import Slider from 'primevue/slider';

const emit = defineEmits<{
    (e: 'update', id: number): void
}>()

export interface Props {
    value: number,
    min: number
    max: number
}

const props = withDefaults(defineProps<Props>(), {
    min: 0,
    max: 254
})

const value = ref<number | null>()
const max = ref<number>(0)
const min = ref<number>(254)


watchEffect(() => min.value = props.min);
watchEffect(() => max.value = props.max);

watch(
    () => props.value,
    () => {

        value.value = props.value
    }, { immediate: true }
)

function valueChanged(event: any) {
    var v = parseInt((event.target as HTMLInputElement).value)
    emit('update', v);
}

</script>

<template>

    <!-- <Slider id="slider" v-model="props.value" :min="min" :max="max" @change="valueChanged" /> -->
    <input type="range" class="form-range form-control border-0" id="rangeSelector" @change="valueChanged"
        v-model="value" :max="max" :min="min" :disabled="value == null">
</template>