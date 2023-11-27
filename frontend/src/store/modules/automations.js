const state = {
  initialized: false,
  items: {},
};

const actions = {
  init({ state, commit }, items) {
    commit("clear", []);

    console.log("automations/init");
    items.forEach((item) => {
      state.items[item.id] = item;
    });
    state.initialized = true;
  },
  save({ state, dispatch, rootState }, id) {
    console.log("automations/save");
    dispatch(
      "ws/emit",
      { event: "saveAutomation", message: state.items[id] },
      { root: true }
    );
  },
};

const getters = {
  items: (state) => state.items,
  name: (state) => (id) => {
    var item = state.items[id];
    if (item == undefined) {
      return {};
    }
    return item.friendlyName;
  },
  description: (state) => (id) => {
    var item = state.items[id];
    return item.description;
  },
  enabled: (state) => (id) => {
    var item = state.items[id];
    return item.enabled;
  },
  action: (state) => (id) => {
    var item = state.items[id];
    return item.action;
  },
  conditions: (state) => (id) => {
    var item = state.items[id];
    return item.conditions;
  },
  isInitialized: (state) => state.initialized,
  find: (state) => (id) => {
    return state.items[id];
  },
};

const mutations = {
  add(state, automation) {
    console.log("automations/add");
    state.items[automation.id] = automation;
  },
  update(state, automation) {
    console.log("automations/update");
    state.items[automation.id] = automation;
  },
  clear(state) {
    console.log("clear automations");
    for (var prop in state.items) {
      if (state.items.hasOwnProperty(prop)) {
        delete state.items[prop];
      }
    }
    state.initialized = false;
  },
};

export default {
  namespaced: true,
  state,
  mutations,
  actions,
  getters,
};
