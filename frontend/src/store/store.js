import { createStore } from "vuex";
import createPersistedState from "vuex-persistedstate";
import wsclient from "./modules/wsclient";
import automations from "./modules/automations";

const state = {
  devices: {},
};

const mutations = {
  deviceUpdated(state, device) {
    state.devices[device.id] = device;
  },

  initDevices(state, devices) {
    devices.forEach((device) => {
      state.devices[device.id] = device;
    });
  },
  dispose() {
    // call from wsclient on disconnect
    // clear devices
    // clear automations
  },
  clear() {
    console.log("clear devices");
    for (var prop in state.devices) {
      if (state.devices.hasOwnProperty(prop)) {
        console.log("deleting:" + prop);
        delete state.devices[prop];
      }
    }
  },
};

const getters = {
  devices: (state) => state.devices,
  findDevice: (state) => (id) => {
    return state.devices[id];
  },
};

//const plugins = [createPersistedState()];
const modules = {
  ws: wsclient,
  automations,
};

export default createStore({
  state,
  getters,
  mutations,
  modules,
});
//plugins

//https://github.com/vuejs/vuex/blob/main/examples/composition/counter/store.js
