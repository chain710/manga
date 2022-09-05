<template>
  <v-container fluid>
    <v-pagination
      v-if="books.length > 0"
      v-model="desiredPage"
      :length="pageCount"
      @input="jumpPage"
      total-visible="8"></v-pagination>
    <div ref="test1">
      <item-browser :width="150" v-if="books.length > 0" :items="items" wrap>
        <!-- <template v-slot:item-card="{item}"><book-card-menu :item="item"></book-card-menu></template> -->
      </item-browser>
    </div>
    <!--empty-->
    <v-card elevation="0" class="mt-6" v-if="isSetup && books.length == 0">
      <v-card-title class="d-flex justify-center align-center">
        <h1>
          <v-icon x-large color="warning">mdi-alert-decagram-outline</v-icon>
          {{ $t("library.empty") }}
        </h1>
      </v-card-title>
    </v-card>
    <!--aux speed dial-->
    <v-speed-dial
      v-if="library != null"
      v-model="fab"
      bottom
      right
      absolute
      fixed
      direction="top"
      transition="slide-y-reverse-transition">
      <template v-slot:activator>
        <v-btn v-model="fab" color="blue darken-2" dark fab>
          <v-icon v-if="fab">mdi-close</v-icon>
          <v-icon v-else>mdi-cog-outline</v-icon>
        </v-btn>
      </template>
      <v-btn v-for="(item, i) in fabItems" fab dark small :color="item.color" @click="item.onClick" :key="i">
        <v-tooltip left nudge-left="5" open-delay="500">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on">{{ item.icon }}</v-icon>
          </template>
          <span>{{ item.tip }}</span>
        </v-tooltip>
      </v-btn>
    </v-speed-dial>
    <!--confirm-->
    <confirm-dialog
      v-model="confirm.enabled"
      :title="confirm.title"
      :body="confirm.body"
      type="error"
      :confirm-func="confirm.do"></confirm-dialog>
    <!--lib edit-->
    <library-edit-dialog
      v-model="showLibraryEdit"
      :library="library"
      @updated="$hub.syncLibraries"></library-edit-dialog>
  </v-container>
</template>
<script>
import _ from "lodash";
import ItemBrowser from "@/components/ItemBrowser.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import LibraryEditDialog from "@/components/LibraryEditDialog.vue";

export default {
  components: { ItemBrowser, ConfirmDialog, LibraryEditDialog },
  data: function () {
    return {
      dataPage: 1, // used in request
      desiredPage: 1, // model
      totalElements: 0,
      books: [],
      pageSize: 20,
      fab: false,
      library: null,
      fabItems: [
        { icon: "mdi-delete", onClick: this.confirmDeleteLibrary, tip: this.$t("library.fab.delete"), color: "red" },
        { icon: "mdi-pencil", onClick: this.editLibrary, tip: this.$t("library.fab.edit"), color: "green" },
        { icon: "mdi-magnify-scan", onClick: this.scanLibrary, tip: this.$t("library.fab.scan"), color: "indigo" },
      ],
      confirm: { enabled: false, title: "", body: "" },
      showLibraryEdit: false,
      isSetup: false,
    };
  },
  props: {
    libraryID: {},
  },
  async mounted() {
    // update pagesize by window width; but dont dynamic adjust
    this.pageSize = this.predictPageSize();
    await this.setup(this.libraryID, this.$route.query.page);
    this.isSetup = true;
  },
  methods: {
    predictPageSize() {
      const windowWidth = window.innerWidth;
      const drawerWidth = 256;
      const showDrawer = this.$vuetify.breakpoint.lgAndUp;
      let innerWidth = windowWidth;
      if (showDrawer) {
        innerWidth = windowWidth - drawerWidth;
      }

      innerWidth = innerWidth - 12 * 2; // padding 12
      const cardWidth = 150 + 16; // margin = 16
      const cardPerRow = Math.floor(innerWidth / cardWidth);
      return cardPerRow * 3;
    },
    jumpPage(page) {
      if (page == this.dataPage) {
        return;
      }
      let q = _.clone(this.$route.query);
      q["page"] = page;
      this.$router.replace({ name: "libraries", params: this.$route.params, query: q });
    },
    async syncLibBooks(lib) {
      let options = {};
      if (_.parseInt(lib)) {
        options["lib"] = lib;
      }
      options["filter"] = "with_progress_relax";
      options["offset"] = this.pageSize * (this.desiredPage - 1);
      options["limit"] = this.pageSize;
      try {
        let resp = await this.$service.listBooks(options);
        this.books = resp.data.data.books;
        this.totalElements = resp.data.data.count;
        this.dataPage = this.desiredPage;
      } catch (error) {
        this.$nerror("list_book", error);
        console.error(`list book error: ${error}`);
      }
    },
    async syncLibrary(lib) {
      if (!_.parseInt(lib)) {
        this.library = null;
        return;
      }
      try {
        let resp = await this.$service.getLibrary(lib);
        this.library = resp.data.data;
      } catch (error) {
        this.$nerror("list_library", error);
        console.error(`get library error: ${error}`);
      }
    },
    editLibrary() {
      this.showLibraryEdit = true;
    },
    async scanLibrary() {
      try {
        let resp = await this.$service.scanLibrary(this.libraryID);
        console.log("scan library resp", resp);
        this.$ninfo("scan_library");
      } catch (error) {
        console.log("scan library error: ", error);
        this.$nerror("scan_library", error);
      }
    },
    confirmDeleteLibrary() {
      this.confirm.title = this.$t("dialog.delete_library.title", { name: this.library.name });
      this.confirm.body = this.$t("dialog.delete_library.body", { path: this.library.path });
      this.confirm.do = this.deleteLibrary;
      this.confirm.enabled = true;
    },
    async deleteLibrary() {
      try {
        let resp = await this.$service.deleteLibrary(this.library.id);
        console.log(`delete library ${this.library.id}`, resp);
        this.$ninfo("delete_library");
      } catch (error) {
        this.$nerror("delete_library", error);
        console.error(`delete library error ${this.library.id}`, error);
      }

      try {
        await this.$hub.syncLibraries();
        this.$router.replace({ name: "libraries", params: { libraryID: this.$LIBRARY_ID_ALL } }).catch((err) => {
          // may navigate to welcome, which is expected
          console.debug(`replace lib error`, err);
        });
      } catch (error) {
        console.error("navigate to libraries error", error);
      }
    },
    async setup(lib, page) {
      const queryPage = _.parseInt(page);
      this.desiredPage = queryPage ? queryPage : 1;
      const p = Promise.all([this.syncLibBooks(lib), this.syncLibrary(lib)]);
      await this.$hub.addTask(p);
      this.$emit("main-enter", {
        name: "library",
        value: {
          id: this.libraryID,
          name: this.library ? this.library.name : this.$t("library.all"),
          count: this.totalElements,
        },
      });
    },
  },
  computed: {
    items() {
      let results = [];
      for (let book of this.books) {
        results.push(this.$convertBook(book));
      }

      return results;
    },
    pageCount() {
      return Math.ceil(this.totalElements / this.pageSize);
    },
  },
  async beforeRouteUpdate(to, from, next) {
    if (to.params.libraryID != from.params.libraryID || to.query.page != from.query.page) {
      await this.setup(to.params.libraryID, to.query.page);
    }
    next();
  },
};
</script>
<style scoped></style>
