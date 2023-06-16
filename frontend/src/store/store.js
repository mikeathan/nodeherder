import { createStore } from "vuex";
import createPersistedState from "vuex-persistedstate";

const state = {
  devices: {},
};

const mutations = {
  deviceUpdated(state, device) {
    state.devices[device.name] = device;
  },

  init(state, devices) {
    devices.forEach((device) => {
      state.devices[device.name] = device;
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
};

const plugins = [createPersistedState()];
export default createStore({
  state,
  getters,
  mutations,
});
//plugins

//https://github.com/vuejs/vuex/blob/main/examples/composition/counter/store.js
