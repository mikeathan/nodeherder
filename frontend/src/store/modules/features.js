const state = {
  initialized: false,
  items: {},
};

const actions = {};

const getters = {
  items: (state) => state.items,
  isInitialized: (state) => state.initialized,
  find: (state) => (id) => {
    return state.items[id];
  },
};

const mutations = {
  init(state, items) {
    console.log("features/init");
    console.log(items);
    items.forEach((item) => {
      state.items[item.id] = item;
    });
    state.initialized = true;
  },

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
