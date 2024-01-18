import { InjectionKey } from "vue";
import { GetterTree, createStore, useStore as baseUseStore, Store } from "vuex";
import { DeviceMap } from "./types/device";
import { Device } from "../contracts/device";

export interface State {
  devices: DeviceMap;
}
// define injection key
export const key: InjectionKey<Store<State>> = Symbol();

export interface Getters extends GetterTree<State, State> {
  find(state: State): Device;
}

const getters: Getters = {
  find(state: State): Device {
    return state.devices[0];
  },
};

export const store_temp = createStore<State>({
  state: {
    devices: {},
  },
  getters,
  mutations: {
    add(state: State, device: Device) {
      state.devices[device.id] = device;
    },
  },
  actions: {
    incrementAsync({ commit }, device: Device) {
      commit("add", device);
    },
  },
  modules: {},
});

export function useStore_temp() {
  return baseUseStore(key);
}
