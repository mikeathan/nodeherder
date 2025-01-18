import { Module } from 'vuex';
import { RootState } from '../../state';
import { DeviceModuleState } from './state';
import { Device, Devices, DeviceMap, DeviceUpdate } from '../../../types/device';
import { KeyValuePair } from '../../../types/types';

export const DeviceModule: Module<DeviceModuleState, RootState> = {
  namespaced: true,

  state: () => ({ deviceMap: {} as DeviceMap }),

  getters: {
    listAll: (state: DeviceModuleState) => (): Devices => {
      return Object.values(state.deviceMap) as Devices;
    },
    find:
      (state: DeviceModuleState) =>
      (id: string): Device => {
        return state.deviceMap[id];
      },
    exists:
      (state: DeviceModuleState) =>
      (id: string): boolean => {
        return state.deviceMap[id] != null;
      },
  },

  mutations: {
    add(state: DeviceModuleState, device: Device) {
      state.deviceMap[device.id] = device;
    },
    updateList(state: DeviceModuleState, devices: Devices) {
      devices.forEach((device: Device) => {
        if (device.id in state.deviceMap) state.deviceMap[device.id] = device;
      });
    },

    update(state: DeviceModuleState, deviceUpdate: DeviceUpdate) {
      if (deviceUpdate.id in state.deviceMap == false) {
        console.error('device ', deviceUpdate.id, ' not found');
        return;
      }
      var device = state.deviceMap[deviceUpdate.id];
      for (var key in deviceUpdate.data) {
        if (key in device.exposes) {
          device.exposes[key].data = deviceUpdate.data[key];
        }
      }
      for (var key in deviceUpdate.properties) {
        if (key in device.properties) {
          device.properties[key] = deviceUpdate.properties[key];
        }
      }
      device.properties.last_seen = deviceUpdate.last_seen;
    },

    clear(state: DeviceModuleState) {
      Object.entries(state.deviceMap).forEach(([key, value]) => {
        delete state.deviceMap[key];
      });
    },
  },

  actions: {
    init({ state, commit }, devices: Devices) {
      commit('clear', state);
      devices.forEach((device: Device) => {
        state.deviceMap[device.id] = device;
      });
    },

    updateDevices({ commit }, devices: Devices) {
      commit('updateDevices', devices);
    },

        // TO MOVE !!!!!!!!!!!11 - ws

    setValue({ dispatch }, payload: KeyValuePair<any>) {
      dispatch('ws/emit', { event: 'deviceSetValue', message: payload }, { root: true });
    },

    // TO MOVE !!!!!!!!!!!11 - ws
    rename({ dispatch }, { name, newName }) {
      var payload = {
        from: name,
        to: newName,
      };

      dispatch('ws/emit', { event: 'deviceRename', message: payload }, { root: true });
    },

    // TO MOVE !!!!!!!!!!!11 - ws
    interview({ dispatch }, { id }) {
      dispatch('ws/emit', { event: 'deviceInterview', message: { id } }, { root: true });
    },
    // TO MOVE !!!!!!!!!!!11 - ws

    remove({ dispatch }, { id, force, block }) {
      var payload = {
        id: id,
        force: force,
        block: block,
      };
      dispatch('ws/emit', { event: 'deviceRemove', message: payload }, { root: true });
    },
  },
};
