<template>
  <div style="position: relative">
    <div style="min-height: 36px">
      <slot name="prepend"></slot>
    </div>
    <div class="top-button-bar">
      <v-btn icon :disabled="!canScrollBackward" @click="doScroll('backward')">
        <v-icon>mdi-chevron-left</v-icon>
      </v-btn>
      <v-btn icon :disabled="!canScrollForward" @click="doScroll('forward')">
        <v-icon>mdi-chevron-right</v-icon>
      </v-btn>
    </div>

    <div
      class="scrolling-wrapper"
      @scroll="computeScrollability"
      v-resize="computeScrollability"
    >
      <div class="d-inline-flex" v-mutate="computeScrollability">
        <slot name="content" class="content"></slot>
        <slot name="content-append"></slot>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  name: "HorizontalScroller",
  data: () => ({
    canScrollBackward: false,
    canScrollForward: true,
  }),
  methods: {
    computeScrollability() {
      console.log("computeScrollability");
    },
    doScroll(dir) {
      console.log(`scroll ${dir}`);
    },
  },
};
</script>
<style scoped>
.top-button-bar {
  position: absolute;
  top: 0;
  right: 0;
}

.scrolling-wrapper {
  -webkit-overflow-scrolling: touch;
  display: flex;
  flex-wrap: nowrap;
  overflow-x: auto;
  scrollbar-width: none;
}

.scrolling-wrapper::-webkit-scrollbar {
  display: none;
}
.content {
  flex: 0 0 auto;
}
</style>
