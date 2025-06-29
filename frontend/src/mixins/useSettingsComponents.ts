import { defineAsyncComponent } from 'vue';

type SettingsKey = string;
type Map = { [key: SettingsKey]: any };

export const ExposeSettingsComponents: Map = {
  rateLimit: defineAsyncComponent(() => import('../components/controls/TimeInterval.vue')),
  debounce: defineAsyncComponent(() => import('../components/settings/DebounceSettings.vue')),
  debounceOverrides: defineAsyncComponent(() => import('../components/settings/DebounceSettings.vue')),
};

// export function deviceConfigOverrideComponent(key: any): any | undefined {
//   if (!(key in ExposeSettingsComponents)) return undefined;

//   const filtered: Map = Object.fromEntries(
//     Object.entries(ExposeSettingsComponents).filter(([k]) => k !== 'defaultDebounceByCategory')
//   );
//   return filtered[key];
// }

// export function deviceConfigDefaultsComponent(key: any): any | undefined {
//   if (!(key in ExposeSettingsComponents)) return undefined;

//   const filtered: Map = Object.fromEntries(
//     Object.entries(ExposeSettingsComponents).filter(([k]) => k !== 'debounceOverrides')
//   );
//   return filtered[key];
// }
