<template>
  <v-hover v-slot="{ hover }">
    <v-card :width="width" @click="onClick" :ripple="false">
      <v-img
        :src="thumbnailUrl"
        :lazy-src="thumbnailError ? coverBase64 : undefined"
        aspect-ratio="0.7071"
        contain
        @error="thumbnailError = true"
        @load="thumbnailError = false">
        <div class="unread" v-if="isUnread"></div>

        <v-fade-transition>
          <v-overlay
            v-if="showOverlay(hover)"
            absolute
            :opacity="overlayOpacity(hover)"
            class="overlay-full">
            <v-icon
              :color="selected ? 'secondary' : ''"
              class="select-icon"
              @click.stop="selectItem">
              {{
                selected || (preSelect && hover)
                  ? "mdi-checkbox-marked-circle"
                  : "mdi-checkbox-blank-circle-outline"
              }}
            </v-icon>
          </v-overlay>
        </v-fade-transition>
        <v-progress-linear
          v-if="readProgressPercentage > 0"
          :value="readProgressPercentage"
          color="orange"
          height="6"
          style="position: absolute; bottom: 0"></v-progress-linear>
      </v-img>
      <v-card-subtitle class="pa-2 pb-1 text--primary subtitle" v-line-clamp="2">
        Book Name Vol X
      </v-card-subtitle>
      <v-card-text class="px-2 font-weight-light">this is body</v-card-text>
    </v-card>
  </v-hover>
</template>
<script>
export default {
  data: () => {
    return {
      width: 150,
      thumbnailError: false,
      thumbnailUrl: "", // props
      preSelect: false, // props
      readProgressPercentage: 50, // props, -1 = noprogress
      selected: false,
      isUnread: true,
    }
  },

  methods: {
    onClick: function () {
      console.log("card on click")
    },
    showOverlay: function (hover) {
      return hover || this.preSelect || this.selected
    },
    overlayOpacity: function (hover) {
      return hover ? 0.3 : 0
    },
    selectItem: function () {
      this.selected = !this.selected
      console.log(`selectedxx = ${this.selected}`)
    },
  },
}
</script>
<style scoped>
.select-icon {
  position: absolute;
  top: 5px;
  left: 10px;
}
</style>
<style>
.unread {
  border-left: 25px solid transparent;
  border-right: 25px solid orange;
  border-bottom: 25px solid transparent;
  height: 0;
  width: 0;
  position: absolute;
  right: 0;
  z-index: 2;
}

.overlay-full .v-overlay__content {
  width: 100%;
  height: 100%;
}

.subtitle {
  word-break: normal !important;
  height: 4em;
}
</style>
