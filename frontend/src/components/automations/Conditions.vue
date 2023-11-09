<script setup>
import { Operators, Condition } from "../../models/automation"
import TriggerCondition from "./TriggerCondition"

import { ref, watchEffect } from 'vue'
const props = defineProps({
    expose: Object
});

const conditions = ref(props.expose.Conditions);
const emit = defineEmits(['add', 'remove'])

watchEffect(() => conditions.value = props.expose.Conditions);

function add(event) {
    conditions.value.push(event);
}


function remove(event) {
    var index = conditions.value.findIndex(item => item.Id === event);
    if (index != -1) {
        conditions.value.splice(index, 1);
    }
}

</script>

<template>
    <div class="container-fluid p-0 h-100">
        <div class="row w-50">
            <TriggerCondition :id="0" :name="props.expose.Name" :operator="Operators[0].value" :data="''"
                @add="add($event)">
            </TriggerCondition>
        </div>

        <div v-for="condition in conditions">
            <div class="row w-50">
                <TriggerCondition :id="condition.Id" :name="condition.Name" :operator="condition.Operator"
                    :key="condition.id" :data="condition.Data" @remove="remove($event)"></TriggerCondition>
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
