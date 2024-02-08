<script setup>

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

<!-- <div class="input-group align-items-center"><div class="btn-group me-2"><button type="button" class="btn btn-outline-secondary" title="Coolest temperature supported">coolest</button><button type="button" class="btn btn-outline-secondary" title="Cool temperature (250 mireds / 4000 Kelvin)">cool</button><button type="button" class="btn btn-outline-secondary active" title="Neutral temperature (370 mireds / 2700 Kelvin)">neutral</button><button type="button" class="btn btn-outline-secondary" title="Warm temperature (454 mireds / 2200 Kelvin)">warm</button><button type="button" class="btn btn-outline-secondary" title="Warmest temperature supported">warmest</button></div><input min="150" max="500" type="range" class="form-range form-control border-0" value="370"><input type="number" class="form-control ms-1" min="150" max="500" style="max-width: 100px;" value="370"><span class="input-group-text" style="min-width: 66px;">mired</span></div> -->

