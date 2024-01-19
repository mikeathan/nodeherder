import { InjectionKey } from "vue";
import { GetterTree, createStore, useStore as baseUseStore, Store } from "vuex";
import { RootState } from "./types/state";
import { Device } from "../contracts/device";
import { DeviceModule } from "./modules/device/index";
import { DeviceAutomation } from "../contracts/automations";

// https://blog.openreplay.com/integrate-vuex-and-typescript/

export const key: InjectionKey<Store<RootState>> = Symbol();

export const store_temp = createStore<RootState>({
  state: {},
  getters: {},
  modules: {
    DeviceModule,
  },
});

export function useStore_temp() {
  return baseUseStore(key);
}
