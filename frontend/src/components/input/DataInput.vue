<script setup>
import { ref, watch } from "vue";

const props = defineProps({
    placeholder: {
        type: String,
        default: ""
    },
    data: null, // could be string or array of string
    items: null,
    type: {
        type: String,
        default: "numeric"
    },
    disabled: Boolean
});

const data = ref(null);
const placeholder = ref("");
const items = ref(['OFF', 'ON'])
const emit = defineEmits(['update:data'])


watch(
    () => props.data,
    () => {
        data.value = props.data
        // data.value = parseInt(props.data) // ????????? maybe cast it 
    }, { immediate: true }
)

watch(
    () => props.items,
    () => {
        if (props.items != null) {
            items.value = props.items
        }
    }, { immediate: true }
)

watch(
    () => props.placeholder,
    () => {
        placeholder.value = props.placeholder
    }, { immediate: true }
)

function dataInputChange(event) {
    data.value = event.target.value.replace(/[^0-9.]/g, '');
    emit("update:data", data.value);
}

function dataSelectionChanged(event) {
    if (event.target.value == null) {
        return;
    }
    data.value = event.target.value
    emit("update:data", data.value);
}

function blurChanged(event) {
    placeholder.value = props.placeholder
}

function focusChanged(event) {
    placeholder.value = ""
}

</script>

<style scoped>
.input-outline {
    border: 0;
    outline: 0;
    background: transparent;
    border-bottom: 1px solid #e5e5e5;
    border-radius: 0;
}

.select-outline {
    border: 0;
    outline: 0;
    border-bottom: 1px solid #e5e5e5;
}

.custom-control-input {
    transform: scale(1.4);
}
</style>
<template>
    <div v-if="props.type == 'binary'">

        <select id="dataSelect" style="text-align:center;" class="form-control form-select select-outline" name="valueinput"
            v-model="data" @change="dataSelectionChanged" :disabled="props.disabled">
            <option v-if="props.placeholder != ''" value="">{{ props.placeholder }}</option>
            <option v-for="(value, key) in items" :value="value" :key="key">
                {{ value }}
            </option>
        </select>
    </div>
    <div v-else>
        <!-- <label class="form-check-label" for="valueinput">Value</label> -->

        <input type="text" class="form-control input-outline" style="text-align:center;" name="valueinput"
            :placeholder="placeholder" v-model="data" @input="dataInputChange" :disabled="props.disabled"
            @focus="focusChanged" @blur="blurChanged">
    </div>
</template>
<!-- onblur="this.placeholder = 'test'" -->
