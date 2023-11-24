<script setup>
import { OperatorKeys, Condition } from "../../models/automation"
import DataInput from "../input/DataInput.vue"

import { ref, watchEffect, watch } from 'vue'
const props = defineProps({
    exposes:
    {
        type: Array,
        default: () => []
    },
    name: String,
    operator: String,
    data: String,
    id: {
        type: Number,
        required: true,
        default: 0,
    }
});

const id = ref(props.id);
const data = ref('')
const operator = ref('')
const name = ref('')

const emit = defineEmits(['add', 'remove', 'update:value', 'update:operator'])

function add() {
    var c = new Condition(name.value, operator.value, data.value)
    var cid = ++id.value;
    c.idx = cid

    emit("add", c)
    reset();
}

function remove() {
    emit("remove", props.id)
}

watch(
    () => props.name,
    () => {
        name.value = props.name
    }, { immediate: true }
)

watch(
    () => props.data,
    () => {
        data.value = props.data
    }, { immediate: true }
)

watch(
    () => props.operator,
    () => {
        operator.value = props.operator
    }, { immediate: true }
)

function reset() {
    data.value = ""
    operator.value = OperatorKeys[0]
}

function exposeSelectionChanged(event) {
    if (event.target.value == null) {
        return
    }

    // TODO: maybe reset operators select on change
}

function operatorUpdated(event) {
    operator.value = event
    emit('update:operator', event)
}
function dataUpdated(event) {
    data.value = event
    emit('update:value', event)
}
</script>

<template>
    <div v-if="props.id == 0" class="col-xl-3">
        <select id="exposeSelector" style="text-align:center;" class="form-control" @change="exposeSelectionChanged"
            v-model="name" :disabled="props.name == ''">
            <option value="">Select trigger</option>
            <option v-for="name in props.exposes" :value="name" :key="name">
                {{ name }}
            </option>
        </select>
    </div>
    <div v-else class="col-xl-3">
        <DataInput placeholder="Name" :data="name" :disabled="true">
        </DataInput>
    </div>
    <div class="col-xl-3">
        <DataInput type="binary" :items="OperatorKeys" :data="operator" :disabled="name == ''"
            @update:data="operatorUpdated">
        </DataInput>
    </div>
    <div class="col-xl-3">
        <DataInput placeholder="Value" :data="data" :disabled="name == ''" @update:data="dataUpdated">
        </DataInput>
    </div>
    <div class="col-3">
        <div class="btn-group">
            <div v-if="props.id == 0">
                <button type="button" class="btn btn-default btn-number" @click="add($event)" :disabled="data.length == 0">
                    <span class="fa fa-plus"></span>
                </button>
            </div>
            <div v-else>
                <button type="button" class="btn btn-default btn-number" @click="remove($event)">
                    <span class="fa fa-minus"></span>
                </button>
            </div>
        </div>
    </div>
</template>
