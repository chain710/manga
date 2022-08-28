<template>
  <v-dialog v-model="show" scrollable @keydown.esc.stop="">
    <v-card :max-height="$vuetify.breakpoint.height * 0.9" dark>
      <v-card-title class="justify-center">
        <v-pagination
          v-model="explorerPage"
          :total-visible="perPage"
          :length="Math.ceil(thumbs.length / perPage)"></v-pagination>
      </v-card-title>
      <v-card-text>
        <v-container fluid>
          <v-row class="mb-2 align-center justify-space-around">
            <div
              v-for="thumb in visibleThumbs"
              :key="thumb.url"
              class="d-flex flex-column justify-center image-container">
              <v-img
                :src="thumb.url"
                aspect-ratio="0.7071"
                contain
                height="200"
                width="140"
                class="ma-2"
                @click="goTo(thumb.page)"
                style="cursor: pointer" />
              <div class="white--text text-center font-weight-bold">
                {{ thumb.page }}
              </div>
            </div>
          </v-row>
        </v-container>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  props: {
    value: { type: Boolean },
    // element of thumbs should be: {url, page}
    thumbs: { type: Array, required: true },
    perPage: { type: Number, default: 8 },
    page: { type: Number, default: 1 },
  },
  data() {
    return {
      explorerPage: 1,
      show: false,
    };
  },
  watch: {
    value(val) {
      this.show = val;
    },
    show(val) {
      this.$emit("input", val);
    },
    page: {
      handler(val) {
        this.explorerPage = Math.ceil(val / this.perPage);
      },
      immediate: true,
    },
  },
  computed: {
    visibleThumbs() {
      const a = (this.explorerPage - 1) * this.perPage;
      const b = this.explorerPage * this.perPage;
      return this.thumbs.slice(a, b);
    },
  },
  methods: {
    goTo(page) {
      this.$emit("go", page);
    },
  },
};
</script>
<style scoped>
.image-container {
  min-height: 220px;
  max-width: 140px;
}
</style>
