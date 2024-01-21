import { Module } from "vuex";
import { RootState } from "../../state";
import { DeviceModuleState } from "./state";
import {
  Device,
  Devices,
  DeviceMap,
  DeviceUpdate,
} from "../../../types/device";

export const DeviceModule: Module<DeviceModuleState, RootState> = {
  namespaced: true,

  state: () => ({ devices: {} as DeviceMap }),

  getters: {
    list: (state: DeviceModuleState) => (): Devices => {
      return Object.values(state.devices) as Devices; // TODO: convert that to typed
    },
    find:
      (state: DeviceModuleState) =>
      (id: string): Device => {
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

    update(state, deviceUpdate: DeviceUpdate) {
      var device = state.devices[deviceUpdate.id];
      if (device == undefined) {
        console.error("device ", deviceUpdate.id, " not found");
        return;
      }

      for (var key in deviceUpdate.data) {
        if (device.exposes[key] != undefined) {
          device.exposes[key].data = deviceUpdate.data[key];
        }
      }
      for (var key in deviceUpdate.properties) {
        if (device.properties[key] != undefined) {
          device.properties[key] = deviceUpdate.properties[key];
        }
      }
      device.properties.last_seen = deviceUpdate.last_seen;
    },

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
