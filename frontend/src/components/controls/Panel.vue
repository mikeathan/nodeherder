<script setup lang="ts">
import { ref, watch, PropType, computed, onUnmounted } from "vue";
import { OpenPanelEvent } from "@/types/events.type";
import { useAutomationEvents } from "@/mixins/useAutomationsEventBus";
import { PanelComponents } from "@/mixins/usePanelComponents";
import { KeyyValuePair } from "@/types/types";

const cleanup = useAutomationEvents({
    openPanel(e: OpenPanelEvent) {
        openComponent(e);
    },
    closePanel(name: string) {
        closeComponent(name)
    },
})

const emit = defineEmits<{
    (e: 'open', component_name: string): void,
    (e: 'close'): void,
}>()

const componentCache = ref<KeyyValuePair<OpenPanelEvent>>({});
const presentationQueue = ref<Array<string>>([]);

const props = defineProps({
    item: {
        type: Object as PropType<OpenPanelEvent>,
        required: true
    },
});

watch(
    () => props.item,
    () => {
        openComponent(props.item)
    }, { immediate: true }
)

function openComponent(event: OpenPanelEvent): void {

    presentationQueue.value.push(event.name);

    console.log("OPEN:" + event.name, " Size ", presentationQueue.value.length);

    if (componentCache.value[event.name] != undefined) {
        componentCache.value[event.name].args = event.args;

    } else {
        componentCache.value[event.name] = event;
    }
}

const currentComponent = computed(() => {
    const lastValue = presentationQueue.value.at(-1)
    return lastValue === undefined ? '' : lastValue;
});


function closeComponent(name: string): void {

    if (componentCache.value[name] === undefined) {
        console.log("CLOSE ", name, " NOT FOUND")
        return
    }

    // TODO: cleanup componentCache ?
    presentationQueue.value.pop();
    const event = componentCache.value[name];

    console.log("CLOSE:" + event.name, " Size ", presentationQueue.value.length);
    if (presentationQueue.value.length == 0) {
        emit('close');
    }
}


onUnmounted(() => {
    console.log("Panel unmounted - deregister eventBus messages")
    cleanup();
    presentationQueue.value = [];
    componentCache.value = {};
});


</script>

<template>
    <div class="row">
        <button type="button" class="btn-close" aria-label="Close" @click="closeComponent(currentComponent)"></button>
        <div class="col">
            <component :is="PanelComponents[currentComponent]" v-bind="componentCache[currentComponent].args"
                v-on="componentCache[currentComponent].events" />
        </div>
    </div>
</template>
