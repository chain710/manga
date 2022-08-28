<template>
  <div>
    <v-item-group :class="groupClass">
      <v-item v-for="item in items" :key="item.id" class="my-2 mx-2 item-wrap">
        <item-card :width="width" :item="item" :configurable="configurable">
          <slot name="item-card" v-bind:item="item"></slot>
        </item-card>
      </v-item>
    </v-item-group>
  </div>
</template>
<script>
import ItemCard from "@/components/ItemCard.vue";
export default {
  components: { ItemCard },
  data: () => {
    return {};
  },
  props: {
    items: { type: Array, required: true },
    wrap: { type: Boolean, default: false },
    configurable: { type: Boolean, default: false },
    width: { type: Number },
  },
  mounted() {},
  computed: {
    groupClass() {
      return this.wrap ? "d-flex flex-wrap" : "d-flex flex-nowrap";
    },
  },
};
</script>
<style scoped>
.item-wrap {
  position: relative;
}
</style>
