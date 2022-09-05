<template>
  <v-list dense>
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
  props: {
    item: { type: Object, required: true },
    onUpdate: { type: Function },
  },
  methods: {
    async markUnread() {
      try {
        console.log(`mark ${this.item.id} unread`);
        await this.$hub.addTask(this.$service.markVolumesUnread([this.item.id]));
        this.onUpdate && this.onUpdate();
      } catch (error) {
        console.error(`mark ${this.item.id} unread error`, error);
      }
    },
    async markRead() {
      try {
        console.log(`mark ${this.item.id} read`);
        await this.$hub.addTask(this.$service.markVolumesRead([this.item.id]));
        this.onUpdate && this.onUpdate();
      } catch (error) {
        console.error(`mark ${this.item.id} read error`, error);
      }
    },
  },
};
</script>
