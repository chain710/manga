<template>
  <v-hover v-slot="{ hover }">
    <v-card :width="width" @click="onClick" :ripple="false">
      <v-img
        :src="item.thumb"
        :lazy-src="thumbnailError ? defaultCover : undefined"
        aspect-ratio="0.7071"
        contain
        @error="thumbnailError = true"
        @load="thumbnailError = false">
        <!--unread triangle-->
        <div class="unread" v-if="item.unread && !onlyThumb"></div>
        <v-fade-transition>
          <v-overlay
            v-if="showOverlay(hover)"
            absolute
            :opacity="overlayOpacity(hover)"
            class="overlay-full"
            :class="overlayClass(hover)">
            <!-- Circle icon for selection (top left) -->
            <v-icon v-if="selectable" :color="selected ? 'secondary' : ''" class="select-icon" @click.stop="selectItem">
              {{
                selected || (preSelect && hover) ? "mdi-checkbox-marked-circle" : "mdi-checkbox-blank-circle-outline"
              }}
            </v-icon>

            <!-- FAB reading (center) -->
            <v-btn
              v-if="isReadable"
              fab
              x-large
              color="accent"
              class="read-fab"
              :to="item.readable"
              @click.native="$event.stopImmediatePropagation()">
              <v-icon>mdi-book-open-page-variant</v-icon>
            </v-btn>

            <!-- Pen icon for edition (bottom left) -->
            <v-btn icon v-if="isEditable" class="edit-icon" @click.stop="editItem">
              <v-icon>mdi-pencil</v-icon>
            </v-btn>

            <!--menu(bottom right)-->
            <div class="menu-icon" v-if="isConfigurable">
              <v-menu offset-y v-model="menuState">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn icon v-bind="attrs" v-on="on">
                    <v-icon>mdi-dots-vertical</v-icon>
                  </v-btn>
                </template>
                <slot></slot>
              </v-menu>
            </div>
          </v-overlay>
        </v-fade-transition>
        <v-progress-linear
          v-if="item.readPercent && !onlyThumb"
          :value="item.readPercent"
          color="orange"
          height="6"
          style="position: absolute; bottom: 0"></v-progress-linear>
      </v-img>
      <v-card-subtitle v-if="!onlyThumb" class="pa-2 pb-1 text--primary subtitle" v-line-clamp="2">
        <router-link :to="item.title.to" class="subtitle-link text-truncate link-underline" v-if="item.title.to">
          {{ item.title.text }}
        </router-link>
        <span v-else>{{ item.title.text }}</span>
        <router-link
          :to="item.subTitle.to"
          class="subtitle-link text-truncate link-underline font-weight-light"
          v-if="item.subTitle.to">
          {{ item.subTitle.text }}
        </router-link>
        <span class="text-truncate font-weight-light" v-else>{{ item.subTitle.text }}</span>
      </v-card-subtitle>
      <v-card-text v-if="!onlyThumb" class="px-2 font-weight-light">
        <span v-if="item.bottomText" v-html="item.bottomText"></span>
      </v-card-text>
    </v-card>
  </v-hover>
</template>
<script>
import { defaultCover } from "@/image";
export default {
  data: () => {
    return {
      thumbnailError: false,
      preSelect: false, // props
      readProgressPercentage: 50, // props, -1 = noprogress
      selected: false,
      isUnread: true,
      defaultCover: defaultCover,
      menuState: false,
    };
  },

  props: {
    width: { type: Number, default: 150 },
    item: Object,
    onlyThumb: { type: Boolean, default: false },
    selectable: { type: Boolean, default: false },
    configurable: { type: Boolean, default: false },
    editable: { type: Boolean, default: false },
  },

  mounted() {},

  methods: {
    onClick() {
      if (this.onlyThumb) {
        return;
      }
      if (this.item.readable) {
        this.$router.push(this.item.readable);
      } else if (this.item.title && this.item.title.to) {
        this.$router.push(this.item.title.to);
      }
    },
    showOverlay(hover) {
      return !this.onlyThumb && (hover || this.preSelect || this.selected || this.menuState);
    },
    overlayOpacity(hover) {
      return hover ? 0.3 : 0;
    },
    overlayClass(hover) {
      return hover || this.preSelect ? "item-border-darken" : "item-border-transparent";
    },
    selectItem() {
      this.selected = !this.selected;
    },
    editItem() {
      this.$emit("edit-item", this.item);
    },
  },

  computed: {
    isReadable() {
      return this.item.readable && !this.selected && !this.preSelect;
    },

    isEditable() {
      return this.editable && !this.selected && !this.preSelect;
    },

    isConfigurable() {
      return this.configurable && !this.selected && !this.preSelect;
    }
  },
};
</script>
<style scoped>
.select-icon {
  position: absolute;
  top: 5px;
  left: 10px;
}
.read-fab {
  position: absolute;
  top: 50%;
  left: 50%;
  margin-left: -36px;
  margin-top: -36px;
}
.edit-icon {
  position: absolute;
  bottom: 5px;
  left: 5px;
}

.menu-icon {
  position: absolute;
  bottom: 5px;
  right: 5px;
}

.item-border-darken {
  border: 3px solid orange;
}

.item-border-transparent {
  border: 3px solid transparent;
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

.subtitle-link {
  display: block;
}

.v-image__image--preload {
  filter: none;
}
</style>
