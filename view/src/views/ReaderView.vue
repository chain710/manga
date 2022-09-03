<template>
  <v-container
    v-resize="onResize"
    class="ma-0 pa-0 full-height touch-less"
    :class="backgroundColor"
    fluid
    v-if="volume">
    <v-overlay :value="thumbSelection.inProgressing">
      <v-progress-circular indeterminate size="64"></v-progress-circular>
    </v-overlay>
    <image-crop
      v-if="thumbSelection.enabled"
      :image="thumbSelection.imageURL"
      @crop="onCropUpdate"
      @close="onCropClose"></image-crop>
    <div v-else>
      <!--top tool bar-->
      <v-slide-y-transition>
        <v-toolbar
          dense
          elevation="1"
          v-if="showToolbars"
          class="settings full-width tool-bar-glass"
          style="position: fixed; top: 0">
          <v-btn icon :to="{ name: 'book', params: { bookID: volume.book_id } }">
            <v-icon>mdi-arrow-left</v-icon>
          </v-btn>
          <v-toolbar-title>{{ volume.book_name }} / {{ volume.title }}</v-toolbar-title>
          <v-spacer></v-spacer>
          <v-btn icon @click="rotate(-90)">
            <v-icon>mdi-file-rotate-left</v-icon>
          </v-btn>
          <v-btn icon @click="rotate(90)">
            <v-icon>mdi-file-rotate-right</v-icon>
          </v-btn>
          <v-btn icon @click="enableThumbSelection">
            <v-icon>mdi-format-wrap-square</v-icon>
          </v-btn>
          <v-btn @click="showThumbExplorer = true" icon>
            <v-icon>mdi-view-grid</v-icon>
          </v-btn>
          <v-btn v-if="fullScreenEnabled" icon @click="toggleFullScreen">
            <v-icon>{{ fullScreenIcon }}</v-icon>
          </v-btn>
          <!-- <v-btn icon @click="onHelp">
            <v-icon>mdi-help-circle</v-icon>
          </v-btn> -->
          <v-btn icon @click="showSettings = true">
            <v-icon>mdi-cog</v-icon>
          </v-btn>
        </v-toolbar>
      </v-slide-y-transition>
      <!-- reader -->
      <div class="full-height">
        <page-reader
          @center-click="showToolbars = !showToolbars"
          :spreads="spreads"
          :page.sync="pageNumber"
          :read-mode="readMode"
          :image-rotate.sync="rotateImage"
          @jump-next-stop="onJumpNextStop"
          @jump-prev-stop="onJumpPrevStop"></page-reader>
      </div>
      <!--bottom tool bar: slider-->
      <v-slide-y-reverse-transition>
        <v-toolbar
          dense
          elevation="1"
          class="settings full-width tool-bar-glass"
          style="position: fixed; bottom: 0"
          v-if="showToolbars">
          <v-row justify="center">
            <v-col class="px-0">
              <v-progress-linear
                :active="loadingVolume"
                indeterminate
                height="6"
                color="deep-orange"></v-progress-linear>
              <v-slider
                v-model="goToPage"
                hide-details
                thumb-label
                class="align-center px-4"
                min="1"
                :max="volume.page_count">
                <template v-slot:prepend>
                  <v-tooltip top open-delay="200">
                    <template v-slot:activator="{ on }">
                      <v-icon
                        :disabled="!volume.prev_volume_id || loadingVolume"
                        color="secondary"
                        @click="prevVolume"
                        v-on="on">
                        mdi-skip-previous-circle
                      </v-icon>
                    </template>
                    <span>{{ $t("read.prev_volume") }}</span>
                  </v-tooltip>
                </template>
                <template v-slot:append>
                  <div class="d-flex align-center">
                    <v-label>{{ pageNumber }}</v-label>
                    <v-label class="px-1">/</v-label>
                    <v-label>{{ volume.page_count }}</v-label>
                    <v-tooltip top open-delay="200">
                      <template v-slot:activator="{ on }">
                        <v-icon
                          :disabled="!volume.next_volume_id || loadingVolume"
                          class="ml-2"
                          color="secondary"
                          @click="nextVolume"
                          v-on="on">
                          mdi-skip-next-circle
                        </v-icon>
                      </template>
                      <span>{{ $t("read.next_volume") }}</span>
                    </v-tooltip>
                  </div>
                </template>
              </v-slider>
            </v-col>
          </v-row>
        </v-toolbar>
      </v-slide-y-reverse-transition>
      <!--thumb explorer dialog-->
      <thumb-explorer-dialog
        v-model="showThumbExplorer"
        @go="goTo"
        :page="pageNumber"
        :thumbs="thumbs"></thumb-explorer-dialog>
      <!--reader settings-->
      <v-bottom-sheet
        v-model="showSettings"
        :close-on-content-click="false"
        max-width="500"
        @keydown.esc.stop=""
        scrollable>
        <v-card>
          <v-toolbar dark color="primary">
            <v-btn icon dark @click="showSettings = false">
              <v-icon>mdi-close</v-icon>
            </v-btn>
            <v-toolbar-title>{{ $t("read.settings.title") }}</v-toolbar-title>
          </v-toolbar>

          <v-card-text class="pa-0">
            <v-list class="full-height full-width">
              <v-subheader class="font-weight-black text-h6">
                {{ $t("read.settings.display") }}
              </v-subheader>
              <v-list-item>
                <settings-select :items="readModes" v-model="readMode" :label="$t('read.settings.reading_mode')" />
              </v-list-item>
              <v-list-item>
                <settings-select
                  :items="backgroundColors"
                  v-model="backgroundColor"
                  :label="$t('read.settings.background_color')" />
              </v-list-item>
              <v-list-item>
                <settings-switch
                  v-model="alwaysFullScreen"
                  :label="$t('read.settings.always_fullscreen')"
                  :disabled="!fullScreenEnabled" />
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-bottom-sheet>
      <!--next/prev book snack bar-->
      <v-snackbar
        v-model="jumpToPrevBook"
        :timeout="jumpConfirmationDelay"
        top
        color="rgba(0, 0, 0, 0.8)"
        multi-line
        class="mt-12">
        <div class="body-1 pa-6">
          <p>{{ $t("read.beginning_of_book") }}</p>
          <p v-if="volume.prev_volume_id">{{ $t("read.move_previous") }}</p>
        </div>
      </v-snackbar>
      <v-snackbar
        v-model="jumpToNextBook"
        :timeout="jumpConfirmationDelay"
        top
        color="rgba(0, 0, 0, 0.8)"
        multi-line
        class="mt-12">
        <div class="text-body-1 pa-6">
          <p>{{ $t("read.end_of_book") }}</p>
          <p v-if="volume.next_volume_id">{{ $t("read.move_next") }}</p>
          <p v-else>{{ $t("read.move_next_exit") }}</p>
        </div>
      </v-snackbar>
    </div>
  </v-container>
