<template>
  <v-container fluid>
    <v-row align="center" justify="center">
      <div class="text-center">
        <h1 class="text-h5 mt-4">{{ $t("welcome.message") }}</h1>
        <p class="text-body-1">{{ $t("welcome.no_libraries_yet") }}</p>
        <v-btn color="primary" @click="addLibrary">{{ $t("welcome.add_library") }}</v-btn>
      </div>
    </v-row>
  </v-container>
</template>
<script>
export default {
  mounted() {
    if (this.$hub.libraries.length > 0) {
      this.$router.replace({ name: "dashboard" });
      return;
    }
    this.$emit("main-enter", {
      name: "welcome",
    });
  },
  methods: {
    addLibrary() {
      this.$hubemit("add-library");
    },
  },

  watch: {
    "$hub.libraries": function (val) {
      if (val.length > 0) {
        this.$router.replace({ name: "dashboard" });
      }
    },
  },
};
</script>
