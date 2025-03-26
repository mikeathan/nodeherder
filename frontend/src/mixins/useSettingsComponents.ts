import { defineAsyncComponent } from 'vue';

type SettingsKey = string;
type Map = { [key: SettingsKey]: any };

export const ExposeSettingsComponents: Map = {
    rateLimit: defineAsyncComponent(() => import('../components/controls/TimeInterval.vue')),
    debounce: defineAsyncComponent(() => import('../components/settings/DebounceSettings.vue')),
};
