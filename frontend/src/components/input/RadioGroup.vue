<script setup>
import { ref, watch } from "vue";

const props = defineProps({
    placeholder: {
        type: String,
        default: ""
    },
    value: null,
    items: Array,
    disabled: Boolean,
});

const emit = defineEmits(['update:data'])

</script>

<template>
    <div class="container">
        <div class="row">
            <div class="col-md-8">
                <form class="form-inline" action="#" method="post">


                    <div class="form-check form-check-inline">
                        <input class="form-check-input" type="checkbox" id="inlineCheckbox1" value="option1">
                        <label class="form-check-label" for="inlineCheckbox1">1</label>
                    </div>
                    <div class="form-check form-check-inline">
                        <input class="form-check-input" type="checkbox" id="inlineCheckbox2" value="option2">
                        <label class="form-check-label" for="inlineCheckbox2">2</label>
                    </div>
                    <div class="form-check form-check-inline">
                        <input class="form-check-input" type="checkbox" id="inlineCheckbox3" value="option3" disabled>
                        <label class="form-check-label" for="inlineCheckbox3">3 (disabled)</label>
                    </div>
                </form>
            </div>
        </div>
    </div>

    <div v-for="item in props.items">
        <div class="form-check form-check-inline">
            <input class="form-check-input " type="radio" :name="`radio${item}`" :id="`radio${item}`" :value="item"
                v-model="props.value">
            <label class="form-check-label" :for="`radio${item}`">
                {{ item }}
            </label>
        </div>
    </div>
</template>


