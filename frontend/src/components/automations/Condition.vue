<script setup>
import { useStore } from "vuex";
import { Operators, ExposeCondition } from "../../models/automation"
import { defineEmits, computed, ref, onMounted, onUpdated } from 'vue'
const store = useStore();
const props = defineProps({
    exposeName: String,
});

const conditions = ref([new ExposeCondition()])
const device = computed(() => {
    return store.getters["devices/find"](props.trigger.id);
});

const emit = defineEmits(['add', 'remove', 'update'])


function add(condition) {
    console.log("add");
    conditions.value.unshift(new ExposeCondition());
    // emit("add", condition)
}

function remove(condition) {
    var index = conditions.value.indexOf(condition);
    if (index !== -1) {
        conditions.value.splice(index, 1);
        //  emit("remove", condition)
    }
}
</script>

<template>
    <div class="container-fluid p-0 h-100">

        <div v-for="condition in conditions">

            <div class="row w-50">
                <div class="col">
                    <input type="text" class="form-control" placeholder="Condition name" onfocus="this.placeholder = ''"
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
                        <div v-if="conditions.indexOf(condition) == 0">
                            <button type="button" class="btn btn-default btn-number" @click="add(condition)"
                                :disabled="condition.Data.length == 0">
                                <span class="fa fa-plus"></span>
                            </button>
                        </div>
                        <div v-else>
                            <button type="button" class="btn btn-default btn-number" @click="remove(condition)">
                                <span class="fa fa-minus"></span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
