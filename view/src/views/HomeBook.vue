<template>
  <div>
    <v-container fluid class="pa-6" v-if="book">
      <v-row>
        <v-col cols="4" sm="4" md="auto" lg="auto" xl="auto">
          <item-card :item="bookItem" :width="200" onlyThumb></item-card>
        </v-col>
        <v-col cols="8">
          <!--book title-->
          <v-row>
            <v-col>
              <span class="text-h4">{{ book.name }}</span>
            </v-col>
          </v-row>
          <!--writer-->
          <v-row v-if="book.writer">
            <v-col>
              <span class="text-caption">
                {{ $t("global.writer", { writer: book.writer }) }}
              </span>
            </v-col>
          </v-row>
          <!--volumes count-->
          <v-row>
            <v-col>
              <span class="text-caption">
                {{ $t("global.total_volumes", { count: book.volume }) }}
              </span>
            </v-col>
          </v-row>
          <!--summary $vuetify.breakpoint.xsOnly-->
          <v-row>
            <v-col>
              <span class="text-caption">
                {{ book.summary }}
                <!--to do read more <read-more>{{ book.metadata.summary }}</read-more>-->
              </span>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
      <v-divider class="mt-4 mb-1"></v-divider>
      <!--volumes-->
      <item-browser v-if="volumeItems.length > 0" :items="volumeItems" wrap></item-browser>
      <!--extras-->
      <div class="ma-0 pa-0" v-if="extraItems.length > 0">
        <div class="text-h5 mt-6">{{ $t("global.extras") }}</div>
        <v-divider class="mt-4 mb-1"></v-divider>
        <item-browser :items="extraItems" wrap></item-browser>
      </div>
      <!--aux speed dial TODO impl fab-->
      <v-speed-dial
        v-if="false"
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
        <v-btn fab dark small color="green" @click="editBook">
          <v-tooltip left nudge-left="5" open-delay="500">
            <template v-slot:activator="{ on, attrs }">
              <v-icon v-bind="attrs" v-on="on">mdi-pencil</v-icon>
            </template>
            <span>{{ $t("book.fab.edit") }}</span>
          </v-tooltip>
        </v-btn>
      </v-speed-dial>
    </v-container>
  </div>
</template>
<script>
import ItemBrowser from "@/components/ItemBrowser.vue";
import ItemCard from "@/components/ItemCard.vue";
export default {
  components: { ItemBrowser, ItemCard },
  data() {
    return {
      volumesCount: 1,
      book: null,
      fab: false,
    };
  },
  props: {
    bookID: { required: true },
  },
  async mounted() {
    await this.onMount();
  },
  methods: {
    async syncBook(bookID) {
      console.log("sync book", bookID);
      try {
        let resp = await this.$service.getBook(bookID);
        this.book = resp.data.data;
      } catch (error) {
        console.log("get book error", error);
        this.$nerror(`get_book`, error);
      }
    },
    editBook() {
      console.log("TODO edit book");
    },
    async onMount() {
      await this.syncBook(this.bookID);
      if (this.book) {
        this.$emit("main-enter", { name: "book", value: this.book });
      }
    },
  },
  computed: {
    volumeItems() {
      if (!this.book || !this.book.volumes) {
        return [];
      }

      return this.$convertVolumes(this.book.volumes);
    },
    extraItems() {
      if (!this.book || !this.book.extras) {
        return [];
      }

      return this.$convertVolumes(this.book.extras);
    },
    bookItem() {
      return this.$convertBook(this.book);
    },
    libraryClass() {
      if (this.$vuetify.breakpoint.smAndUp) {
        return "mx-2";
      } else if (this.$vuetify.breakpoint.xsOnly) {
        return "d-display";
      }
      return "";
    },
  },
  async beforeRouteUpdate(to, from, next) {
    if (to.params.bookID != from.params.bookID) {
      // TODO: maybe this.$hub.addTask?
      await this.syncBook(to.params.bookID);
    }
    next();
  },
};
</script>
<style scoped></style>
