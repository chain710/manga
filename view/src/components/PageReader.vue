<template>
  <div v-resize="onResize" v-touch="touch">
    <v-carousel
      :show-arrows="false"
      :continuous="false"
      :reverse="flipDirection"
      :vertical="vertical"
      hide-delimiters
      touchless
      height="100%"
      v-model="carouselPage">
      <!--  Carousel: pages  -->
      <v-carousel-item
        v-for="(spread, i) in spreads"
        :key="`spread${i}`"
        class="full-height"
        :eager="eagerLoad(i)"
        active-class="active-carousel">
        <div :class="`full-height d-flex flex-row${flipDirection ? '-reverse' : ''} px-0 mx-0`">
          <div
            :ref="`container${i}`"
            class="d-flex"
            :class="`${carouselItems[i].containerClass} ${imageContainerTransition}`"
            :style="{
              transform: `translate(${imageTransformX}px, ${imageTransformY}px) rotate(${imageRotate}deg)`,
            }">
            <!--div for measurement: with, height-->
            <img
              v-for="(page, j) in spread"
              :alt="page.name"
              :key="`spread${i}-${j}`"
              :src="page.url"
              @load="imageLoaded($event, i, j)"
              :class="carouselItems[i].imageClass"
              class="img-fit-all" />
          </div>
        </div>
      </v-carousel-item>
    </v-carousel>
    <!--  clickable zone: left  -->
    <div v-if="!vertical" @click="turnLeft()" class="left-quarter" style="z-index: 1" />

    <!--  clickable zone: right  -->
    <div v-if="!vertical" @click="turnRight()" class="right-quarter" style="z-index: 1" />

    <!--  clickable zone: top  -->
    <div v-if="vertical" @click="prev()" class="top-quarter" style="z-index: 1" />

    <!--  clickable zone: bottom  -->
    <div v-if="vertical" @click="next()" class="bottom-quarter" style="z-index: 1" />

    <!--  clickable zone: menu  -->
    <div
      @click="centerClick()"
      :class="`${vertical ? 'center-vertical' : 'center-horizontal'}`"
      style="z-index: 1"></div>
  </div>
</template>
<script>
const Shortcuts = {
  " ": {
    display: "Space",
    exec: (ctx) => {
      ctx.turnNext();
    },
  },
  ArrowLeft: {
    display: "←",
    exec: (ctx) => {
      ctx.turnLeft();
    },
  },
  PageDown: {
    display: "PgDn",
    exec: (ctx) => {
      ctx.next();
    },
  },
  ArrowRight: {
    display: "→",
    exec: (ctx) => {
      ctx.turnRight();
    },
  },
  PageUp: {
    display: "PgUp",
    exec: (ctx) => {
      ctx.prev();
    },
  },
};

const tolerationWidth = 20;

