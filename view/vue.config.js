const { defineConfig } = require("@vue/cli-service");
module.exports = defineConfig({
  transpileDependencies: ["vuetify"],
  publicPath: ".",
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

  pages: {
    index: {
      entry: "src/main.js",
      title: "MangaDepot",
    },
  },

  pwa: {
    name: "MangaDepot",
    assetsVersion: "1.2",
    themeColor: "#ffffff",
    msTileColor: "#da532c",
    appleMobileWebAppCapable: "yes",
    appleMobileWebAppStatusBarStyle: "black",
    manifestCrossorigin: "use-credentials",
    workboxOptions: {
      exclude: ["index.html"],
    },
    manifestOptions: {
      icons: [
        {
          src: "android-chrome-192x192.png",
          sizes: "192x192",
          type: "image/png",
        },
        {
          src: "android-chrome-512x512.png",
          sizes: "512x512",
          type: "image/png",
        },
      ],
    },
    iconPaths: {
      faviconSVG: "favicon.svg",
      favicon32: "favicon-32x32.png",
      favicon16: "favicon-16x16.png",
      appleTouchIcon: "apple-touch-icon.png",
      maskIcon: "safari-pinned-tab.svg",
      msTileImage: null,
    },
  },
});
