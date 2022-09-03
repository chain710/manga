import { defineStore } from "pinia";

const useSettings = defineStore("settings", {
  state: () => {
    return {
      readMode: "rtl",
      backgroundColor: "bg-black",
      alwaysFullScreen: false,
    };
  },
  persist: {
    key: "settings",
    storage: localStorage,
  },
});

export default {
  install(vue) {
    Object.defineProperty(vue.prototype, "$settings", {
      get: function () {
        if (!this.__settings) {
          this.__settings = useSettings();
        }

        return this.__settings;
      },
    });
  },
};
