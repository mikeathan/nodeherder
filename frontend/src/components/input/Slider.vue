<script setup lang="ts">
import { ref, watch, watchEffect } from "vue";

const emit = defineEmits<{
    (e: 'change', id: number): void
}>()

const props = defineProps<{
    placeholder?: string
    value?: number,
    min: { type: number, default: 0 }
    max: { type: number, default: 254 }
}>()

watchEffect(() => value.value = props.value);
watchEffect(() => min.value = props.min);
watchEffect(() => max.value = props.max);



function valueChanged(event: Event) {
    var v = parseInt((event.target as HTMLInputElement).value)
    emit('change', v);
}

const value = ref()
const max = ref()
const min = ref()

</script>

<template>
    <div>
        <label v-if="props.placeholder != null" for="rangeSelector" class="form-label">{{ props.placeholder }}</label>
        <input type="range" class="form-range" id="rangeSelector" @change="valueChanged" v-model="value" :max="max"
            :min="min">
    </div>
</template>