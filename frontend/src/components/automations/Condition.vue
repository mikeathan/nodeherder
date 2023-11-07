<script setup>
import { Operators, Condition } from "../../models/automation"
import Condition from "../../models/automation"
import { defineEmits, ref } from 'vue'
const props = defineProps({
    name: String,
    operator: String,
    data: String,
    id: String
});

const emit = defineEmits(['add', 'remove', 'update'])

function add() {
    var c = new Condition(props.name, props.value, props.value)
    c.id = props.id + 1
    emit("add", c)
}

function remove() {
    emit("remove", props.id)
}

</script>

<template>
    <div class="container-fluid p-0 h-100">

        <div class="row w-50">

        </div>

        <!-- <div class="row w-50">
            <div class="col">
                <input type="text" class="form-control" placeholder="Condition name" onfocus="this.placeholder = ''"
                    onblur="this.placeholder='Condition name'" v-model="props.name" disabled />
            </div>
            <div class="col">
                <select id="selectOperators" style="text-align:center;" class="form-control" v-model="props.operator">
                    <option v-for="operator in Operators" :value="operator.value" :key="operator.value">
                        {{ operator.text }}
                    </option>
                </select>
            </div>
            <div class="col">
                <input type="text" style="text-align:center;" class="form-control" placeholder="Condition value"
                    onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'" v-model="props.data" />
            </div>
            <div class="col-3">
                <div class="btn-group">
                    <div v-if="props.id == 0">
                        <button type="button" class="btn btn-default btn-number" @click="add()"
                            :disabled="props.data.length == 0">
                            <span class="fa fa-plus"></span>
                        </button>

                    </div>
                    <div v-else>
                        <button type="button" class="btn btn-default btn-number" @click="remove()">
                            <span class="fa fa-minus"></span>
                        </button>
                    </div>
                </div>
            </div>
        </div> -->
    </div>
</template>
