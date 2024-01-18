import { InjectionKey } from "vue";
import { createStore, useStore as baseUseStore, Store } from "vuex";
import { DeviceMap } from "./types/device";

export interface State {
  devices: DeviceMap;
}
// define injection key
export const key: InjectionKey<Store<State>> = Symbol();

export const store_temp = createStore<State>({
  state: {
    devices: {},
  },
});

export function useStore_temp() {
  return baseUseStore(key);
}
