<script setup>
import { Operators, ExposeCondition } from "../../models/automation"
import { defineEmits, computed } from 'vue'

const props = defineProps({
    name: String,
    condition: Object
});

const condition = computed(() => {
    console.log("computed ", props.condition);
    if (props.condition == null) {
        console.log("new condition");
        return new ExposeCondition();
    }
    console.log("existing condition");

    return props.condition;
})

const emit = defineEmits(['add', 'remove', 'update'])
function addCondition() {
    emit("add", condition)
}

function remove() {
    emit("remove", condition)
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
                    <!-- <div v-if="props.condition == null"> -->
                    <input type="button" class="btn btn-secondary text-nowrap" value="Add" @click="addCondition()" />
                    <!-- </div>
                    <div v-else> -->
                    <input type="button" class="btn btn-secondary text-nowrap" value="Remove" @click="remove()" />
                    <!-- </div> -->
                </div>
            </div>
        </div>

    </div>
</template>
