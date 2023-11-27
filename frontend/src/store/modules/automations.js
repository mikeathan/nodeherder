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
  save({ commit, dispatch, rootState }, automation) {
    console.log("automations/save");
    commit("add", automation)
    dispatch(
      "ws/emit",
      { event: "saveAutomation", message: automation },
      { root: true }
    );
  },
};

const getters = {
  items: (state) => state.items,
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
