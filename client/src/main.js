import { createApp } from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import { loadFonts } from './plugins/webfontloader'
import router from './router'
import { store } from "./store/index.js";
import VueGtag from "vue-gtag";

loadFonts()

const app = createApp(App).
    use(router).
    use(vuetify).
    use(store)

app.use(VueGtag,{    
    config: {id:'G-678VSVZR56'},
},router)

app.mount('#app')
