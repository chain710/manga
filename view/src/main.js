import Vue from "vue";
import App from "./App.vue";
import vuetify from "./plugins/vuetify";
import service from "./plugins/service";
import convert from "./plugins/convert";
import hub from "./plugins/hub";
import vconst from "./plugins/const";
import settings from "./plugins/settings";
import router from "./router";
import lineClamp from "vue-line-clamp";
import i18n from "./i18n";
import { createPinia, PiniaVuePlugin } from "pinia";
import persistedstate from 'pinia-plugin-persistedstate'

Vue.config.productionTip = false;
Vue.use(PiniaVuePlugin);
Vue.use(lineClamp);
Vue.use(service);
Vue.use(convert);
Vue.use(vconst);
Vue.use(hub);
Vue.use(settings);
const pinia = createPinia();
pinia.use(persistedstate)
new Vue({
  router,
  vuetify,
  i18n,
  pinia,
  render: (h) => h(App),
}).$mount("#app");
