<template>
  <v-list dense>
    <v-list-item @click="scan">
      <v-list-item-title>{{ $t("book.scan") }}</v-list-item-title>
    </v-list-item>
    <v-list-item @click="markUnread">
      <v-list-item-title>{{ $t("book.mark_unread") }}</v-list-item-title>
    </v-list-item>
    <v-list-item @click="markRead">
      <v-list-item-title>{{ $t("book.mark_read") }}</v-list-item-title>
    </v-list-item>
  </v-list>
</template>
<script>
export default {
  methods: {
    scan() {
      console.log(`scan item`, this.item);
    },
    markUnread() {},
    markRead() {},
  },
  props: {
    item: { type: Object, required: true },
  },
};
</script>
