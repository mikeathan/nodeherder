import { createStore } from "vuex";

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

export default createStore({
  state,
  getters,
  mutations,
});

//https://github.com/vuejs/vuex/blob/main/examples/composition/counter/store.js
