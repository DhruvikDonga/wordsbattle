import { createApp } from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import { loadFonts } from './plugins/webfontloader'
import router from './router'
import { store } from "./store/index.js";
loadFonts()

createApp(App).use(router).use(vuetify).use(store).mount('#app')
