<script setup>
import { Operators, ExposeCondition } from "../../models/automation"
import { defineEmits, ref } from 'vue'

const condition = ref(new ExposeCondition());
const props = defineProps({
    Expose: Object,
    Condition: Object,

    name: String
});

const emit = defineEmits(['add', 'remove', 'update'])
function addCondition() {
    emit(.....)
}

function removeLastCondition() {
    emit(....)

}

</script>

<template>
    <div class="container-fluid p-0 h-100">
        <div class="row w-50">
            <div class="col">
                <input type="text" class="form-control" placeholder="Condition name" onfocus="this.placeholder = ''"
                    onblur="this.placeholder='Condition name'" v-model="props.name" disabled />
            </div>
            <div class="col">
                <select id="selectOperators" style="text-align:center;" class="form-control" v-model="condition.Operator">
                    <option v-for="operator in Operators" :value="operator.value" :key="operator.value">
                        {{ operator.text }}
                    </option>
                </select>
            </div>
            <div class="col">
                <input type="text" style="text-align:center;" class="form-control" placeholder="Condition value"
                    onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'" v-model="condition.Value" />
            </div>
            <div class="col-3 pt-3">
                <div class="btn-group" role="group">
                    <input type="button" class="btn btn-secondary text-nowrap" value="Add" @click="addCondition()" />
                    <input type="button" class="btn btn-secondary text-nowrap" value="Remove"
                        @click="removeLastCondition()" />
                </div>
            </div>
        </div>
        <!-- <div class="row w-50">
            <div class="col">
                <input type="text" class="form-control" placeholder="Condition name" onfocus="this.placeholder = ''"
                    onblur="this.placeholder='Condition name'" v-model="props.Expose.Name" disabled />
            </div>
            <div class="col">
                <select id="selectOperators" style="text-align:center;" class="form-control"
                    v-model="props.Condition.Operator">
                    <option v-for="operator in Operators" :value="operator.value" :key="operator.value">
                        {{ operator.text }}
                    </option>
                </select>
            </div>
            <div class="col">
                <input type="text" style="text-align:center;" class="form-control" placeholder="Condition value"
                    onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'"
                    v-model="props.Condition.Value" />
            </div>

        </div> -->
    </div>
</template>
