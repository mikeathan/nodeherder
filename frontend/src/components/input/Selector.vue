<script setup>
import { ref, watch } from "vue";

const props = defineProps({
    placeholder: {
        type: String,
        default: ""
    },
    value: null, // could be string or array of string
    items: Object,
    disabled: Boolean,
    alignment: {
        type: String,
        default: 'center'
    }
});

const selected = ref(null);
const emit = defineEmits(['update:data'])

watch(
    () => props.value,
    () => {
        selected.value = props.value

    }, { immediate: true }
)

function dataSelectionChanged(event) {
    if (event.target.value == null) {
        return;
    }

    if (typeof selected.value === "boolean") {
        selected.value = Boolean(event.target.value).valueOf()
    } else {
        selected.value = event.target.value
    }

    emit("update:data", selected.value);
}

</script>

<style scoped>
.select-outline {
    border: 0;
    outline: 0;
    border-bottom: 1px solid #e5e5e5;
}
</style>
<template>
    <select id="dataSelect" :style="'text-align:' + props.alignment + ';'" class="form-control form-select"
        name="valueinput" v-model="selected" @change="dataSelectionChanged" :disabled="props.disabled">
        <option v-if="props.placeholder != ''" value="">{{ props.placeholder }}</option>
        <option v-for="(value, key) in props.items" :value="value" :key="value">
            {{ key }}
        </option>
    </select>
</template>
