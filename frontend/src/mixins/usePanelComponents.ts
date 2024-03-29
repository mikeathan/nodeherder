import { defineAsyncComponent } from "vue";

type PanelKey = string;
type Map = { [key: PanelKey]: any };

export const PanelComponents: Map = {
  Trigger: defineAsyncComponent(
    () => import("../components/automations/Trigger.vue")
  ),
  TriggerAction: defineAsyncComponent(
    () => import("../components/automations/actions/TriggerAction.vue")
  ),
  StepAction: defineAsyncComponent(
    () => import("../components/automations/actions/StepAction.vue")
  ),
  ActionEditor: defineAsyncComponent(
    () => import("../components/automations/actions/ActionEditor.vue")
  ),
  ActionViewer: defineAsyncComponent(
    () => import("../components/automations/actions/ActionViewer.vue")
  ),
  Dropdown: defineAsyncComponent(
    () => import("../components/automations/controls/Dropdown.vue")
  ),
  Button: defineAsyncComponent(
    () => import("../components/automations/controls/Button.vue")
  ),
};
