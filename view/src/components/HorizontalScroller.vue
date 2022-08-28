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
      :id="id"
      :ref="id"
      @scroll="computeScrollability"
      v-resize="computeScrollability">
      <div class="d-inline-flex" v-mutate="computeScrollability">
        <slot name="content" class="content"></slot>
      </div>
    </div>
  </div>
</template>
<script>
import _ from "lodash";
export default {
  name: "HorizontalScroller",
  data: () => {
    const id = _.uniqueId();
    return {
      id: id,
      canScrollBackward: false,
      canScrollForward: true,
      el: null,
    };
  },
  mounted: function () {
    this.el = this.$refs[this.id];
  },
  methods: {
    computeScrollability() {
      if (!this.el) {
        console.debug("no scroll container yet", this.id);
        return;
      }

      const sl = this.el.scrollLeft;
      const sw = this.el.scrollWidth;
      const cw = this.el.clientWidth;
      this.canScrollBackward = Math.round(sl) > 0;
      this.canScrollForward = Math.round(sl) + cw < sw;
      // console.log(
      //   `id=${this.id} sl=${sl} sw=${sw} cw=${cw} backward=${this.canScrollBackward} forward=${this.canScrollForward}`
      // );
    },
    doScroll(dir) {
      console.log(`scroll ${dir}`);
      if (!this.el) {
        return;
      }

      const adjustment = 100;
      const delta = this.el.clientWidth - adjustment;
      let scrollLeft = Math.round(this.el.scrollLeft)
      let target;
      if (dir == "backward") {
        target = scrollLeft - delta;
      } else if (dir == "forward") {
        target = scrollLeft + delta;
      }
      this.el.scrollTo({
        top: 0,
        left: target,
        behavior: "smooth",
      })
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
