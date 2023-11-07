<script setup>
import { Operators, Condition } from "../../models/automation"
import Conditions from "./Conditions"

import { defineEmits, ref } from 'vue'
const props = defineProps({
    exposeName: String,
    expose: Object
});

const operator = ref(Operators[0].value);
const data = ref("")


const emit = defineEmits(['add', 'remove', 'update'])

function add() {
    props.expose.Conditions.push(new Condition(props.exposeName, operator.value, data.value));
    reset()
}

function reset() {
    data.value = ""
    operator.value = Operators[0].value
}

function remove(condition) {
    var index = props.expose.Conditions.indexOf(condition);
    if (index !== -1) {
        props.expose.Conditions.splice(index, 1);
        // emit("remove", condition)
    }
}

</script>

<template>
    <div class="container-fluid p-0 h-100">

        <div class="row w-50">
            <Condition id="0" :name="props.exposeName" :operator="Operators[0]" data=""></Condition>
        </div>

        <div v-for="condition in props.expose.Conditions">
            <div class="row w-50">
                <Condition :id="condition.id" :name="props.exposeName" :operator="condition.operator"
                    :data="condition.data"></Condition>
            </div>
        </div>
        <!-- <div class="row w-50">
            <div class="col">
                <input type="text" class="form-control" placeholder="Condition name" onfocus="this.placeholder = ''"
                    onblur="this.placeholder='Condition name'" v-model="props.exposeName" disabled />
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
                    <button type="button" class="btn btn-default btn-number" @click="add()" :disabled="data.length == 0">
                        <span class="fa fa-plus"></span>
                    </button>

                </div>
            </div>
        </div>

        <div v-for="condition in props.expose.Conditions">
            <div class="row w-50">
                <div class="col">
                    <input type="text" class="form-contro" placeholder="Condition name" onfocus="this.placeholder = ''"
                        onblur="this.placeholder='Condition name'" v-model="props.exposeName" disabled />
                </div>
                <div class="col">
                    <select id="selectOperators" style="text-align:center;" class="form-control"
                        v-model="condition.Operator">
                        <option v-for="operator in Operators" :value="operator.value" :key="operator.value">
                            {{ operator.text }}
                        </option>
                    </select>
                </div>
                <div class="col">
                    <input type="text" style="text-align:center;" class="form-control" placeholder="Condition value"
                        onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'"
                        v-model="condition.Data" />
                </div>
                <div class="col-3">
                    <div class="btn-group">

                        <button type="button" class="btn btn-default btn-number" @click="remove(condition)">
                            <span class="fa fa-minus"></span>
                        </button>
                    </div>
                </div>
            </div>
        </div> -->
    </div>
</template>
