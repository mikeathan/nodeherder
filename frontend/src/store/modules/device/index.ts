import { Module } from "vuex";
import { RootState } from "../../state";
import { DeviceModuleState } from "./state";
import { Device, Devices, DeviceMap } from "../../../types/device";

export const DeviceModule: Module<DeviceModuleState, RootState> = {
  namespaced: true,

  state: () => ({ devices: {} as DeviceMap }),

  getters: {
    list: (state: DeviceModuleState) => state.devices, // TODO that needs to return only the values !!!!!!!!!!!!!!11
    find: (state: DeviceModuleState) => (id: string) => {
      return state.devices[id];
    },
  },

  mutations: {
    add(state: DeviceModuleState, device: Device) {
      state.devices[device.id] = device;
    },

    updateDevices(state: DeviceModuleState, devices: Devices) {
      devices.forEach((updated: Device) => {
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
    init({ state, commit }, devices: Devices) {
      commit("clear", state);
      devices.forEach((device: Device) => {
        state.devices[device.id] = device;
      });
    },

    updateDevices({ commit }, devices: Devices) {
      commit("updateDevices", devices);
    },
  },
};