</template>
<script>
import PageReader from "@/components/PageReader.vue";
import ThumbExplorerDialog from "@/components/ThumbExplorerDialog.vue";
import SettingsSelect from "@/components/SettingsSelect.vue";
import SettingsSwitch from "@/components/SettingsSwitch.vue";
import ImageCrop from "@/components/ImageCrop.vue";
import screenfull from "screenfull";
import _ from "lodash";
export default {
  components: {
    PageReader,
    ThumbExplorerDialog,
    SettingsSelect,
    SettingsSwitch,
    ImageCrop,
  },
  data: function () {
    return {
      showToolbars: true,
      pageNumber: 1,
      volume: null,
      goToPage: 1,
      loadingVolume: true,
      showThumbExplorer: false,
      showSettings: false,
      rotateImage: 0,
      backgroundColors: [
        { text: this.$t("read.settings.background_color_white").toString(), value: "bg-white" },
        {
          text: this.$t("read.settings.background_color_gray").toString(),
          value: "bg-gray",
        },
        { text: this.$t("read.settings.background_color_black").toString(), value: "bg-black" },
      ],
      readModes: [
        { text: this.$t("read.settings.read_mode_ltr").toString(), value: "ltr" },
        { text: this.$t("read.settings.read_mode_rtl").toString(), value: "rtl" },
      ],
      jumpToNextBook: false,
      jumpToPrevBook: false,
      jumpConfirmationDelay: 3000,
      thumbSelection: {
        enabled: false,
        imageURL: null,
        // crop positions
        crop: { top: 0, left: 0, width: 0, height: 0 },
        inProgressing: false,
      },
      windowWidth: window.innerWidth,
      windowHeight: window.innerHeight,
      fullScreen: false,
      fullScreenEnabled: false,
    };
  },
  props: {
    volumeID: { required: true },
  },
  created() {
    this.fullScreenEnabled = screenfull.isEnabled;
    this.fullScreen = screenfull.isFullscreen;
    if (screenfull.isEnabled) {
      screenfull.on("change", this.fullScreenChanged);
    }
  },
  async mounted() {
    await this.setup(this.volumeID, _.parseInt(this.$route.query.page));
    if (this.alwaysFullScreen) {
      try {
        this.enterFullScreen();
      } catch (error) {
        console.log(`enterfull screen error, maybe cause by refresh page`, error);
      }
    }
  },
  destroyed() {
    if (screenfull.isEnabled) {
      screenfull.off("change", this.fullscreenChanged);
      screenfull.exit();
    }
  },
  methods: {
    async syncVolume(volumeID) {
      try {
        console.log(this.volume);
      } catch (error) {
        console.error(`get volume ${volumeID} error ${error}`);
      }
    },
    async setup(volumeID, pageNo) {
      console.log(`setup vol ${volumeID} page ${pageNo}`);
      this.loadingVolume = true;
      try {
        let resp = await this.$service.getVolume(volumeID);
        this.volume = resp.data.data;
      } catch (error) {
        console.error(`fetch volume data error: ${error}`);
      } finally {
        this.loadingVolume = false;
      }

      if (pageNo) {
        this.pageNumber = _.parseInt(pageNo);
      } else {
        this.pageNumber = 1;
      }
    },
    onHelp() {
      console.log("on help");
    },
    prevVolume() {
      this.jumpToVolume(this.volume.prev_volume_id, 1);
    },
    nextVolume() {
      this.jumpToVolume(this.volume.next_volume_id, 1);
    },
    jumpToVolume(volumeID, page) {
      let q = _.clone(this.$route.query);
      q["page"] = page.toString();
      let link = {
        name: this.$route.name,
        params: { volumeID: volumeID },
        query: q,
      };
      return this.$router.push(link);
    },
    goTo(page) {
      this.pageNumber = page;
      this.showThumbExplorer = false;
    },
    rotate(deg) {
      this.rotateImage = this.rotateImage + deg;
    },
    onJumpNextStop() {
      if (this.jumpToNextBook) {
        if (this.volume.next_volume_id) {
          this.jumpToVolume(this.volume.next_volume_id, 1);
        } else {
          this.$router.push({ name: "book", params: { bookID: this.volume.book_id } });
        }
      } else {
        this.jumpToNextBook = true;
      }
    },
    onJumpPrevStop() {
      if (this.jumpToPrevBook) {
        if (this.volume.prev_volume_id) {
          this.jumpToVolume(this.volume.prev_volume_id, 1);
        }
      } else {
        this.jumpToPrevBook = true;
      }
    },
    markProgress: _.debounce(function (page) {
      this.$service.updateVolumeProgress(this.volumeID, page).catch((err) => {
        console.error(`mark volume ${this.volumeID} progress ${page} error: ${err}`);
      });
    }, 100),
    async setVolumeThumb() {
      this.thumbSelection.inProgressing = true;
      const rect = this.getSelectRect();
      try {
        let resp = await this.$service.cropImage(this.volume.id, this.pageNumber, rect);
        await this.$service.setVolumeThumb(this.volume.id, resp.data);
        this.$ninfo("set_volume_thumb");
      } catch (error) {
        console.error(`set volume ${this.volume.id} thumb error ${error}`);
        this.$nerror("set_volume_thumb", { err: error });
      } finally {
        this.thumbSelection.enabled = false;
        this.thumbSelection.inProgressing = false;
      }
    },
    async setBookThumb() {
      this.thumbSelection.inProgressing = true;
      const rect = this.getSelectRect();
      try {
        let resp = await this.$service.cropImage(this.volume.id, this.pageNumber, rect);
        await this.$service.setBookThumb(this.volume.book_id, resp.data);
        this.$ninfo("set_book_thumb");
      } catch (error) {
        console.error(`set book ${this.volume.book_id} thumb error ${error}`);
        this.$nerror("set_book_thumb", { err: error });
      } finally {
        this.thumbSelection.enabled = false;
        this.thumbSelection.inProgressing = false;
      }
    },
    async onCropUpdate(event) {
      this.thumbSelection.inProgressing = true;
      try {
        const rect = { left: event.left, top: event.top, width: event.width, height: event.height };
        let resp = await this.$service.cropImage(this.volume.id, this.pageNumber, rect);
        console.log(resp);
        switch (event.target) {
          case "book":
            await this.$service.setBookThumb(this.volume.book_id, resp.data);
            this.$ninfo("set_book_thumb");
            break;
          case "volume":
            await this.$service.setVolumeThumb(this.volume.id, resp.data);
            this.$ninfo("set_volume_thumb");
            break;
          default:
            throw `wrong target ${event.target}`;
        }
      } catch (error) {
        console.error(`set book/volume ${this.volume.book_id} thumb error ${error}`);
        this.$nerror("set_thumb", { err: error });
      } finally {
        this.thumbSelection.enabled = false;
        this.thumbSelection.inProgressing = false;
      }
    },
    onCropClose() {
      this.thumbSelection.enabled = false;
    },
    onResize() {
      this.windowWidth = window.innerWidth;
      this.windowHeight = window.innerHeight;
    },
    enableThumbSelection() {
      // TODO: page reader spreads=>pages
      const pageURL = this.$service.pageURL(this.volume.id, this.pageNumber);
      this.thumbSelection.imageURL = pageURL;
      this.thumbSelection.enabled = true;
      this.thumbSelection.inProgressing = false;
    },
    enterFullScreen() {
      if (!screenfull.isEnabled) {
        return;
      }
      return screenfull.request(document.documentElement, { navigationUI: "hide" });
    },
    toggleFullScreen() {
      if (!screenfull.isEnabled) {
        return;
      }
      if (screenfull.isFullscreen) {
        screenfull.exit();
      } else {
        screenfull.request(document.documentElement, { navigationUI: "hide" });
      }
    },

    fullScreenChanged() {
      this.fullScreen = screenfull.isEnabled && screenfull.isFullscreen;
    },
  },
  computed: {
    fullScreenIcon() {
      return this.fullScreen ? "mdi-fullscreen-exit" : "mdi-fullscreen";
    },
    spreads() {
      let spreads = [];
      for (const [index, file] of this.volume.files.entries()) {
        spreads.push([
          {
            name: file.path,
            size: file.size,
            url: this.$service.pageURL(this.volume.id, index + 1),
            page: index + 1,
            thumb: this.$service.pageThumbURL(this.volume.id, index + 1),
          },
        ]);
      }
      return spreads;
    },
    thumbs() {
      let thumbs = [];
      for (const [index] of this.volume.files.entries()) {
        thumbs.push({
          url: this.$service.pageThumbURL(this.volume.id, index + 1),
          page: index + 1,
        });
      }
      return thumbs;
    },

    backgroundColor: {
      get: function () {
        return this.$settings.backgroundColor;
      },
      set: function (color) {
        if (this.backgroundColors.map((x) => x.value).includes(color)) {
          this.$settings.backgroundColor = color;
        }
      },
    },

    alwaysFullScreen: {
      get: function () {
        return this.$settings.alwaysFullScreen;
      },
      set: function (val) {
        this.$settings.alwaysFullScreen = val;
        if (screenfull.isEnabled) {
          if (val != screenfull.isFullscreen) {
            this.toggleFullScreen();
          }
        }
      },
    },

    readMode: {
      get: function () {
        return this.$settings.readMode;
      },
      set: function (val) {
        if (this.readModes.map((x) => x.value).includes(val)) {
          this.$settings.readMode = val;
        }
      },
    },
  },

  watch: {
    pageNumber(val, old) {
      this.markProgress(val);
      console.log(`reader pageNumber: ${old}=>${val}`);
      let q = _.clone(this.$route.query);
      q["page"] = val.toString();
      const replace = { name: this.$route.name, params: this.$route.params, query: q };
      // reload page cause NavigationDuplicated error, here we ignore it
      this.$router.replace(replace).catch((error) => {
        if (error.name != "NavigationDuplicated") {
          throw error;
        }
      });
      this.goToPage = val;
    },
    goToPage: _.debounce(function (newVal) {
      if (newVal != this.pageNumber) {
        this.pageNumber = newVal;
        this.rotateImage = 0;
        console.log("goToPage pageNumber", newVal);
      }
    }, 100),
  },

  beforeRouteUpdate(to, from, next) {
    if (to.params.volumeID != from.params.volumeID) {
      this.setup(to.params.volumeID, to.query.page);
    }
    next();
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

.bg-black {
  background-color: black;
}

.bg-white {
  background-color: white;
}

.tool-bar-glass {
  background: rgba(255, 255, 255, 0.7) !important;
  box-shadow: 0 4px 30px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(5px);
  -webkit-backdrop-filter: blur(5px);
}

.fit-screen {
  width: 100vw;
  height: 100vh;
}

.touch-less {
  touch-action: none;
}
</style>
