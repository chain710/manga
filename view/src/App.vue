<template>
  <v-app>
    <router-view></router-view>
    <v-snackbar v-model="showMessage" timeout="3000" :color="color">
      {{ message }}
      <template v-slot:action="{ attrs }">
        <v-btn text v-bind="attrs" @click="showMessage = false">
          {{ $t("global.close") }}
        </v-btn>
      </template>
    </v-snackbar>
  </v-app>
</template>
<script>
export default {
  data() {
    return {
      showMessage: false,
      message: "",
      color: "grey darken-3",
    };
  },
  mounted() {
    this.$hubon("snack-message", this.showSnack);
  },

  methods: {
    showSnack(msg) {
      switch (msg.level) {
        case "error":
          this.message = msg.message;
          this.color = "error";
          break;
        case "info":
          this.message = msg.message;
          this.color = "grey darken-3";
          break;
        default:
          console.error("invalid app-message", msg);
          return;
      }
      this.showMessage = true;
    },
  },
  watch: {
    snackSeq() {},
  },
};
</script>
<style>
@import "styles/global.css";
@import "@mdi/font/css/materialdesignicons.css";
</style>
