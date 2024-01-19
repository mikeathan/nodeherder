import { Module, createStore } from "vuex";

interface Automation {
    name: string;
}


export interface State { // test
    automations: Automation | null;


}

export interface AutomationModuleState {
    automation: Automation | null;
}

const userModule: Module<AutomationModuleState, State> = {
    state: () => ({ automation: null }),
    mutations: {
        setAutomation(state: AutomationModuleState, automation: Automation) {
            state.automation = automation;
        },
    },
    actions: {
        something_test({ commit }, user) {
            commit("setAutomation", user);
        },
    },
};