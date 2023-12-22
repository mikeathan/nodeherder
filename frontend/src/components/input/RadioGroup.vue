<script setup>
import { ref, watch, watchEffect } from "vue";

const props = defineProps({
    name: String,
    value: null,
    items: Array,
});

const emit = defineEmits(['update:data'])
const id = getID()
function isChecked(value) {
    return props.value == value
}

function selectionChanged(event) {

    if (event.target.value == null) {
        return;
    }
    emit("update:data", event.target.value);
}

function getID() {
    return (new Date()).getTime();
}

</script>

<template>
    <div>
        <form>


            <div v-for="item in props.items" class="btn-group">


                <div>
                    <input type="radio" class="btn-check" name="options-outlined" :id="`radioSelection${item.name}${id}`"
                        :value="item.value" :checked="isChecked(item.value)" @change="selectionChanged">

                    <label class="btn btn-outline-secondary" :for="`radioSelection${item.name}${id}`"> {{ item.name
                    }}</label>
                </div>
            </div>
        </form>
    </div>
</template>


