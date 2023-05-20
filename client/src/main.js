import { createApp } from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import { loadFonts } from './plugins/webfontloader'
import router from './router'
import { store } from "./store/index.js";
const VueAnalytics = require('vue-analytics')

loadFonts()

const app = createApp(App).
    use(router).
    use(vuetify).
    use(store)

app.use(VueAnalytics,{
    
    id: 'G-678VSVZR56',
    // If you're using vue-router, pass the router instance here.
    router,
})

app.mount('#app')
