<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent, computed, onMounted, inject } from "vue";
import { Emitter } from 'mitt'
import { EventActions, Events, OpenPanelEvent } from "@/types/events.type";


const emitter = inject('emitter') as Emitter<Events>;
emitter.on('openPanel', (e: OpenPanelEvent) => {

    console.log("openpanel received ", e)
    openComponent(e.name, e.args, e.events);
});

emitter.on('closePanel', (e: string) => { console.log("close panel event received . do nothing") });

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
    "Action": defineAsyncComponent(() =>
        import("../automations/actions/Action.vue"),
    ),
};

const emit = defineEmits<{
    //(e: 'save', item: any): void,
    (e: 'delete', item: any): void,
    (e: 'open', component_name: string): void,
    (e: 'close'): void,
}>()

const history = ref<Array<OpenPanelEvent>>([]);
const props = defineProps({
    component_name: {
        type: String,
        required: true
    },
    component_props: {
        type: Object,
        required: true
    },
    component_events: {
        type: Object,
        required: true
    }

});

function openComponent(component_name: string, args: any, events: EventActions): void {

    history.value.push({ name: component_name, args: args, events: events })
    console.log("open component: history=", currentComponent.value)
}

function closeComponent(): void {
    history.value.pop();
    console.log("close component: history=", history.value)

    if (history.value.length == 0) {
        emit('close');
    }
}

const currentComponent = computed(() => {
    const lastValue = history.value.at(-1)
    return lastValue === undefined ? { name: '', args: '', events: {} } : lastValue as OpenPanelEvent;
});

function saveItem(item: any): void {
    /// emit('save', item);
}

function removeItem(item: any): void {
    emit('delete', item);
}

watch(
    () => props.component_name,
    () => {
        openComponent(props.component_name, props.component_props, props.component_events)
    }, { immediate: true }
)

onMounted(() => {

});

// component that we pass in can raise event to be changed


</script>

<template>
    <!-- PANEL : {{ currentComponent }} <br> -->

    <div class="row">
        <button type="button" class="btn-close" aria-label="Close" @click="closeComponent"></button>
        <div class="col">
            <!-- @delete="removeItem"
        @save="saveItem" @open="openComponent" @close="closeComponent"  -->
            <component :is="componentMap[currentComponent.name]" v-bind="currentComponent.args"
                v-on="currentComponent.events" />
        </div>
    </div>

</template>


<!-- TODO

import mitt from 'mitt'

const emitter = mitt()

// listen to an event
emitter.on('foo', e => console.log('foo', e) )

// listen to all events
emitter.on('*', (type, e) => console.log(type, e) )

// fire an event
emitter.emit('foo', { a: 'b' })

// clearing all events
emitter.all.clear()

// working with handler references:
function onFoo() {}
emitter.on('foo', onFoo)   // listen
emitter.off('foo', onFoo)  // unlisten

------------------------
import mitt from 'mitt';

type Events = {
  foo: string;
  bar?: number;
};

const emitter = mitt<Events>(); // inferred as Emitter<Events>

emitter.on('foo', (e) => {}); // 'e' has inferred type 'string'

emitter.emit('foo', 42); // Error: Argument of type 'number' is not assignable to parameter of -->