const state = {
  initialized: false,
  items: {},
};

const actions = {
  init({ state, commit }, items) {
    commit("clear", []);

    console.log("features/init");
    items.forEach((item) => {
      state.items[item.id] = item;
    });
    state.initialized = true;
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
  clear(state) {
    console.log("clear features");
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
