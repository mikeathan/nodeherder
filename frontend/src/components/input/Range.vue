<script setup lang="ts">
import { InputNumberInputEvent } from "primevue";
import { ref, watch, watchEffect } from "vue";
import { prop } from "vue-class-component";

const emit = defineEmits<{
    (e: 'update', id: number): void
}>()

export interface Props {
    value: any | null,
    min?: number,
    max?: number,
    showInput?: boolean
}

const props = withDefaults(defineProps<Props>(), {
    value: 0,
    min: 0,
    max: 254,
    showInput: false
})

const value = ref<number>()
const max = ref<number>(0)
const min = ref<number>(254)


watchEffect(() => min.value = props.min);
watchEffect(() => max.value = props.max);

watch(
    () => props.value,
    () => {

        if (props.value == null || typeof props.value != "number") {
            value.value = 0;
        } else {
            value.value = props.value
        }
    }, { immediate: true }
)

function sliderChanged(value: number) {
    emit('update', value);
}

function inputChanged(event: InputNumberInputEvent) {
    var v = parseInt(event.value?.toString() || "0")
    emit('update', v);
}


    //@trackClick="onTrackClick"  event
// onTrackClick(event) {
//       const trackRect = event.target.getBoundingClientRect();
//       const clickX = event.clientX - trackRect.left;
//       const trackWidth = trackRect.width;
//       const newPosition = (clickX / trackWidth) * (this.max - this.min);
//       this.value = Math.round(newPosition);
//     },


// onchange event

const releasedValue = ref(50);
const onValueChange = (event) => {
        // Update the releasedValue only when the mouse/touch is released
        releasedValue.value = event.value;
    };
</script>
<style scoped></style>
<template>
    <InputNumber v-model.number="value" v-if="props.showInput" @input="inputChanged" class="w-full mb-4" />
    <Slider v-model="value" @change="sliderChanged" :max="max" :min="min" :disabled="props.value == null"
        class="w-full" />
</template>