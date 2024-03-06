<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent, computed, onMounted, Component, DefineComponent } from "vue";
type PanelKey = string
type Map = { [key: PanelKey]: any }

const componentMap: Map = {
    "Trigger": defineAsyncComponent(() =>
        import("../automations/Trigger.vue"),
    ),
    "TriggerAction": defineAsyncComponent(() =>
        import("../automations/actions/TriggerAction.vue"),
    ),
    "StepAction": defineAsyncComponent(() =>
        import("../automations/actions/StepAction.vue"),
    ),
};
const emit = defineEmits<{
    (e: 'save', item: any): void,
    (e: 'delete', item: any): void,
    (e: 'open', component_name: string): void,
    (e: 'close'): void,
}>()

const history = ref<Array<string>>();
const props = defineProps({
    component_name: {
        type: String,
        required: true
    },
    component_props: {
        type: Object,
        required: true
    }

});

function openComponent(component_name: string): void {
    history.value?.push(component_name)

    console.log("open component", component_name)
}

function closeComponent(): void {

}

const currentComponent = computed<PanelKey>((e) => { history.value?.slice(-1) as PanelKey });
function saveItem(item: any): void {
    emit('save', item);
}

function removeItem(item: any): void {
    emit('delete', item);
}

watch(
    () => props.component_name,
    () => {
        openComponent(props.component_name)
    }, { immediate: true }
)

onMounted(() => {

});

// component that we pass in can raise event to be changed


</script>

<template>
    PANEL : {{ currentComponent }}

    <component :is="componentMap[currentComponent]" v-bind="props.component_props" @delete="removeItem" @save="saveItem"
        @open="" @close="" />
</template>
