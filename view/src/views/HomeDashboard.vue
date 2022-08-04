<template>
  <v-container fluid>
    <p>home-dashboardxx</p>
    <horizontal-scroller>
      <template v-slot:prepend>
        <div class="title">reading</div>
      </template>
      <template v-slot:content>
        <v-item-group class="d-flex flex-wrap">
          <v-item
            v-for="book in books"
            :key="book.id"
            class="my-2 mx-2"
            style="position: relative"
          >
            <book-card></book-card>
          </v-item>
        </v-item-group>
      </template>
    </horizontal-scroller>
  </v-container>
</template>
<script>
import HorizontalScroller from "@/components/HorizontalScroller.vue";
import BookCard from "@/components/BookCard.vue";
export default {
  components: { HorizontalScroller, BookCard },
  data: () => ({
    books: [
      { id: 1, name: "book1" },
      { id: 2, name: "book2" },
    ],
  }),
};
</script>
<style scoped>

</style>
