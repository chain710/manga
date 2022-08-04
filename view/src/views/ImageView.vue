<template>
  <v-app>
    <Transition name="fade">
      <v-app-bar app v-if="showNav" class="app-bar">
        <v-app-bar-nav-icon to="/">
          <v-icon>mdi-arrow-left</v-icon>
        </v-app-bar-nav-icon>
        <v-toolbar-title>book xx page xx</v-toolbar-title>
        <v-spacer></v-spacer>
        <v-btn icon @click="showDebug = !showDebug">
          <v-icon>mdi-heart</v-icon>
        </v-btn>
      </v-app-bar>
    </Transition>
    <v-container fluid @click="showNav = !showNav">
      <v-img src="Jjdjr01-001.jpg"></v-img>
    </v-container>

    <v-bottom-sheet
      hide-overlay
      persistent
      no-click-animation
      v-model="showNav"
    >
      <v-sheet class="text-center">
        <v-slider v-model="pageNumber" thumb-label ticks></v-slider>
      </v-sheet>
    </v-bottom-sheet>
  </v-app>
</template>

<script>
export default {
  name: "ImageView",

  data: () => {
    return {
      showNav: true,
      showDebug: true,
      pageNumber: 0,
    };
  },
  created: function () {
    console.log("created", this.$route.params.bid, this.$route.params.vid);
  },
  mounted: function () {
    console.log("mounted", this.$route.params.bid, this.$route.params.vid);
  },
};
</script>
<style scoped>
.app-bar {
  background: rgba(255, 255, 255, 0.25) !important;
  border-radius: 16px;
  box-shadow: 0 4px 30px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(5px);
  -webkit-backdrop-filter: blur(5px);
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.25s ease-in-out;
}

.fade-enter,
.fade-leave-to {
  transform: translateY(-100%) !important;
  opacity: 0;
}
</style>
