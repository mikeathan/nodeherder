<script setup>
import { Operators, Condition } from "../../models/automation"
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

const emit = defineEmits(['add', 'remove'])

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

// watchEffect(() => name.value = props.name);
// watchEffect(() => data.value = props.data);
// watchEffect(() => operator.value = props.operator);

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

watch(
    () => props.id,
    (i) => {
        if (i == 0 && props.exposes != null) {
            //name.value = props.exposes[0]
        }
    },
    { immediate: true }
);

function reset() {
    data.value = ""
    operator.value = Operators[0].value
}

</script>

<template>
    <div v-if="props.id == 0" class="col">
        <select id="exposeSelector" style="text-align:center;" class="form-control" @change="exposeSelectionChanged"
            v-model="name" :disabled="name == ''">
            <option value="">Select trigger</option>
            <option v-for="name in props.exposes" :value="name" :key="name">
                {{ name }}
            </option>
        </select>
    </div>
    <div v-else class="col">
        <input type="text" class="form-control" placeholder="Condition name" onfocus="this.placeholder = ''"
            style="text-align: center;" onblur="this.placeholder='Condition name'" v-model="name" disabled />
    </div>
    <div class="col">
        <select id="selectOperators" style="text-align:center;" class="form-control" v-model="operator"
            :disabled="name == ''">
            <option v-for="operator in Operators" :value="operator.value" :key="operator.value">
                {{ operator.text }}
            </option>
        </select>
    </div>
    <div class="col">
        <input type="text" style="text-align:center;" class="form-control" placeholder="Condition value"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'" v-model="data"
            :disabled="name == ''" />
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
