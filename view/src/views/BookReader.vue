<template>
  <v-container class="ma-0 pa-0 full-height bg-gray" fluid>
    <v-slide-y-transition>
      <v-toolbar
        dense
        elevation="1"
        v-if="showToolbars"
        class="settings full-width tool-bar-glass"
        style="position: fixed; top: 0"
      >
        <v-btn icon to="/">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <v-toolbar-title> booktitle </v-toolbar-title>
        <v-spacer></v-spacer>

        <v-btn icon>
          <v-icon>mdi-view-grid</v-icon>
        </v-btn>
        <v-btn icon @click="showToolbars = !showToolbars">
          <v-icon>mdi-help-circle</v-icon>
        </v-btn>
        <v-btn icon>
          <v-icon>mdi-cog</v-icon>
        </v-btn>
      </v-toolbar>
    </v-slide-y-transition>

    <!-- reader -->
    <div class="full-height">
      <paged-reader @center-click="showToolbars = !showToolbars"></paged-reader>
    </div>
    <v-slide-y-reverse-transition>
      <v-toolbar
        dense
        elevation="1"
        class="settings full-width tool-bar-glass"
        style="position: fixed; bottom: 0"
        v-if="showToolbars"
      >
        <v-row justify="center">
          <v-col class="px-0">
            <v-slider
              v-model="pageNumber"
              hide-details
              thumb-label
              class="align-center"
              min="1"
              :max="pagesCount"
            >
              <template v-slot:append>
                <v-label>{{ pagesCount }}</v-label>
              </template>
            </v-slider>
          </v-col>
        </v-row>
      </v-toolbar>
    </v-slide-y-reverse-transition>
  </v-container>
</template>
<script>
import PagedReader from "@/components/PageReader.vue";

export default {
  components: {
    PagedReader,
  },
  data: () => {
    return {
      showToolbars: true,
      pageNumber: 0,
      pagesCount: 987,
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
.settings {
  z-index: 2;
}

.full-height {
  height: 100%;
}

.full-width {
  width: 100%;
}

.bg-gray {
  background-color: gray;
}

.tool-bar-glass {
  background: rgba(255, 255, 255, 0.25) !important;
  box-shadow: 0 4px 30px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(5px);
  -webkit-backdrop-filter: blur(5px);
}
</style>
