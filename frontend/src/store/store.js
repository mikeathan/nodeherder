import { createStore } from "vuex";
import createPersistedState from "vuex-persistedstate";
import wsclient from "./modules/wsclient";
import automations from "./modules/automations";
import devices from "./modules/devices";

const state = {};

const mutations = {};

const actions = {
  cleanup({ commit }) {
    commit("devices/clear", []);
    commit("automations/clear", []);
  },
};

const getters = {};

//const plugins = [createPersistedState()];
const modules = {
  ws: wsclient,
  automations,
  devices,
};

export default createStore({
  state,
  getters,
  actions,
  mutations,
  modules,
});
//plugins

//https://github.com/vuejs/vuex/blob/main/examples/composition/counter/store.js
