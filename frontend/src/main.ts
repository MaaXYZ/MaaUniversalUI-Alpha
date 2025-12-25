import { createApp } from 'vue'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import './style.css'
import './tailwind.css'
import App from './App.vue'
import router from './router'
import store from './store'
import i18n from './locales'

const app = createApp(App)

app.use(PrimeVue, {
  unstyled: true,
})
app.use(ToastService)
app.use(router)
app.use(store)
app.use(i18n as any)

app.mount('#app')
