<template>
  <div>
    <v-app-bar app color="primary" dark>
      <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>
      <v-tooltip bottom v-for="btn in barButtons" :key="btn.id">
        <template v-slot:activator="{ on }">
          <v-btn icon v-on="on" :to="btn.to">
            <v-icon>{{ btn.icon }}</v-icon>
          </v-btn>
        </template>
        <span>{{ btn.tip }}</span>
      </v-tooltip>
      <v-toolbar-title v-if="barTitle" class="d-flex align-center">
        <span>{{ barTitle.text }}</span>
        <v-chip small class="mx-4" color="secondary--lighten">
          <span style="font-size: 1.1rem">{{ barTitle.chip }}</span>
        </v-chip>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-progress-circular v-if="$hub.tasks.length > 0" class="mr-2" indeterminate :width="2">
        <v-btn icon>
          <v-icon>mdi-triangle-wave</v-icon>
        </v-btn>
      </v-progress-circular>
      <v-btn v-else icon class="mr-0">
        <v-icon>mdi-triangle-wave</v-icon>
      </v-btn>
    </v-app-bar>
    <v-navigation-drawer app v-model="drawer">
      <v-list-item @click="$router.push({ name: 'home' })" inactive class="pb-2">
        <v-list-item-avatar rounded>
          <v-img src="../assets/logo.png" />
        </v-list-item-avatar>

        <v-list-item-content>
          <v-list-item-title class="title">MangaDepot</v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-divider></v-divider>

      <v-list>
        <!--home-->
        <v-list-item :to="{ name: 'dashboard' }">
          <v-list-item-icon>
            <v-icon>mdi-home</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>{{ $t("drawer.dashboard") }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <!--library:all-->
        <v-list-item :to="{ name: 'libraries', params: { libraryID: $LIBRARY_ID_ALL } }">
          <v-list-item-icon>
            <v-icon>mdi-bookshelf</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>{{ $t("drawer.libraries") }}</v-list-item-title>
          </v-list-item-content>
          <v-list-item-action>
            <v-btn icon @click.stop.capture.prevent="addLibrary">
              <v-icon>mdi-plus</v-icon>
            </v-btn>
          </v-list-item-action>
        </v-list-item>
        <!--libraries-->
        <v-list-item
          v-for="(l, index) in $hub.libraries"
          :key="index"
          dense
          :to="{ name: 'libraries', params: { libraryID: l.id } }">
          <v-list-item-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>{{ l.name }}</v-list-item-title>
          </v-list-item-content>
          <!-- <v-list-item-action>
            <library-actions-menu :library="l"></library-actions-menu>
          </v-list-item-action> -->
        </v-list-item>
      </v-list>
    </v-navigation-drawer>
    <v-main>
      <router-view @main-enter="onMainEnter"></router-view>
    </v-main>
    <file-browser-dialog
      v-model="showFileBrowser"
      path="/"
      :dialog-title="$t('dialog.file_browser.title')"
      :confirm-text="$t('dialog.file_browser.confirm')"></file-browser-dialog>
    <library-edit-dialog v-model="showLibraryEdit" @updated="$hub.syncLibraries"></library-edit-dialog>
  </div>
</template>
<script>
import FileBrowserDialog from "@/components/FileBrowserDialog.vue";
import LibraryEditDialog from "@/components/LibraryEditDialog.vue";

export default {
  components: { FileBrowserDialog, LibraryEditDialog },
  data: function () {
    return {
      drawer: this.$vuetify.breakpoint.lgAndUp,
      barButtons: [],
      barTitle: null,
      showFileBrowser: false,
      showLibraryEdit: false,
    };
  },

  mounted: async function () {
    this.$hubon("add-library", () => {
      this.addLibrary();
    });
  },
  methods: {
    addLibrary() {
      this.showLibraryEdit = true;
    },
    onMainEnter(options) {
      switch (options.name) {
        case "book":
          this.barButtons = [
            {
              id: 1,
              icon: "mdi-arrow-left",
              to: { name: "libraries", params: { libraryID: options.value.library_id } },
              tip: this.$t("global.go_to_library"),
            },
          ];
          this.barTitle = { text: options.value.name, chip: options.value.volume };
          break;
        case "library":
          this.barButtons = [];
          this.barTitle = { text: options.value.name, chip: options.value.count };
          break;
        case "welcome":
        case "dashboard":
          this.barButtons = [];
          this.barTitle = null;
          break;
        default:
          throw `unknown main-enter ${options.name}`;
      }
    },
  },
};
</script>
