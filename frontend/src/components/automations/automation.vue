<script setup>
import { useStore } from "vuex";
import { computed, onMounted } from "vue";


const props = defineProps({
    item: Object,
});

onMounted(() => {
});

</script>

<template>
    <p>

        Description : {{ props.item.description }}

    <div v-for="(sensorTriggers, name) in props.item.sensor_triggers">
        Sensor: {{ name }}<br>

        <div v-for="(trigger) in sensorTriggers">

            <p>
            <h5>Condition:</h5>
            <div v-for="(condition) in trigger.conditions">
                {{ condition.name }} {{ condition.equalityoperator }} {{ condition.value }}
            </div>

            <div>
                <h5>Action:</h5>
                Friendlyname: {{ trigger.action.friendlyname }}<br>
                Type: {{ trigger.action.type }}<br>
                Name:{{ trigger.action.name }}<br>
                Value:{{ trigger.action.value }}<br>
                Delay:{{ trigger.action.delay }}<br>
            </div>
            </p>

        </div>


    </div>

    </p>
</template>
