import { defineAsyncComponent } from 'vue';

type PanelKey = string;
type Map = { [key: PanelKey]: any };

export const PanelComponents: Map = {
  Trigger: defineAsyncComponent(() => import('../components/automations/Trigger.vue')),
  trigger: defineAsyncComponent(() => import('../components/automations/actions/TriggerAction.vue')),
  step: defineAsyncComponent(() => import('../components/automations/actions/StepAction.vue')),
  ActionEditor: defineAsyncComponent(() => import('../components/automations/actions/ActionEditor.vue')),
  ActionViewer: defineAsyncComponent(() => import('../components/automations/actions/ActionViewer.vue')),
  preset: defineAsyncComponent(() => import('../components/automations/actions/PresetRotationAction.vue')),
  Scheduler: defineAsyncComponent(() => import('../components/automations/Scheduler.vue')),
};
