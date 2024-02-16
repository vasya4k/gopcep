// import lspapi from '../../api/lspapi'

// initial state
// shape: { id, quantity }
const state = () => ({
    lspToAdd: {},
    lspToAddFailure: false
});

// getters
const getters = {
    lspToAdd: (state) => {
        console.log(state.lspToAdd);
        return JSON.parse(JSON.stringify(state.lspToAdd));
    }
};

// actions
const actions = {
    saveLSP({ commit }, lsp) {
        commit('setLSPToAdd', lsp);
    }
};

// mutations
const mutations = {
    setLSPToAdd(state, lsp) {
        state.lspToAdd = lsp;
    },
    setLSPAddFailed(state, status) {
        state.checkoutStatus = status;
    }
};

export default {
    namespaced: true,
    state,
    getters,
    actions,
    mutations
};
