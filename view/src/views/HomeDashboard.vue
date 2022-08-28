<template>
  <v-container fluid>
    <horizontal-scroller v-if="recentReadVolumeItems.length > 0">
      <template v-slot:prepend>
        <div class="title">{{ $t("dashboard.recent_read") }}</div>
      </template>
      <template v-slot:content>
        <item-browser :items="recentReadVolumeItems"></item-browser>
      </template>
    </horizontal-scroller>

    <horizontal-scroller>
      <template v-slot:prepend>
        <div class="title">{{ $t("dashboard.latest") }}</div>
      </template>
      <template v-slot:content>
        <item-browser :items="latestUpdateBookItems"></item-browser>
      </template>
    </horizontal-scroller>
  </v-container>
</template>
<script>
import HorizontalScroller from "@/components/HorizontalScroller.vue";
import ItemBrowser from "@/components/ItemBrowser.vue";
export default {
  components: { HorizontalScroller, ItemBrowser },
  data: () => ({
    booksLimit: 20,
    latestUpdateBooks: [],
    recentReadVolumes: [],
  }),

  async mounted() {
    await this.syncBooks();
    this.$emit("main-enter", { name: "dashboard" });
  },

  methods: {
    async syncBooks() {
      await Promise.all([this.syncLatestUpdateBooks(), this.syncRecentReadVolumes()]);
    },

    async syncLatestUpdateBooks() {
      try {
        let resp = await this.$service.listBooks({
          sort: "latest",
          limit: this.booksLimit,
          filter: "with_progress_relax",
        });
        this.latestUpdateBooks = resp.data.data.books;
      } catch (error) {
        console.log("list latest books error", error);
        // TODO emit error to home view?
      }
    },

    async syncRecentReadVolumes() {
      try {
        let resp = await this.$service.listVolume({
          filter: "reading",
        });
        if (resp.data.data) {
          this.recentReadVolumes = resp.data.data;
        } else {
          this.recentReadVolumes = [];
        }
      } catch (error) {
        console.log("list recent read books error", error);
        this.$nerror(`list_volume`, error);
      }
    },
  },
  computed: {
    latestUpdateBookItems() {
      return this.$convertBooks(this.latestUpdateBooks);
    },
    recentReadVolumeItems() {
      return this.$convertVolumes(this.recentReadVolumes, true);
    },
  },
};
</script>
<style scoped></style>
