import { createStore } from "vuex";
import createPersistedState from "vuex-persistedstate";
import wsclient from "./modules/wsclient";

const state = {
  devices: {},
  automations: {},
};

const mutations = {
  deviceUpdated(state, device) {
    state.devices[device.id] = device;
  },

  init(state, devices) {
    devices.forEach((device) => {
      state.devices[device.id] = device;
    });
  },

  initAutomations(state, automations) {
    automations.forEach((automation) => {
      state.automations[automation.name] = automation;
    });
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
  automations: (state) => state.automations,
  findDevice: (state) => (id) => {
    return state.devices[id];
  },
};

//const plugins = [createPersistedState()];
const modules = {
  ws: wsclient,
};

export default createStore({
  state,
  getters,
  mutations,
  modules,
});
//plugins

//https://github.com/vuejs/vuex/blob/main/examples/composition/counter/store.js
