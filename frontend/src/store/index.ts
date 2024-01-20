import { InjectionKey } from "vue";
import { createStore, useStore as baseUseStore, Store } from "vuex";
import { RootState } from "./state";
import { DeviceModule as devices } from "./modules/device/index";

// https://blog.openreplay.com/integrate-vuex-and-typescript/

export const key: InjectionKey<Store<RootState>> = Symbol();

export const store_temp = createStore<RootState>({
  state: {},
  modules: {
    devices,
  },
});

export function useStore_temp() {
  return baseUseStore(key);
}
