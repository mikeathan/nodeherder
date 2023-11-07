<script setup>
import { Operators, ExposeCondition } from "../../models/automation"
import { defineEmits, computed, ref, onMounted, onUpdated } from 'vue'

const props = defineProps({
    name: String,
    condition: Object
});

const emit = defineEmits(['add', 'remove', 'update'])

const isNull = computed(() => {
    if (condition.value == null) {
        return false;
    }
    return condition.value.Data == null
})

const hasData = computed(() => {

    console.log("hasData", condition.value.Data?.length > 0);
    return condition.value.Data?.length > 0
})


const condition = computed(() => {

    if (props.condition.Data == null) {
        return new ExposeCondition();
    }

    return props.condition;
})

function add() {
    console.log("add");
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
                    onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'" v-model="condition.Data" />
            </div>
            <div class="col-3">
                <div class="btn-group">
                    <div v-if="isNull">
                        <button type="button" class="btn btn-default btn-number" @click="add()"
                            :disabled="hasData == false">
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
        </div>
    </div>
</template>
