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
select.form-control {
    border: 0;
    outline: 0;
    border-radius: 0%;
    border-bottom: 1px solid white;
    background-image: none;
}

select.form-control:disabled {
    color: gray;
    background-color: transparent;
    background-image: none
}

select.form-control:hover:not([disabled]) {
    background-image: url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 16 16%27%3E%3Cpath fill=%27none%27 stroke=%27%23d4d6d9%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27 stroke-width=%272%27 d=%27m2 5 6 6 6-6%27/%3E%3C/svg%3E");
    box-shadow: none;
}

select.form-control:focus,
:active {
    box-shadow: none;
}

select.form-control:first-of-type {
    border-bottom: 0px solid white;
}

select.form-control:required:invalid {
    color: gray;
    border-bottom: 1px solid white;
}
</style>
<template>
    <div>
        <select required id="dataSelect" :style="'text-align:' + props.alignment + ';'" class="form-control form-select"
            v-model="selected" @change="dataSelectionChanged" :disabled="props.disabled">
            <option v-if="props.placeholder != ''" value="">{{ props.placeholder }}</option>
            <option v-for="(value, key) in props.items" :value="value" :key="value">
                {{ key }}
            </option>
        </select>
    </div>
</template>
