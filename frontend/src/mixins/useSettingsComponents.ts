import { defineAsyncComponent } from 'vue';

type SettingsKey = string;
type Map = { [key: SettingsKey]: any };

export const DeviceConfigDefaultComponents: Map = {
  rateLimit: defineAsyncComponent(() => import('../components/controls/TimeInterval.vue')),
  debounce: defineAsyncComponent(() => import('../components/settings/ExposeDebounceEditor.vue')),
  defaultDebounceByCategory: defineAsyncComponent(() => import('../components/settings/ExposeCategoryDebounceEditor.vue')),
};

export const DeviceConfigOverrideComponents: Map = {
  rateLimit: defineAsyncComponent(() => import('../components/controls/TimeInterval.vue')),
  debounce: defineAsyncComponent(() => import('../components/settings/ExposeDebounceEditor.vue')),
  debounceOverrides: defineAsyncComponent(() => import('../components/settings/ExposeDebounceEditor.vue')),
};
