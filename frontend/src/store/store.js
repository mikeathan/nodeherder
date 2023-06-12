import { createStore } from "vuex";
import createPersistedState from "vuex-persistedstate";
// import { Device } from "../modules/sensors/device";
const state = {
  devices: {},
};

const mutations = {
  deviceUpdated(state, device) {
    state.devices[device.name] = device;
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
  plugins,
});

//https://github.com/vuejs/vuex/blob/main/examples/composition/counter/store.js
