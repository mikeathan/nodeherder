<script setup>
import { ref, watch } from "vue";

const props = defineProps({
    name: {
        type: String,
        required: true,
    },
    data: null, // could be string or array of string
    type: {
        type: String,
        default: "numeric"
    },
    disabled: Boolean
});

const data = ref(null);
const name = ref("");
const emit = defineEmits(['update:data'])


watch(
    () => props.data,
    () => {
        if (props.type == "binary" && typeof props.data == "object") {
            var values = Object.values(props.data)
            data.value = values[0]
            return
        }
        data.value = props.data
        // data.value = parseInt(props.data) // ????????? maybe cast it 
    }, { immediate: true }
)

watch(
    () => props.name,
    () => {
        name.value = props.name
    }, { immediate: true }
)

function dataInputChange(event) {
    data.value = event.target.value.replace(/[^0-9.]/g, '');
    console.log("datainput - data ", data.value)
    emit("update:data", data.value);
}

function dataSelectionChanged(event) {
    if (event.target.value == null) {
        return;
    }
    emit("update:data", event.target.value);
}

</script>

<style scoped>
.inputName {
    border: 0;
    outline: 0;
    background: transparent;
    border-bottom: 1px solid #e5e5e5;
    border-radius: 0
}

.custom-control-input {
    transform: scale(1.4);
}
</style>
<template>
    <div v-if="type == 'binary'">
        <!-- <label class="form-check-label" for="valueinput">Value</label> -->

        <select id="dataSelect" style="text-align:center;" class="form-control inputName" name="valueinput" v-model="data"
            @change="dataSelectionChanged" :disabled="props.disabled">
            <option value="">Select item</option>
            <option v-for="(value, key) in props.data" :value="value" :key="key">
                {{ value }}
            </option>
        </select>
    </div>
    <div v-else>
        <!-- <label class="form-check-label" for="valueinput">Value</label> -->

        <input type="text" class="form-control inputName" name="valueinput" :placeholder="props.name"
            onfocus="this.placeholder = ''" v-model="data" @input="dataInputChange" :disabled="props.disabled">
    </div>
</template>
<!-- onblur="this.placeholder = 'test'" -->
