const { defineConfig } = require("@vue/cli-service");
module.exports = defineConfig({
  transpileDependencies: ["vuetify"],

  devServer: {
    proxy: {
      "/apis": {
        target: process.env.VUE_APP_DEV_BACKEND,
      },
    },
    port: process.env.VUE_APP_DEV_PORT,
  },

  pluginOptions: {
    i18n: {
      locale: "zh",
      fallbackLocale: "en",
      localeDir: "locales",
      enableInSFC: true,
      includeLocales: false,
      enableBridge: true,
    },
  },
});
