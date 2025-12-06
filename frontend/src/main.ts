import { createApp } from 'vue'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import './style.css'
import './tailwind.css'
import App from './App.vue'
import router from './router'
import store from './store'

const app = createApp(App)

app.use(PrimeVue, {
  unstyled: true,
})
app.use(ToastService)
app.use(router)
app.use(store)

app.mount('#app')
