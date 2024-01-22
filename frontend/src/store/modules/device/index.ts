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

    updateList(state: DeviceModuleState, devices: Devices) {
      devices.forEach((device: Device) => {
        if (device.id in state.devices)
          state.devices[device.id] = device;
      });
    },

    update(state, deviceUpdate: DeviceUpdate) {
      if (deviceUpdate.id in state.devices == false) {
        console.error("device ", deviceUpdate.id, " not found");
        return;
      }

      var device = state.devices[deviceUpdate.id];
      for (var key in deviceUpdate.data) {
        if (key in device.exposes[key]) {
          device.exposes[key].data = deviceUpdate.data[key];
        }
      }
      for (var key in deviceUpdate.properties) {
        if (key in device.properties[key]) {
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

    // TODO once wsclient store is finished

    // setValue({ dispatch }, payload) {
    //   dispatch(
    //     "ws/emit",
    //     { event: "deviceSetValue", message: payload },
    //     { root: true }
    //   );
    // },

    // rename({ dispatch }, { name, newName }) {
    //   var payload = {
    //     from: name,
    //     to: newName,
    //   };
    //   dispatch(
    //     "ws/emit",
    //     { event: "deviceRename", message: payload },
    //     { root: true }
    //   );
    // },
  },
};
