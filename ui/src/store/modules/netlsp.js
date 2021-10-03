
const state = () => ({  
  lsp: {},
})

// getters
const getters = {
  lspToGet: (state) => {
    return JSON.parse(JSON.stringify(state.lsp));
  }
}

// actions
const actions = {
  saveLSP ({ commit }, lsp) {  
    console.log(JSON.stringify(lsp, null, 2));        
    commit('setLSP', lsp) 
  }
}

// mutations
const mutations = {
  setLSP (state, lsp) {
    state.lsp = lsp
  }  
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
