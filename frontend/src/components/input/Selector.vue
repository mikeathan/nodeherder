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
select.form-control,
input.form-control {
    border: 0;
    outline: 0;
    border-radius: 0%;
    border-bottom: 1px solid white;
}


select.form-control:disabled,
input.form-control:disabled {
    border: 0;
    outline: 0;
    border-radius: 0%;
    background-color: transparent;
    border-bottom: 0px solid white;
}


input.form-control:focus,
:active,
:hover,
select.form-control:focus,
:active,
:hover {
    box-shadow: none;
}
</style>
<template>
    <div>
        <select id="dataSelect" :style="'text-align:' + props.alignment + ';'" class="form-control" v-model="selected"
            @change="dataSelectionChanged" :disabled="props.disabled">
            <option v-if="props.placeholder != ''" value="">{{ props.placeholder }}</option>
            <option v-for="(value, key) in props.items" :value="value" :key="value">
                {{ key }}
            </option>
        </select>
    </div>
</template>
