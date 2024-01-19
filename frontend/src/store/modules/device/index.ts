import { Module, GetterTree } from "vuex";
import { RootState } from "../../types/state";
import { DeviceModuleState } from "./state";
import { Device } from "../../../contracts/device";

////
export interface Getters extends GetterTree<DeviceModuleState, RootState> {
  find(state: DeviceModuleState, id: string): Device;
}

const getters: Getters = {
  find(state: DeviceModuleState, id: string): Device {
    return state.devices[id];
  },
};
///
export const DeviceModule: Module<DeviceModuleState, RootState> = {
  namespaced: true,
  state: () => ({ devices: {} }), need to define it correctly
  getters: {
    find(state: DeviceModuleState, id: string): Device {
      try {
        return state.devices[id];
      } catch (error) {
        console.log(error);
        return state.devices[id]; // TODO: return emtpy object
      }
    },
  },

  mutations: {
    add(state: DeviceModuleState, device: Device) {
      state.devices[device.id] = device;
    },

    updateDevices(state: DeviceModuleState, devices: Array<Device>) {
      devices.forEach((updated) => {
        const device: Device = state.devices[updated.id];
        if (device != undefined) {
          state.devices[updated.id] = updated;
        }
      });
    },
    // update(state, payload) {
    //     var device = state.items[payload.id];
    //     if (device == undefined) {
    //         console.error("device ", payload.id, " not found");
    //         return;
    //     }
    //     for (var key in payload.data) {
    //         if (device.exposes.hasOwnProperty(key)) {
    //             device.exposes[key].data = payload.data[key];
    //         }
    //     }
    //     for (var key in payload.properties) {
    //         if (device.properties.hasOwnProperty(key)) {
    //             device.properties[key] = payload.properties[key];
    //         }
    //     }
    //     device.properties.last_seen = payload.last_seen;
    // },
    clear(state: DeviceModuleState) {
      for (var id in state.devices) {
        delete state.devices[id];
      }
    },
  },
  actions: {
    init({ state, commit }, devices: Array<Device>) {
      // TODO: convert array<device> to type
      commit("clear", state);
      devices.forEach((device) => {
        state.devices[device.id] = device;
      });
    },
    updateDevices({ commit }, devices: Array<Device>) {
      commit("updateDevices", devices);
    },
  },
};
