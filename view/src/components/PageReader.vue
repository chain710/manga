<template>
  <div>
    <v-carousel
      :show-arrows="false"
      :continuous="false"
      :reverse="flipDirection"
      :vertical="vertical"
      hide-delimiters
      touchless
      height="100%"
      v-model="carouselPage"
    >
      <!--  Carousel: pages  -->
      <v-carousel-item
        v-for="(spread, i) in spreads"
        :key="`spread${i}`"
        class="full-height"
      >
        <div class="full-height d-flex flex-column justify-center">
          <div
            :class="`d-flex flex-row${
              flipDirection ? '-reverse' : ''
            } justify-center px-0 mx-0`"
          >
            <img
              v-for="(page, j) in spread"
              :alt="`Page ${page.number}`"
              :key="`spread${i}-${j}`"
              :src="page.url"
              class="img-fit-all img-fit-screen"
            />
          </div>
        </div>
      </v-carousel-item>
    </v-carousel>
    <!--  clickable zone: left  -->
    <div
      v-if="!vertical"
      @click="turnLeft()"
      class="left-quarter"
      style="z-index: 1"
    />

    <!--  clickable zone: right  -->
    <div
      v-if="!vertical"
      @click="turnRight()"
      class="right-quarter"
      style="z-index: 1"
    />

    <!--  clickable zone: top  -->
    <div
      v-if="vertical"
      @click="prev()"
      class="top-quarter"
      style="z-index: 1"
    />

    <!--  clickable zone: bottom  -->
    <div
      v-if="vertical"
      @click="next()"
      class="bottom-quarter"
      style="z-index: 1"
    />

    <!--  clickable zone: menu  -->
    <div
      @click="centerClick()"
      :class="`${vertical ? 'center-vertical' : 'center-horizontal'}`"
      style="z-index: 1"
    />
  </div>
</template>
<script>
export default {
  data: () => {
    return {
      carouselPage: 0, // starting from 0
      flipDirection: false,
      vertical: false,
      spreads: [
        [{ url: "Jjdjr01-001.jpg" }],
        [{ url: "Jjdjr01-002.png" }],
        [{ url: "Jjdjr01-003.png" }],
      ],
    };
  },
  methods: {
    turnLeft: function () {
      this.flipDirection ? this.next() : this.prev();
    },
    turnRight: function () {
      this.flipDirection ? this.prev() : this.next();
    },
    next: function () {
      console.log(`next: ${this.carouselPage} ${this.spreads.length}`);
      if (!this.hasNext()) {
        this.$emit("jump-next-stop");
        return;
      }
      this.carouselPage = this.carouselPage + 1;
      this.$emit("jump-next");
    },
    prev: function () {
      console.log(`prev: ${this.carouselPage}`);
      if (!this.hasPrev()) {
        this.$emit("jump-prev-stop");
        return;
      }
      this.carouselPage = this.carouselPage - 1;
      this.$emit("jump-prev");
    },
    hasNext: function () {
      return this.carouselPage < this.spreads.length-1;
    },
    hasPrev: function () {
      return this.carouselPage > 0;
    },
    centerClick: function () {
      console.log("center click");
      this.$emit("center-click");
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

.img-fit-width {
  width: 100vw;
  min-height: 100vh;
  align-self: flex-start;
}

.img-fit-screen {
  width: 100vw;
  height: 100vh;
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
