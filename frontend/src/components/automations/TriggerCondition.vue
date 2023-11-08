<script setup>
import { Operators, Condition } from "../../models/automation"
import { defineEmits, ref } from 'vue'
const props = defineProps({
    name: String,
    operator: String,
    data: String,
    id: {
        type: Number,
        required: true,
        default: 0,
    },
});

const id = ref(props.id);
const data = ref(props.data)
const operator = ref(props.operator)

const emit = defineEmits(['add', 'remove', 'update'])

function add() {
    var c = new Condition(props.name, operator.value, data.value)
    c.id = ++id.value
    emit("add", c)

    reset();
}

function remove() {
    emit("remove", id)
}

function reset() {
    data.value = ""
    operator.value = Operators[0].value
}
</script>

<template>
    <div class="col">
        {{ id }}
    </div>
    <div class="col">
        <input type="text" class="form-control" placeholder="Condition name" onfocus="this.placeholder = ''"
            onblur="this.placeholder='Condition name'" v-model="props.name" disabled />
    </div>
    <div class="col">
        <select id="selectOperators" style="text-align:center;" class="form-control" v-model="operator">
            <option v-for="operator in Operators" :value="operator.value" :key="operator.value">
                {{ operator.text }}
            </option>
        </select>
    </div>
    <div class="col">
        <input type="text" style="text-align:center;" class="form-control" placeholder="Condition value"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'" v-model="data" />
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
