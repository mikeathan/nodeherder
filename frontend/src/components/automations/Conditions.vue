<script setup>
import { useStore } from "vuex";
import { Operators, Condition } from "../../models/automation"
import { defineEmits, computed, ref, onMounted, onUpdated } from 'vue'
const store = useStore();
const props = defineProps({
    exposeName: String,
    expose: Object
});

// onMounted(() => {
//     console.log("onMounted ", conditions.value.length);
//     if (conditions.value.length == 0) {
//         conditions.value.push(new Condition());
//     }
// });

const conditions = ref(props.expose.Conditions)
const condition = ref(new Condition())

// const device = computed(() => {
//     return store.getters["devices/find"](props.trigger.id);
// });

const emit = defineEmits(['add', 'remove', 'update'])

function add(condition) {
    condition.Name = props.exposeName;
    var newcondition = new Condition();
    newcondition.Name = props.exposeName;
    conditions.value.unshift(new Condition());
    emit("add", condition)
}

function remove(condition) {
    var index = conditions.value.indexOf(condition);
    if (index !== -1) {
        conditions.value.splice(index, 1);
        emit("remove", condition)
    }
}
</script>

<template>
    <div class="container-fluid p-0 h-100">
        <div v-if="props.expose.Conditions.length > 0">
            <div v-for="condition in props.expose.Conditions">

                <div class="row w-50">
                    <div class="col">
                        <input type="text" class="form-controlTEST" placeholder="Condition name"
                            onfocus="this.placeholder = ''" onblur="this.placeholder='Condition name'"
                            v-model="props.exposeName" disabled />
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
        <div v-else> FIRST TIME

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
                        <button type="button" class="btn btn-default btn-number" @click="add(condition)">
                            <span class="fa fa-plus"></span>
                        </button>

                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