export default {
  data: function () {
    return {
      carouselPage: 0,
      carouselItems: [],
      vertical: false,
      windowWidth: window.innerWidth,
      windowHeight: window.innerHeight,
      touch: {
        // left: this.turnLeft,
        // right: this.turnRight,
        move: this.onTouchMove,
        start: this.onTouchStart,
        end: this.onTouchEnd,
        startTransformX: 0,
        startTransformY: 0,
        disableTransition: false,
      },
      imageTransformX: 0,
      imageTransformY: 0,
    };
  },
  props: {
    spreads: { type: Array, required: true },
    page: { type: Number, required: true },
    readMode: { type: String, required: true },
    imageRotate: { type: Number, default: 0 },
  },
  created() {
    console.debug("create page reader");
    window.addEventListener("keydown", this.keyPressed);
  },
  destroyed() {
    console.debug("destroy page reader");
    window.removeEventListener("keydown", this.keyPressed);
  },
  methods: {
    keyPressed(event) {
      Shortcuts[event.key]?.exec(this);
    },
    turnNext: function () {
      let changePage = this.flipDirection ? this.moveWindowLeft() : this.moveWindowRight();
      if (changePage) {
        this.next();
      }
    },
    turnLeft: function () {
      if (this.moveWindowLeft()) {
        this.flipDirection ? this.next() : this.prev();
      }
    },
    turnRight: function () {
      if (this.moveWindowRight()) {
        this.flipDirection ? this.prev() : this.next();
      }
    },
    next: function () {
      if (!this.hasNext()) {
        this.$emit("jump-next-stop");
        return;
      }
      this.setCarouselPage(this.carouselPage + 1);
      this.emitUpdatePageEvent();
    },
    prev: function () {
      if (!this.hasPrev()) {
        this.$emit("jump-prev-stop");
        return;
      }
      this.setCarouselPage(this.carouselPage - 1);
      this.emitUpdatePageEvent();
    },
    hasNext: function () {
      return this.carouselPage < this.spreads.length - 1;
    },
    hasPrev: function () {
      return this.carouselPage > 0;
    },
    centerClick: function () {
      this.$emit("center-click");
    },
    eagerLoad(idx) {
      const val = Math.abs(this.carouselPage - idx) <= 2;
      return val;
    },
    emitUpdatePageEvent() {
      const sp = this.spreads[this.carouselPage];
      const page = sp[sp.length - 1];
      this.$emit("update:page", page.page);
      this.$emit("update:image-rotate", 0);
    },
    toSpreadIndex(page) {
      for (let i = 0; i < this.spreads.length; i++) {
        for (let j = 0; j < this.spreads[i].length; j++) {
          if (this.spreads[i][j].page === page) {
            return i;
          }
        }
      }
      return page - 1;
    },
    imageLoaded(event, i, j) {
      let image = event.target;
      this.carouselItems[i].pages[j] = { width: image.naturalWidth, height: image.naturalHeight };
      console.debug(`image[${i}][${j}] load w=${image.naturalWidth} h=${image.naturalHeight}`);
      this.updateCarouselClass(i);
    },
    getCarouselDimensions(item) {
      let width = 0;
      let height = 0;
      for (const page of item.pages) {
        // ltr and rtl
        if (page.width == null) {
          return; // not load yet
        }
        width += page.width;
        if (page.height > height) {
          height = page.height;
        }
      }

      return { width, height };
    },
    // compute container & image best class by width,height
    updateCarouselClass(i) {
      let item = this.carouselItems[i];
      if (!item) {
        // maybe carouselItems not computed yet
        return;
      }
      const d = this.getCarouselDimensions(item);
      if (!d) {
        return;
      }
      const windowAspect = this.windowWidth / this.windowHeight;
      const carouselAspect = d.width / d.height;
      if (carouselAspect > windowAspect) {
        item.imageClass = "img-fit-height";
        item.containerClass = "justify-start";
      } else {
        item.imageClass = "img-fit-screen";
        item.containerClass = "justify-center";
      }

      console.debug(`update carousel[${i}] class=${item.imageClass}`);
      this.$set(this.carouselItems, i, item);
    },
    onResize() {
      const p = window.innerWidth / this.windowWidth;
      this.windowWidth = window.innerWidth;
      this.windowHeight = window.innerHeight;
      // adjust image scroll according to resize scale
      let imageTransformX = this.imageTransformX * p;
      let imageTransformY = this.imageTransformY * p;
      this.transformImageContainer(imageTransformX, imageTransformY);
      this.updateCarouselClass(this.carouselPage);
    },
    onTouchMove(event) {
      const dx = event.touchmoveX - event.touchstartX;
      // const dy = event.touchmoveY - event.touchstartY;
      this.transformImageContainer(this.touch.startTransformX + dx, 0);
    },
    onTouchStart() {
      this.touch.startTransformX = this.imageTransformX;
      this.touch.startTransformY = this.imageTransformY;
      this.touch.disableTransition = true;
    },
    onTouchEnd() {
      this.touch.disableTransition = false;
    },
    transformImageContainer(x, y) {
      const containers = this.$refs[`container${this.carouselPage}`];
      if (containers.length <= 0) {
        return;
      }
      let container = containers[0];
      const transformXMax = this.flipDirection
        ? Math.max(0, container.scrollWidth - this.windowWidth)
        : 0;
      const transformXMin = this.flipDirection
        ? 0
        : Math.min(0, this.windowWidth - container.scrollWidth);
      const tx = Math.max(transformXMin, Math.min(transformXMax, x));
      const ty = y; // should with max, min
      const dx = this.imageTransformX - tx;
      const dy = this.imageTransformY - ty;
      if (tx != this.imageTransformX || ty != this.imageTransformY) {
        console.debug(`image container tranform(${tx},${ty})`);
        this.imageTransformX = tx;
        this.imageTransformY = ty;
      }
      return { dx, dy };
    },
    moveWindowLeft() {
      // moveWindowLeft = swipeRight, dx > 0
      let d = this.transformImageContainer(this.windowWidth + this.imageTransformX, 0);
      return Math.abs(d.dx) < tolerationWidth;
    },
    moveWindowRight() {
      let d = this.transformImageContainer(-this.windowWidth + this.imageTransformX, 0);
      // console.log(`move right ${bound.min} ${this.imageTransformX}`);
      return Math.abs(d.dx) < tolerationWidth;
    },

    setCarouselPage(val) {
      if (val == this.carouselPage) {
        return;
      }

      this.imageTransformX = 0;
      this.imageTransformY = 0;
      this.touch.startTransformX = 0;
      this.touch.startTransformY = 0;
      this.updateCarouselClass(val);
      this.carouselPage = val;
    },
  },
  watch: {
    page: {
      handler(val) {
        this.setCarouselPage(this.toSpreadIndex(val));
        console.log(`set carouselPage = ${this.carouselPage} from page ${val}`);
      },
      immediate: true,
    },
    spreads: {
      handler(val) {
        console.debug("compute carouselClass");
        let carouselItems = [];
        for (const pages of val) {
          carouselItems.push({
            pages: Array(pages.length).fill({ width: null, height: null }),
            imageClass: "img-fit-original", // use fit-original to calculate width & height
            containerClass: "justify-center",
          });
        }
        this.carouselItems = carouselItems;
      },
      immediate: true,
    },
  },

  computed: {
    imageContainerTransition() {
      return this.touch.disableTransition ? "" : "image-container-transision";
    },
    flipDirection() {
      return this.readMode === "rtl";
    },
  },
};
</script>
<style scoped>
.full-height {
  height: 100%;
}

.img-fit-all {
  object-fit: contain;
  object-position: center;
}

.image-container-transision {
  transition: transform 0.4s;
}

.img-fit-width {
  width: 100vw;
  min-height: 100vh;
  align-self: flex-start;
}

.img-fit-height {
  min-height: 100vh;
  height: 100vh;
}

.img-fit-screen {
  width: 100vw;
  height: 100vh;
}

.img-fit-original {
  width: auto;
  height: auto;
}

.left-quarter {
  top: 0;
  left: 0;
  width: 25%;
  height: 100%;
  position: absolute;
}

.right-quarter {
  top: 0;
  right: 0;
  width: 25%;
  height: 100%;
  position: absolute;
}

.top-quarter {
  top: 0;
  height: 25%;
  width: 100%;
  position: absolute;
}

.bottom-quarter {
  bottom: 0;
  height: 25%;
  width: 100%;
  position: absolute;
}

.center-horizontal {
  top: 0;
  left: 25%;
  width: 50%;
  height: 100%;
  position: absolute;
}

.center-vertical {
  top: 25%;
  height: 50%;
  width: 100%;
  position: absolute;
}
</style>
