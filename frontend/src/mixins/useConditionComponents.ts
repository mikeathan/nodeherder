import { defineAsyncComponent } from 'vue';

type ConditionKey = string;
type Map = { [key: ConditionKey]: any };

export const ConditionComponents: Map = {
  expose: defineAsyncComponent(() => import('../components/automations/conditions/ExposeCondition.vue')),
};
