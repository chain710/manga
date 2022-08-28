import { defineStore } from "pinia";
import Vue from "vue";

// a reactive hub; store state across multi components
export const useHub = defineStore("hub", {
  state: () => ({ tasks: [] }),
  actions: {
    addTask(t) {
      let wrapper = { done: false, promise: t };
      this.tasks.push(wrapper);
      return t.then(
        (ret) => {
          wrapper.done = true;
          this.purgeTasks();
          console.debug(`task done, left ${this.tasks.length}, ret`, ret);
          return ret;
        },
        (err) => {
          wrapper.done = true;
          console.log(`task error`, err);
        }
      );
    },
    purgeTasks() {
      while (this.tasks.length > 0 && this.tasks[0].done) {
        this.tasks.splice(0, 1);
      }
    },
  },
});

function extractError(err) {
  if (err && err.response && err.response.data) {
    return err.response.data.error || err.toString();
  } else {
    return err.toString();
  }
}

export default {
  install(vue) {
    Object.defineProperty(vue.prototype, "$hub", {
      get: function () {
        if (!this.__hub) {
          this.__hub = useHub();
        }

        return this.__hub;
      },
    });

    vue.prototype.__v = new Vue();
    vue.prototype.$hubon = function (event, handler) {
      this.__v.$on(event, handler);
    };
    vue.prototype.$ninfo = function (message, args) {
      this.__v.$emit("snack-message", { level: "info", message: this.$t(`message.info.${message}`, args) });
    };
    vue.prototype.$nerror = function (message, error, args) {
      const msgargs = Object.assign({ err: extractError(error) }, args);
      this.__v.$emit("snack-message", { level: "error", message: this.$t(`message.error.${message}`, msgargs) });
    };
    vue.prototype.$syncLibrary = function () {
      this.__v.$emit("sync-library");
    };
  },
};
