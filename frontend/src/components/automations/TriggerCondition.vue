<script setup>
import { OperatorKeys, Condition } from "../../models/automation"
import DataInput from "../input/DataInput.vue"
import { useStore } from "vuex";

import { ref, computed, watch } from 'vue'
const props = defineProps({
    id: {
        type: String,
        required: true,
    },
    name: String,
    operator: String,
    data: null,
    index: {
        type: Number,
        required: true,
        default: -1,
    }
});

const store = useStore()
const index = ref(props.index);
const data = ref(null)
const operator = ref('')
const name = ref('')

const emit = defineEmits(['add', 'remove', 'update:value', 'update:operator'])

function add() {
    var c = new Condition(name.value, operator.value, data.value)
    var cid = ++index.value;
    c.idx = cid

    emit("add", c)
    reset();
}

function remove() {
    emit("remove", props.index)
}

function getExposes() {
    return Object.keys(device.value.exposes)
}

const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const feature = computed(() => {
    if (name.value == '') {
        return []
    }

    var device = store.getters["devices/find"](props.id);
    if (device.exposes[name.value] == undefined) {

        return []
    }

    return device.exposes[name.value];
});

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

    data.value = "";
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
    <div v-if="props.index == 0" class="col">
        <select id="exposeSelector" style="text-align:center;" class="form-control" @change="exposeSelectionChanged"
            v-model="name" :disabled="props.name == ''">
            <option value="">Select trigger</option>
            <option v-for="name in getExposes()" :value="name" :key="name">
                {{ name }}
            </option>
        </select>
    </div>
    <div v-else class="col">
        <DataInput type="string" placeholder="Name" :data="name" :disabled="true">
        </DataInput>
    </div>
    <div class="col">
        <DataInput type="binary" :items="OperatorKeys" :data="operator" :disabled="name == ''"
            @update:data="operatorUpdated">
        </DataInput>
    </div>
    <div class="col">
        <DataInput :type="feature.type" :placeholder="feature.type == 'binary' ? 'Select' : 'Value'" placeholder="Value"
            :data="data" :items="feature.type == 'binary' ? [true, false] : null" :disabled="name == ''"
            @update:data="dataUpdated">
        </DataInput>
    </div>
    <div class="col-3">
        <div class="btn-group">
            <div v-if="props.index == 0">
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
