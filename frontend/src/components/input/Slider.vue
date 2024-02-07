<script setup lang="ts">
import { ref, watch, watchEffect } from "vue";

const emit = defineEmits<{
    (e: 'update:value', id: number): void
}>()

export interface Props {
    placeholder?: string
    value?: number,
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

function valueChanged(event: Event) {
    var v = parseInt((event.target as HTMLInputElement).value)
    emit('update:value', v);
}

</script>

<template>
    <label v-if="props.placeholder != null" for="rangeSelector" class="form-label">{{ props.placeholder }}</label>
    <input type="range" class="form-range form-control border-0" id="rangeSelector" @change="valueChanged" v-model="value"
        :max="max" :min="min">
</template>