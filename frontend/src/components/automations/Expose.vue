<script setup>
import Conditions from "./Conditions"

const props = defineProps({
    expose: Object,
    id: String
});

</script>
<template>
    <div class="container-fluid p-0 h-100">

        <h4>Condition</h4>
        condtion list gets used in other triggers with same expose name - nedsfixing
        {{ props.expose.Name }} {{ props.id }}
        <div class="col">
            <Conditions :expose="props.expose" :exposeName="props.expose.Name">
            </Conditions>
        </div>
    </div>
</template>
