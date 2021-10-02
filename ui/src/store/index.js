import { createStore } from 'vuex'
import lsp from './modules/lsp'


const debug = process.env.NODE_ENV !== 'production'

export default createStore({
  modules: {
    lsp    
  },
  strict: debug  
})
