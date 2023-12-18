<script setup >
import { ref, watch, watchEffect } from "vue";

const emit = defineEmits(['update:value'])

const props = defineProps({
    placeholder:
    {
        type: String,
        default: null
    },
    value: {
        type: Number,
        default: null
    },
    min: {
        type: Number,
        default: 0
    },
    max: {
        type: Number,
        default: 254
    }
});

const value = ref(0)
const max = ref(0)
const min = ref(254)

watchEffect(() => min.value = props.min);
watchEffect(() => max.value = props.max);

watch(
    () => props.value,
    () => {
        value.value = props.value
    }, { immediate: true }
)
function valueChanged(event) {
    var v = parseInt(event.target.value)
    emit('update:value', v);
}

</script>

<template>
    <div>
        <label v-if="props.placeholder != null" for="rangeSelector" class="form-label">{{ props.placeholder }}</label>
        <input type="range" class="form-range" id="rangeSelector" @change="valueChanged" v-model="value" :max="max"
            :min="min">
    </div>
</template>




<!-- <script setup lang="ts">
import { ref, watch, watchEffect } from "vue";

const emit = defineEmits<{
    (e: 'change', id: number): void
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

watchEffect(() => value.value = props.value);
watchEffect(() => min.value = props.min);
watchEffect(() => max.value = props.max);

function valueChanged(event: Event) {
    var v = parseInt((event.target as HTMLInputElement).value)
    emit('change', v);
}

const value = ref()
const max = ref(0)
const min = ref()

</script>

<template>
    https://github.com/vuejs/vue-eslint-parser/issues/49
    <div>
        <label v-if="props.placeholder != null" for="rangeSelector" class="form-label">{{ props.placeholder }}</label>
        <input type="range" class="form-range" id="rangeSelector" @change="valueChanged" v-model="value" :max="max"
            :min="min">
    </div>
</template> -->