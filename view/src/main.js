import Vue from "vue";
import App from "./App.vue";
import vuetify from "./plugins/vuetify";
import router from "./router";
import lineClamp from "vue-line-clamp";

Vue.config.productionTip = false;
Vue.use(lineClamp);
new Vue({
  vuetify,
  router,
  render: (h) => h(App),
}).$mount("#app");
