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
import persistedstate from "pinia-plugin-persistedstate";
import VueRouter from "vue-router";

Vue.config.productionTip = false;
Vue.use(PiniaVuePlugin);
const pinia = createPinia();
pinia.use(persistedstate);
Vue.use(VueRouter);
Vue.use(lineClamp);
Vue.use(service);
Vue.use(convert);
Vue.use(vconst);
Vue.use(hub);
Vue.use(settings);

new Vue({
  pinia,
  router,
  vuetify,
  i18n,
  render: (h) => h(App),
}).$mount("#app");
