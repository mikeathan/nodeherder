<script setup lang="ts">
import {
  ref,
  watch,
  PropType,
  computed,
  onUnmounted,
} from 'vue';
import { OpenPanelEvent } from '@/types/events.type';
import { useAutomationEvents } from '@/mixins/useAutomationsEventBus';
import { PanelComponents } from '@/mixins/usePanelComponents';
import { KeyValuePair } from '@/types/types';

const cleanup = useAutomationEvents({
  openPanel(e: OpenPanelEvent) {
    openComponent(e);
  },
  closePanel(name: string) {
    closeComponent(name);
  },
  closeLastPanel() {
    closeLastComponent();
  },
});

const emit = defineEmits<{
  (e: 'open', component_name: string): void;
  (e: 'close'): void;
  (e: 'componentDisplayed'): void;

}>();

const componentCache = ref<KeyValuePair<OpenPanelEvent>>(
  {},
);
const presentationQueue = ref<Array<string>>([]);

function openComponent(event: OpenPanelEvent): void {

  presentationQueue.value.push(event.name);
  if (presentationQueue.value.length == 1) {
    emit('componentDisplayed')
  }

  if (componentCache.value[event.name] != undefined) {
    componentCache.value[event.name].args = event.args;
  }

  componentCache.value[event.name] = event;
}

const currentComponent = computed(() => {
  const lastValue = presentationQueue.value.at(-1);
  return lastValue === undefined ? '' : lastValue;
});

function closeComponent(name: string): void {
  console.log('panel close clicked');
  if (componentCache.value[name] === undefined) {
    return;
  }

  // TODO: cleanup componentCache ?
  presentationQueue.value.pop();
  const event = componentCache.value[name];


  if (presentationQueue.value.length == 0) {
    emit('close');
  }
}

function closeLastComponent(): void {

  // TODO: cleanup componentCache ?
  presentationQueue.value.pop();

  if (presentationQueue.value.length == 0) {
    emit('close');
  }
}
onUnmounted(() => {

  cleanup();
  presentationQueue.value = [];
  componentCache.value = {};
});
</script>

<template>

  <component v-if="currentComponent != ''" :is="PanelComponents[currentComponent]"
    v-bind="componentCache[currentComponent].args" v-on="componentCache[currentComponent].events" />
</template>
