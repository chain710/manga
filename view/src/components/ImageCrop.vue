<template>
  <div class="ma-0 pa-0 main">
    <dragable-resizable
      :w="renderWidth"
      :h="renderHeight"
      :resizable="false"
      :z="1"
      drag-handle=".drag-image"
      class-name="vdr-borderless">
      <img class="drag-image" @load="onLoad" :src="image" :style="imageStyle" />
      <dragable-resizable
        :w="setCropWidth"
        :h="setCropHeight"
        :max-width="renderWidth"
        :max-height="renderHeight"
        lock-aspect-ratio
        :z="2"
        parent
        drag-handler=".crop-area"
        class="dragable-crop-area"
        @dblclick="toggleCropArea"
        v-slot:default="props">
        <div class="crop-area d-flex justify-end align-end">
          <v-btn-toggle dense>
            <v-tooltip bottom>
              <template v-slot:activator="{ on, attrs }">
                <v-btn @click="crop('book', props.rect)" v-bind="attrs" v-on="on">
                  <v-icon>mdi-book-open</v-icon>
                </v-btn>
              </template>
              <span>{{ $t("read.set_book_thumb") }}</span>
            </v-tooltip>
            <v-tooltip bottom>
              <template v-slot:activator="{ on, attrs }">
                <v-btn @click="crop('volume', props.rect)" v-bind="attrs" v-on="on">
                  <v-icon>mdi-image-album</v-icon>
                </v-btn>
              </template>
              <span>{{ $t("read.set_volume_thumb") }}</span>
            </v-tooltip>
            <v-btn @click="close">
              <v-icon>mdi-close</v-icon>
            </v-btn>
          </v-btn-toggle>
        </div>
      </dragable-resizable>
    </dragable-resizable>
  </div>
</template>
<script>
import DragableResizable from "./DragableResizable.vue";
const defaultCropWidth = 210;
const defaultCropHeight = 297;
export default {
  data() {
    return {
      show: false,
      windowWidth: window.innerWidth,
      windowHeight: window.innerHeight,
      naturalWidth: null,
      naturalHeight: null,
      renderWidth: null,
      renderHeight: null,
      setCropWidth: defaultCropWidth,
      setCropHeight: defaultCropHeight,
      oldCrop: null,
    };
  },
  props: {
    image: { type: String, required: true },
  },
  methods: {
    onLoad(event) {
      this.naturalWidth = event.target.naturalWidth;
      this.naturalHeight = event.target.naturalHeight;
      this.resizeImage();
      this.show = true;
    },
    onResize() {
      this.windowWidth = window.innerWidth;
      this.windowHeight = window.innerHeight;
      if (this.naturalWidth) {
        this.resizeImage();
      }
    },
    close() {
      this.$emit("close");
    },
    crop(target, rect) {
      const event = {
        left: rect.left / this.renderWidth,
        top: rect.top / this.renderHeight,
        width: rect.width / this.renderWidth,
        height: rect.height / this.renderHeight,
      };
      this.$emit("crop", Object.assign({ target }, event));
    },
    resizeImage() {
      if (this.shouldFit()) {
        this.imageFitScreen();
      } else {
        this.imageOverflow();
      }
    },
    shouldFit() {
      // small than screen
      return this.naturalWidth <= this.windowWidth && this.naturalHeight <= this.windowHeight;
    },
    imageOverflow() {
      const wa = this.windowWidth / this.windowHeight;
      const ia = this.naturalWidth / this.naturalHeight;
      if (wa == ia) {
        this.renderWidth = this.windowWidth;
        this.renderHeight = this.windowHeight;
      } else if (wa < ia) {
        this.renderHeight = this.windowHeight;
        this.renderWidth = this.renderHeight * ia;
      } else {
        this.renderWidth = this.windowWidth;
        this.renderHeight = this.renderWidth / ia;
      }
    },
    imageFitScreen() {
      const f = this.fitInto(this.naturalWidth, this.naturalHeight, this.windowWidth, this.windowHeight);
      this.renderWidth = f.width;
      this.renderHeight = f.height;
      // const wa = this.windowWidth / this.windowHeight;
      // const ia = this.naturalWidth / this.naturalHeight;
      // if (wa == ia) {
      //   this.renderWidth = this.windowWidth;
      //   this.renderHeight = this.windowHeight;
      // } else if (wa < ia) {
      //   this.renderWidth = this.windowWidth;
      //   this.renderHeight = Math.floor(this.renderWidth / ia);
      // } else {
      //   this.renderHeight = this.windowHeight;
      //   this.renderWidth = Math.floor(this.renderHeight * ia);
      // }
    },
    toggleCropArea() {
      let f = {};
      if (this.oldCrop == null) {
        this.oldCrop = { width: this.setCropWidth, height: this.setCropHeight };
        f = this.fitInto(this.setCropWidth, this.setCropHeight, this.renderWidth, this.renderHeight);
      } else {
        f = this.oldCrop;
        this.oldCrop = null;
      }

      this.setCropWidth = f.width;
      this.setCropHeight = f.height;
    },
    // fit w0 h0 into w1 h1
    fitInto(w0, h0, w1, h1) {
      const a0 = w0 / h0;
      const a1 = w1 / h1;
      if (a0 == a1) {
        return { width: w1, height: h1 };
      } else if (a1 < a0) {
        return { width: w1, height: Math.floor(w1 / a0) };
      } else {
        return { height: h1, width: Math.floor(h1 * a0) };
      }
    },
    // onImageDrag(x, y) {
    //   const dx = this.renderWidth - this.windowWidth;
    //   const dy = this.renderHeight - this.windowHeight;
    //   let left = x;
    //   let top = y;
    //   if (dx > 0) {
    //     x < -dx && (left = -dx);
    //     x > 0 && (left = 0);
    //   } else if (x < 0 || x > dx) {
    //     x < 0 && (left = 0);
    //     x > dx && (left = dx);
    //   }

    //   if (dy > 0) {
    //     y < -dy && (top = -dy);
    //     y > 0 && (top = 0);
    //   } else if (y < 0 || y > dy) {
    //     y < 0 && (top = 0);
    //     y > dy && (top = dy);
    //   }

    //   console.log(`rd(${this.renderWidth}, ${this.renderHeight}) wd(${this.windowWidth}, ${this.windowHeight})`);
    //   console.log(`input(${x}, ${y}) output(${left}, ${top})`);
    //   return { left, top };
    // },
  },
  computed: {
    imageStyle() {
      return {
        width: `${this.renderWidth}px`,
        height: `${this.renderHeight}px`,
      };
    },
    containerStyle() {
      return {
        width: `${this.renderWidth}px`,
        height: `${this.renderHeight}px`,
      };
    },
    minCropHeight() {
      return Math.min(defaultCropHeight, this.renderHeight);
    },
    minCropWidth() {
      return Math.min(defaultCropWidth, this.renderWidth);
    },
  },
  components: { DragableResizable },
};
</script>
<style scoped>
.main {
  user-select: none;
}

img {
  object-fit: contain;
}

.dragable-crop-area {
  top: 0px;
  left: 0px;
}

.crop-area {
  width: 100%;
  height: 100%;
}

.vdr-borderless {
  touch-action: none;
  position: absolute;
  box-sizing: border-box;
  border: 0px;
}
</style>
