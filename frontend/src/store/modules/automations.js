const state = {
  initialized: false,
  items: {},
};

const actions = {
  save({ state, dispatch, rootState }, name) {
    console.log("automations/save");
    dispatch(
      "ws/emit",
      { event: "saveAutomation", message: state.items[name] },
      { root: true }
    );
  },
};

const getters = {
  items: (state) => state.items,
  isInitialized: (state) => state.initialized,
  find: (state) => (name) => {
    return state.items[name];
  },
};

const mutations = {
  init(state, items) {
    console.log("automations/init");
    items.forEach((item) => {
      state.items[item.name] = item;
    });
    state.initialized = true;
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
