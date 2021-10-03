import { createStore } from 'vuex'
import lsp from './modules/lsp'
import netlsp from './modules/netlsp'


const debug = process.env.NODE_ENV !== 'production'

export default createStore({
  modules: {
    lsp,
    netlsp    
  },
  strict: debug  
})
